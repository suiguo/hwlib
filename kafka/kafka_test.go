package kafka

import (
	"fmt"
	"testing"
	"time"
)

func TestKafka(m *testing.T) {
	cli, err := GetDefaultKafka(ALLType, "localhost", "test_group", "", nil)
	if err != nil {
		fmt.Println(err)
	}
	// err = cli.Subscribe("topic_test")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// go func() {
	// 	data := cli.MessageChan()
	// 	for {
	// 		out_msg := <-data
	// 		fmt.Println("msg===", string(out_msg.Value))
	// 	}
	// }()
	number := 0
	for {
		time.Sleep(time.Second * 1)
		if number < 10 {
			err = cli.Produce("topic_test", &KafkaMsg{
				Msg: []byte("hello word"),
			})
			if err != nil {
				fmt.Println(err)
			}
		}
		number++
	}

}
