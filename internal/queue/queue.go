package queue

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrEnqueueTimeout = errors.New("enqueue timeout")
	ErrQueueClosed    = errors.New("queue closed")
)

type Message struct {
	Timestamp time.Time
	Data      map[string]string
}

type RecordQueue struct {
	queue     chan *Message
	closeOnce sync.Once
}

func NewRecordQueue() *RecordQueue {
	return &RecordQueue{
		queue:     make(chan *Message, 10),
		closeOnce: sync.Once{},
	}
}

func (q *RecordQueue) Enqueue(ctx context.Context, msg *Message) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	select {
	case q.queue <- msg:
		return nil
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return ErrEnqueueTimeout
		}
		return ctx.Err()
	}
}

func (q *RecordQueue) Dequeue(ctx context.Context) (*Message, error) {
	select {
	case msg, ok := <-q.queue:
		if !ok {
			return &Message{}, ErrQueueClosed
		}
		return msg, nil
	case <-ctx.Done():
		return &Message{}, ctx.Err()
	}
}

func (q *RecordQueue) Close() {
	q.closeOnce.Do(func() {
		close(q.queue)
	})
}
