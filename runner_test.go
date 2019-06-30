package runner

import (
	"errors"
	"testing"
)

func TestRunner(t *testing.T) {
	dummyErr := errors.New("dummy error")

	r := New()

	stopping := false
	stopped := false
	r.Run(func() error {
		select {
		case <-r.Stopping():
			stopping = true
		}

		stopped = true
		return dummyErr
	})

	if err := r.Stop(); err != dummyErr {
		t.Fatal("Incorrect error.")
	}

	if !stopping {
		t.Fatal("Not stopping.")
	}

	if !stopped {
		t.Fatal("Not stopped.")
	}
}
