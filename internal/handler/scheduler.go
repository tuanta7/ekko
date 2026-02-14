package handler

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrEnqueueTimeout  = errors.New("enqueue timeout")
	ErrQueueClosed     = errors.New("scheduler closed")
	ErrSchedulerClosed = errors.New("scheduler closed")
)

type Job struct {
	SeqNumber uint32
	RawData   []float32
}

type JobHandler func(ctx context.Context, job *Job) error

type Scheduler struct {
	wg          sync.WaitGroup
	mu          sync.Mutex
	closed      atomic.Bool
	queue       chan *Job
	workerCount int
}

func New(bufferSize int) *Scheduler {
	return &Scheduler{
		queue:       make(chan *Job, bufferSize),
		workerCount: max(1, runtime.NumCPU()/2),
	}
}

// Start launches worker goroutines that process jobs using the provided handler.
func (s *Scheduler) Start(ctx context.Context, handler JobHandler) {
	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.worker(ctx, handler)
	}
}

func (s *Scheduler) worker(ctx context.Context, handler JobHandler) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-s.queue:
			if !ok {
				return
			}
			_ = handler(ctx, job)
		}
	}
}

// Enqueue adds a job to the scheduler with a timeout.
func (s *Scheduler) Enqueue(ctx context.Context, job *Job) error {
	if s.closed.Load() {
		return ErrSchedulerClosed
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	select {
	case s.queue <- job:
		return nil
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return ErrEnqueueTimeout
		}
		return ctx.Err()
	}
}

// Dequeue retrieves a job from the scheduler.
func (s *Scheduler) Dequeue(ctx context.Context) (*Job, error) {
	select {
	case job, ok := <-s.queue:
		if !ok {
			return nil, ErrQueueClosed
		}
		return job, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Close gracefully shuts down the scheduler.
func (s *Scheduler) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed.Swap(true) {
		return ErrSchedulerClosed
	}

	close(s.queue)
	s.wg.Wait()
	return nil
}

// Len returns the current number of jobs in the scheduler.
func (s *Scheduler) Len() int {
	return len(s.queue)
}

// IsClosed returns whether the scheduler has been closed.
func (s *Scheduler) IsClosed() bool {
	return s.closed.Load()
}
