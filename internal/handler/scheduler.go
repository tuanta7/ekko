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
	ErrSchedulerClosed = errors.New("scheduler closed")
)

type Job struct {
	SeqNumber uint32
	RawData   []float32
}

type JobHandler func(ctx context.Context, job *Job) error

// ErrorHandler is called when a job processing error occurs.
type ErrorHandler func(job *Job, err error)

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

// ProcessJobs launches worker goroutines that process jobs using the provided handler.
// It runs in the background and calls onError for any processing errors.
func (s *Scheduler) ProcessJobs(ctx context.Context, onJob JobHandler, onError ErrorHandler) {
	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.startWorker(ctx, onJob, onError)
	}
}

func (s *Scheduler) startWorker(ctx context.Context, handler JobHandler, onError ErrorHandler) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-s.queue:
			if !ok {
				return
			}
			if err := handler(ctx, job); err != nil {
				if onError != nil {
					onError(job, err)
				}
			}
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
