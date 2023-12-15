package rocket

import (
	"fmt"
	"testing"
	"time"
)

func TestXxx(t *testing.T) {
	p, _ := NewProducer("", "test", "default_test", "", "")
	p.SendASync("ttt", "hello", time.Second*5, func(data any, err error) {
		fmt.Println(data, err)
	})
}
