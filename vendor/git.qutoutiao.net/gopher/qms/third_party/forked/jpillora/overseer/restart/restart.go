package restart

import (
	"fmt"
	"os/exec"
	"sync/atomic"
	"time"
)

var ErrTimeout = fmt.Errorf("reload deadline exceeded")

type Restarter interface {
	RestartBegin() error
	ProgramStart() error
	ProgramReady() (*exec.Cmd, error)
	ProgramFailed(err error) error
	RestartEnd(old *exec.Cmd) error
}

type Restart struct {
	state        int32
	restarter    Restarter
	readyC       chan struct{}
	failedC      chan error
	readyTimeout time.Duration
}

func New(restarter Restarter, timeout time.Duration) *Restart {
	return &Restart{
		restarter:    restarter,
		readyTimeout: timeout,
	}
}

func (r *Restart) Run() (err error) {
	// wait --> begin
	if !atomic.CompareAndSwapInt32(&r.state, wait, begin) {
		// already begin
		return nil
	}

	// defer reset
	defer r.reset()

	// begin
	if err = r.begin(); err != nil {
		return err
	}

	// start
	if err := r.start(); err != nil {
		return err
	}

	// wait [ready or failed]
	var old *exec.Cmd
	select {
	case <-r.readyC:
		old, err = r.ready()
	case e := <-r.failedC:
		err = r.failed(e)
	case <-time.After(r.readyTimeout):
		err = r.failed(ErrTimeout)
	}
	if err != nil {
		return err
	}

	// end
	if err = r.end(old); err != nil {
		return err
	}

	return nil
}

func (r *Restart) ProgramReady() bool {
	if state := atomic.LoadInt32(&r.state); state >= begin && state < ready {
		r.readyC <- struct{}{}
		return true
	}
	return false
}

func (r *Restart) ProgramFailed(err error) bool {
	if state := atomic.LoadInt32(&r.state); state >= begin && state < ready {
		r.failedC <- err
		return true
	}
	return false
}

func (r *Restart) begin() error {
	r.readyC = make(chan struct{}, 1)
	r.failedC = make(chan error, 1)

	if err := r.restarter.RestartBegin(); err != nil {
		return err
	}

	return nil
}

func (r *Restart) start() error {
	if err := r.restarter.ProgramStart(); err != nil {
		return err
	}

	atomic.StoreInt32(&r.state, start)
	return nil
}

func (r *Restart) ready() (*exec.Cmd, error) {
	old, err := r.restarter.ProgramReady()
	if err != nil {
		return nil, err
	}

	atomic.StoreInt32(&r.state, ready)
	return old, nil
}

func (r *Restart) failed(e error) error {
	if err := r.restarter.ProgramFailed(e); err != nil {
		return err
	}

	atomic.StoreInt32(&r.state, failed)
	return nil
}

func (r *Restart) end(old *exec.Cmd) error {
	if err := r.restarter.RestartEnd(old); err != nil {
		return err
	}

	atomic.StoreInt32(&r.state, end)
	return nil
}

func (r *Restart) reset() {
	close(r.readyC)
	close(r.failedC)

	r.readyC = nil
	r.failedC = nil
	atomic.StoreInt32(&r.state, wait)
}
