package queue

import "context"

type Job struct {
	cancel    context.CancelFunc
	isRunning bool
}
