package queue

type Pool struct {
	workers chan struct{}
}

func NewPool(size int) *Pool {
	return &Pool{
		workers: make(chan struct{}, size),
	}
}

func (p *Pool) Run(job func()) {
	p.workers <- struct{}{}
	go func() {
		defer func() { <-p.workers }()
		job()
	}()
}
