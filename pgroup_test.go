package run_test

import (
	"errors"
	"github.com/oklog/run"
	"testing"
	"time"
)

func TestPGroup_Zero(t *testing.T) {
	var g run.PGroup
	res := make(chan error)
	go func() { res <- g.Run() }()
	select {
	case err := <-res:
		if err != nil {
			t.Errorf("%v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}
}

func TestPGroup_One(t *testing.T) {
	myError := errors.New("foobar")
	var g run.PGroup
	g.Add(func() error { return myError }, func(error) {}, -1)
	res := make(chan error)
	go func() { res <- g.Run() }()
	select {
	case err := <-res:
		if want, have := myError, err; want != have {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}
}

func TestPGroup_Many(t *testing.T) {
	interrupt := errors.New("interrupt")
	var g run.PGroup
	g.Add(func() error {
		t.Logf("pos: %d", 2)
		return interrupt
	}, func(error) {}, 2)
	cancel := make(chan struct{})
	g.Add(func() error {
		t.Logf("pos: %d", 1)
		<-cancel
		return nil
	}, func(error) { close(cancel) }, 1)
	res := make(chan error)
	go func() { res <- g.Run() }()
	select {
	case err := <-res:
		if want, have := interrupt, err; want != have {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("timeout")
	}
}
