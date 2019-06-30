package runner

import "sync"

type Runner interface {
	Run(func() error)
	Stopping() <-chan struct{}
	Error(error)
	Stop() error
	Errors() []error
}

type runner struct {
	stopping chan struct{}

	errorsMutex sync.Mutex
	errors      []error

	wg sync.WaitGroup
}

func New() Runner {
	return &runner{
		stopping: make(chan struct{}),
	}
}

func (r *runner) closeStopping() {
	// So we don't close an already closed channel.
	select {
	case <-r.stopping:
	default:
		close(r.stopping)
	}

}

func (r *runner) Run(f func() error) {
	r.wg.Add(1)
	go func() {
		if err := f(); err != nil {
			r.errorsMutex.Lock()
			r.errors = append(r.errors, err)
			r.errorsMutex.Unlock()
			r.closeStopping()
		}
		r.wg.Done()
	}()
}

func (r *runner) Stopping() <-chan struct{} {
	return r.stopping
}

func (r *runner) Error(err error) {
	r.errorsMutex.Lock()
	r.errors = append(r.errors, err)
	r.errorsMutex.Unlock()
	r.closeStopping()
}

func (r *runner) Stop() error {
	r.closeStopping()
	r.wg.Wait()
	if len(r.errors) > 0 {
		return r.errors[0]
	}
	return nil
}

func (r *runner) Errors() []error {
	return r.errors
}
