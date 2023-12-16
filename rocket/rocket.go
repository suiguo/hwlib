package rocket

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/apache/rocketmq-clients/golang"
	"github.com/apache/rocketmq-clients/golang/credentials"
	"github.com/goccy/go-json"
)

// 异步处理函数
type ErrHandler func(data any, err error)
type Producer interface {
	SendSync(topic string, msg any, delay time.Duration) error
	SendASync(topic string, data any, delay time.Duration, h ErrHandler)
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
func (r *rocketProducer) SendASync(topic string, data any, delay time.Duration, h ErrHandler) {
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
		Topic: topic,
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
func (r *rocketProducer) SendSync(topic string, data any, delay time.Duration) error {
	if r.Producer == nil {
		return errors.New("Producer is nil")
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	msg := &golang.Message{
		Topic: topic,
		Body:  body,
	}
	msg.SetDelayTimestamp(time.Now().Add(delay))
	_, err = r.Send(context.TODO(), msg)
	if err != nil {
		return err
	}
	return nil
}

func NewProducer(endpoint string, group string, namespace string, accessKey string, secretKey string, topic ...string) (Producer, error) {
	os.Setenv("mq.consoleAppender.enabled", "true")
	golang.ResetLogger()
	// golang.ResetLogger()
	producer, err := golang.NewProducer(
		&golang.Config{
			Endpoint:      endpoint,
			ConsumerGroup: group,
			NameSpace:     namespace,
			Credentials: &credentials.SessionCredentials{
				AccessKey:    accessKey,
				AccessSecret: secretKey,
			},
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
