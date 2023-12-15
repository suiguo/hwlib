package rocket

import (
	"context"
	"errors"
	"time"

	"github.com/apache/rocketmq-clients/golang"
	"github.com/apache/rocketmq-clients/golang/credentials"
	"github.com/goccy/go-json"
)

const (
	Topic     = "xxxxxx"
	GroupName = "xxxxxx"
	Endpoint  = "xxxxxx"
	Region    = "xxxxxx"
	AccessKey = "xxxxxx"
	SecretKey = "xxxxxx"
)

// 异步处理函数
type ErrHandler func(data any, err error)
type Producer interface {
	SendSync(msg any, delay time.Duration) error
	SendASync(data any, delay time.Duration, h ErrHandler)
	Close()
}
type rocketProducer struct {
	golang.Producer
}

func (r *rocketProducer) Close() {
	if r.Producer != nil {
		r.Producer.GracefulStop()
	}
}

// 异步发送
func (r *rocketProducer) SendASync(data any, delay time.Duration, h ErrHandler) {
	if r.Producer == nil {
		if h != nil {
			h(data, errors.New("Producer is nil"))
		}
		return
	}
	body, err := json.Marshal(data)
	if err != nil {
		if h != nil {
			h(data, err)
		}
		return
	}
	msg := &golang.Message{
		Topic: Topic,
		Body:  body,
	}
	msg.SetDelayTimestamp(time.Now().Add(delay))
	r.Producer.SendAsync(context.TODO(), nil, func(ctx context.Context, sr []*golang.SendReceipt, err error) {
		if h != nil {
			h(data, err)
		}
	})
}

// 同步发送
func (r *rocketProducer) SendSync(data any, delay time.Duration) error {
	if r.Producer == nil {
		return errors.New("Producer is nil")
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	msg := &golang.Message{
		Topic: Topic,
		Body:  body,
	}
	msg.SetDelayTimestamp(time.Now().Add(delay))
	_, err = r.Send(context.TODO(), msg)
	if err != nil {
		return err
	}
	return nil
}

func NewProducer(group string, namespace string, accessKey string, secretKey string, topic ...string) (Producer, error) {
	var credential *credentials.SessionCredentials
	if accessKey != "" && secretKey != "" {
		credential = &credentials.SessionCredentials{
			AccessKey:    AccessKey,
			AccessSecret: SecretKey,
		}
	}
	producer, err := golang.NewProducer(
		&golang.Config{
			Endpoint:      Endpoint,
			ConsumerGroup: group,
			NameSpace:     namespace,
			Credentials:   credential,
		},
		golang.WithTopics(topic...),
	)
	if err != nil {
		return nil, err
	}
	err = producer.Start()
	if err != nil {
		return nil, err
	}
	return &rocketProducer{Producer: producer}, nil
}
