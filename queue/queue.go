package queue

import (
	"fmt"
	"sync"
)

type Queue[T any] struct {
	data []*T
	idx  int
	head int
	sync.RWMutex
}

// NewQueue create a new queue with fixed length
// len must be greater than 0
// fifo queue
func NewQueue[T any](len int) *Queue[T] {
	if len <= 0 {
		len = 20
	}
	return &Queue[T]{data: make([]*T, len)}
}

// If the queue is full, the oldest value will be replaced
func (q *Queue[T]) Enqueue(v *T) (*T, error) {
	if v == nil {
		return nil, fmt.Errorf("value is nil")
	}
	q.Lock()
	defer q.Unlock()
	if q.isFull() {
		tmp := q.data[q.head]
		q.data[q.idx] = v
		q.idx = (q.idx + 1) % len(q.data)
		q.head = (q.head + 1) % len(q.data)
		return tmp, nil
	}
	q.data[q.idx] = v
	q.idx = (q.idx + 1) % len(q.data)
	return nil, nil
}

// Pop a value from the queue
// If the queue is empty, return an error
func (q *Queue[T]) Dequeue() (*T, error) {
	q.Lock()
	defer q.Unlock()
	if q.isEmpty() {
		return nil, fmt.Errorf("queue is empty")
	}
	return q.dequeueOne()
}
func (q *Queue[T]) dequeueOne() (*T, error) {
	if q.isEmpty() {
		return nil, fmt.Errorf("queue is empty")
	}
	v := q.data[q.head]
	q.data[q.head] = nil
	q.head = (q.head + 1) % len(q.data)
	return v, nil
}

func (q *Queue[T]) DequeueN(n int) ([]*T, error) {
	if n <= 0 {
		return nil, fmt.Errorf("n must be greater than 0")
	}
	tmp := make([]*T, 0, n)
	q.Lock()
	defer q.Unlock()
	for i := 0; i < n; i++ {
		v, err := q.dequeueOne()
		if err == nil {
			tmp = append(tmp, v)
		} else {
			break
		}
	}
	return tmp, nil
}
func (q *Queue[T]) Peek() (*T, error) {
	if q.isEmpty() {
		return nil, fmt.Errorf("queue is empty")
	}
	return q.data[q.head], nil
}
func (q *Queue[T]) Empty() bool {
	return q.isEmpty()
}

func (q *Queue[T]) Clear() {
	q.data = make([]*T, len(q.data))
	q.head = 0
	q.idx = 0
}

func (q *Queue[T]) Len() int {
	q.RLock()
	defer q.RUnlock()
	if q.isEmpty() {
		return 0
	}
	if q.head < q.idx {
		return q.idx - q.head
	}
	return len(q.data) - q.head + q.idx
}

func (q *Queue[T]) isFull() bool {
	return q.head == q.idx && q.data[q.head] != nil
}

func (q *Queue[T]) isEmpty() bool {
	return q.head == q.idx && q.data[q.head] == nil
}
