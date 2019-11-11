package runner

import (
	"context"
	"sync"
)

type Runner interface {
	Run(func() error)
	Stopping() <-chan struct{}
	Context() context.Context

	Wait() error
	Stop() error
	Errors() []error
}

type runner struct {
	stoppingMutex sync.Mutex
	stopping      chan struct{}

	errorsMutex sync.Mutex
	errors      []error

	wg sync.WaitGroup

	ctx       context.Context
	cancelCtx func()
}

func New() Runner {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	return &runner{
		stopping:  make(chan struct{}),
		ctx:       ctx,
		cancelCtx: cancel,
	}
}

func (r *runner) stop() {
	// So we don't close an already closed channel.

	r.stoppingMutex.Lock()
	select {
	case <-r.stopping:
	default:
		close(r.stopping)
		r.cancelCtx()
	}
	r.stoppingMutex.Unlock()
}

func (r *runner) Run(f func() error) {
	r.wg.Add(1)
	go func() {
		if err := f(); err != nil {
			r.errorsMutex.Lock()
			r.errors = append(r.errors, err)
			r.errorsMutex.Unlock()
			r.stop()
		}
		r.wg.Done()
	}()
}

func (r *runner) Context() context.Context {
	return r.ctx
}

func (r *runner) Stopping() <-chan struct{} {
	return r.stopping
}

func (r *runner) Wait() error {
	r.wg.Wait()
	if len(r.errors) > 0 {
		return r.errors[0]
	}
	return nil
}

func (r *runner) Stop() error {
	r.stop()
	return r.Wait()
}

func (r *runner) Errors() []error {
	return r.errors
}
