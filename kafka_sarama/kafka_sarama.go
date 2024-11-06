package kafkasarama

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/suiguo/hwlib/logger"
)

type Algorithm string
type RequiredAcks sarama.RequiredAcks

const (
	NoResponse   RequiredAcks = 0
	WaitForLocal RequiredAcks = 1
	WaitForAll   RequiredAcks = -1
)

const (
	SHA_256 Algorithm = "sha256"
	SHA_512 Algorithm = "sha512"
)

const KafkaSaramaTag = "KafkaSarama"

type Producer interface {
	PushMsg(topic string, msg []byte) error
	Close() error
}

// 同步生产者
type syncproducer struct {
	sarama.SyncProducer
	log logger.Logger
}

func (a *syncproducer) PushMsg(topic string, msg []byte) error {
	if a.SyncProducer == nil {
		return fmt.Errorf("SyncProducer is nil")
	}
	productMsg := &sarama.ProducerMessage{Topic: topic, Value: sarama.ByteEncoder(msg)}
	_, _, err := a.SendMessage(productMsg)
	if err != nil {
		if a.log != nil {
			a.log.Error(KafkaSaramaTag, "PushMsg", err)
		}
		return err
	}
	return nil
}
func (a *syncproducer) Close() error {
	if a.SyncProducer != nil {
		return a.SyncProducer.Close()
	}
	return nil
}

type asyncproducer struct {
	sarama.AsyncProducer
	log logger.Logger
}

func (a *asyncproducer) PushMsg(topic string, msg []byte) error {
	if a.AsyncProducer == nil {
		return fmt.Errorf("SyncProducer is nil")
	}
	productMsg := &sarama.ProducerMessage{Topic: topic, Value: sarama.ByteEncoder(msg)}
	select {
	case a.AsyncProducer.Input() <- productMsg:
		return nil
	case err := <-a.AsyncProducer.Errors():
		return err
	}
}
func (a *asyncproducer) Close() error {
	if a.AsyncProducer != nil {
		return a.AsyncProducer.Close()
	}
	return nil
}

// type ProductConfig func(*sarama.Config)

// ack
func WithProductAcks(ack RequiredAcks) Config {
	return func(c *sarama.Config) {
		c.Producer.RequiredAcks = sarama.RequiredAcks(ack)
	}
}

// 超时
func WithProductTimeOut(t time.Duration) Config {
	return func(c *sarama.Config) {
		c.Producer.Timeout = t
	}
}

func WithProductReTryTimes(max int) Config {
	return func(c *sarama.Config) {
		c.Producer.Retry.Max = max
	}
}
func WithVersion(version sarama.KafkaVersion) Config {
	return func(c *sarama.Config) {
		c.Version = version
	}
}

// 地址  是否是同步 配置
func NewSarProducer(addrs []string, is_sync bool, log logger.Logger, cfg ...Config) (Producer, error) {
	config := sarama.NewConfig()
	for _, c := range cfg {
		c(config)
	}
	// config.Net.SASL.
	if is_sync {
		config.Producer.Return.Successes = true
		p, err := sarama.NewSyncProducer(addrs, config)
		if err == nil {
			return &syncproducer{SyncProducer: p,
				log: log}, nil
		}
		return nil, err
	}
	p, err := sarama.NewAsyncProducer(addrs, config)
	if err == nil {
		return &asyncproducer{AsyncProducer: p,
			log: log}, nil
	}
	// sarama.ConsumerMessage
	return nil, err
}
func WithSASLAuth(user string, pwd string, algorithm Algorithm) Config {
	return func(conf *sarama.Config) {
		if user != "" && pwd != "" {
			conf.Net.SASL.Enable = true
			conf.Net.SASL.User = user
			conf.Net.SASL.Password = pwd
			conf.Net.SASL.Handshake = true
			if algorithm == SHA_256 {
				conf.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &xDGSCRAMClient{HashGeneratorFcn: sHA512} }
				conf.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
			} else if algorithm == SHA_512 {
				conf.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &xDGSCRAMClient{HashGeneratorFcn: sHA256} }
				conf.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
			}
		}
	}
}

// , useTLS bool
func WithTls(certFile string, keyFile string, caFile string, skip bool) Config {
	return func(cfg *sarama.Config) {
		cfg.Net.TLS.Enable = true
		cfg.Net.TLS.Config = createTLSConfiguration(certFile, keyFile, caFile, skip)
	}
}

//消费者

type Consumer interface {
	SubscribeTopics([]string, Handler) error
	Close(topic string) error
	Pause()
	Resume()
}
type comsumer struct {
	addr  []string
	group string
	cfg   *sarama.Config
	log   logger.Logger
	context.Context
	autoCommit bool
	topics     sync.Map //map[string]Handler
}
type Handler func(topic string, partition int32, offset int64, msg []byte) error
type topicHandler struct {
	h Handler
	sarama.ConsumerGroup
	cancel context.CancelFunc
	sync.Once
}

func (c *comsumer) Pause() {
	c.topics.Range(func(key, value any) bool {
		h, ok := value.(*topicHandler)
		if ok {
			h.ConsumerGroup.PauseAll()
		}
		return true
	})
}

func (c *comsumer) Resume() {
	c.topics.Range(func(key, value any) bool {
		h, ok := value.(*topicHandler)
		if ok {
			h.ConsumerGroup.ResumeAll()
		}
		return true
	})
}
func (c *comsumer) SubscribeTopics(topic []string, h Handler) error {
	str := ""
	for _, val := range topic {
		str += fmt.Sprintf("[%s]", val)
	}
	if str == "" {
		return fmt.Errorf("no topic")
	}
	_, ok := c.topics.Load(str)
	if !ok {
		client, err := sarama.NewClient(c.addr, c.cfg)
		if err != nil {
			return err
		}
		group_client, err := sarama.NewConsumerGroupFromClient(c.group, client)
		if err != nil {
			return err
		}
		newCtx, cancleFun := context.WithCancel(c.Context)
		th := &topicHandler{h: h, ConsumerGroup: group_client, cancel: cancleFun}
		//输出错误
		go func(ctx context.Context, topH *topicHandler) {
			for {
				select {
				case <-ctx.Done():
					topH.Once.Do(func() {
						topH.Close()
					})
					return
				case err := <-group_client.Errors():
					if c.log != nil {
						c.log.Error(KafkaSaramaTag, "err", err)
					}
				}
			}
		}(newCtx, th)
		//消费消息
		go func(t []string, ctx context.Context, topH *topicHandler) {
			for {
				select {
				case <-ctx.Done():
					topH.Once.Do(func() {
						topH.Close()
					})
					return
				default:
				}
				err := topH.Consume(c.Context, t, c)
				if c.log != nil {
					c.log.Error(KafkaSaramaTag, "err", err)
				}
			}
		}(topic, newCtx, th)
		c.topics.Store(str, th)
	}
	return nil
}

func (c *comsumer) Close(topic string) error {
	cli, ok := c.topics.Load(topic)
	if ok {
		t, ok := cli.(*topicHandler)
		if ok {
			t.Once.Do(func() {
				t.cancel()
			})
			return t.Close()
		}
		return fmt.Errorf("cant conver to ConsumerGroup")
	}
	return nil
}
func (c *comsumer) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}
func (c *comsumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}
func (c *comsumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		c.topics.Range(func(key, value any) bool {
			k := key.(string)
			if strings.Contains(k, fmt.Sprintf("[%s]", msg.Topic)) {
				h := value.(*topicHandler)
				if h.h(msg.Topic, msg.Partition, msg.Offset, msg.Value) == nil {
					sess.MarkMessage(msg, "")
					if !c.autoCommit {
						sess.Commit()
					}
				}
			}
			return true
		})
	}
	return nil
}

// func (c *comsumer) Close() error {
// 	return c.Close()
// }

type Config func(*sarama.Config)

func WithConsumerAutoCommit(is_auto bool) Config {
	return func(c *sarama.Config) {
		c.Consumer.Offsets.AutoCommit.Enable = is_auto
	}
}

func WithConsumerAutoInterval(Interval time.Duration) Config {
	return func(c *sarama.Config) {
		c.Consumer.Offsets.AutoCommit.Interval = Interval
	}
}

func WithConsumerAutoDelGroup(timeout time.Duration) Config {
	return func(c *sarama.Config) {
		c.Consumer.Group.Session.Timeout = timeout
	}
}

type OffsetType int64

const (
	OffsetOldest OffsetType = OffsetType(sarama.OffsetOldest)
	OffsetNewest OffsetType = OffsetType(sarama.OffsetNewest)
)

func WithConsumerOffsets(i OffsetType) Config {
	return func(c *sarama.Config) {
		c.Consumer.Offsets.Initial = int64(i)
	}
}

func NewSarConsumer(addrs []string, group string, log logger.Logger, cfg ...Config) (Consumer, error) {
	config := sarama.NewConfig()
	// config.Consumer.Offsets.AutoCommit.Enable = false
	for _, c := range cfg {
		c(config)
	}
	config.Consumer.Return.Errors = true
	return &comsumer{
		group:      group,
		addr:       addrs,
		cfg:        config,
		log:        log,
		Context:    context.Background(),
		autoCommit: config.Consumer.Offsets.AutoCommit.Enable,
	}, nil
}
