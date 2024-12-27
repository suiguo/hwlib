package queue

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestConcurrentQueue(t *testing.T) {
	f := func() {
		queue := NewQueue[int](10)
		var wg sync.WaitGroup
		const numGoroutines = 200
		const numOperations = 20000
		b := time.Now()
		// 启动多个协程来并发地进行入队和出队操作
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < numOperations; j++ {
					_, err := queue.Enqueue(&j)
					if err != nil {
						t.Errorf("Failed to enqueue: %v", err)
					}
					// time.Sleep(time.Millisecond)
				}
			}()
		}
		wg.Wait()
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < numOperations; j++ {
					_, err := queue.Dequeue()
					if err != nil && err != ErrEmpty {
						t.Errorf("Failed to dequeue: %v", err)
					}
					// time.Sleep(time.Millisecond)
				}
			}()
		}

		// 等待所有协程完成
		wg.Wait()
		fmt.Printf("cost %s\n", time.Since(b))
		// 检查队列是否为空
		if !queue.Empty() {
			t.Errorf("Expected queue to be empty, but it contains %d ", queue.Len())
		}
	}
	var w sync.WaitGroup
	for i := 0; i < 100; i++ {
		w.Add(1)
		go func() {
			defer w.Done()
			f()
		}()
	}
	w.Wait()
}
