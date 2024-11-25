package queue

import (
	"fmt"
	"testing"
	"time"
)

func TestQue(t *testing.T) {
	q := NewQueue[int](500)
	now := time.Now()
	for i := 0; i < 10000000; i++ {
		q.Enqueue(&i)
	}
	fmt.Println("Enqueue 10000000 values:", time.Since(now))
	now = time.Now()
	for i := 0; i < 10000000; i++ {
		q.Dequeue()
	}
	fmt.Println("Dequeue 10000000 values:", time.Since(now))
	now = time.Now()
	for i := 0; i < 10000000; i++ {
		q.Dequeue()
	}
	fmt.Println("Dequeue 10000000 values:", time.Since(now))
	now = time.Now()
	for i := 0; i < 10000000; i++ {
		// if i%3 == 0 {
		// }
		q.Enqueue(&i)
		q.Dequeue()
	}
	fmt.Println("Dequeue 10000000 values and Enqueue 10000000 values:", time.Since(now))
}
