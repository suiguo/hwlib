package stream

import (
	"context"
	"fmt"
	"time"

	"github.com/suiguo/hwlib/logger"
	redis_cli "github.com/suiguo/hwlib/redis"

	"github.com/redis/go-redis/v9"
)

var s *StreamClient

type HandlerStream func(stream_name string, event_name string, data map[string]interface{}) bool

func Init(r *redis_cli.Client,
	streamkey string,
	group string,
	consumer string,
	log logger.Logger,
	h HandlerStream) bool {
	if r != nil && s == nil {
		s = &StreamClient{cli: r}
		err := s.Run(streamkey, group, consumer, log, h)
		if err == nil {
			return true
		}
	}
	return false
}

type StreamClient struct {
	cli *redis_cli.Client
}

func (s *StreamClient) Run(streamkey string, group string, consumer string, log logger.Logger, h HandlerStream) error {
	if s == nil || s.cli == nil {
		return fmt.Errorf("stream run error")
	}
	s.cli.Cc.XGroupCreateMkStream(context.Background(), streamkey, group, "0")
	go func() {
		for {
			r, err := s.cli.Cc.XReadGroup(context.Background(), &redis.XReadGroupArgs{
				Group:    group,
				Consumer: consumer,
				Count:    10,
				Streams:  []string{streamkey, ">"},
				Block:    0,
				NoAck:    false,
			}).Result()
			if err != nil {
				if log != nil {
					log.Error("StreamError", err)
				}
				time.Sleep(time.Second * 10)
			} else {
				if log != nil {
					log.Info("Stream Work", "Group", group, "Consumer", consumer)
				}
				for _, data := range r {
					stream_name := data.Stream
					for _, s_data := range data.Messages {
						event_name, ok := s_data.Values["event"]
						ev_name, ok2 := event_name.(string)
						if ok && ok2 {
							if h != nil && h(stream_name, ev_name, s_data.Values) {
								s.cli.Cc.XAck(context.Background(), streamkey, group, s_data.ID)
							}
						}
					}
				}
			}
		}
	}()
	return nil
}
