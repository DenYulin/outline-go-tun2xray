package runner

import (
	"sync"
	"sync/atomic"
)

// S is a function that will return true if the
// goroutine should stop executing.
type S func() bool

const (
	FALSE int32 = 0
	TRUE  int32 = 1
)

// Go executes the function in a goroutine and returns a
// Task capable of stopping the execution.
func Go(fn func(S) error) *Task {
	var run, stop int32
	t := &Task{
		stopChan:   make(chan struct{}),
		running:    &run,
		shouldStop: &stop,
	}
	atomic.StoreInt32(t.shouldStop, FALSE)
	atomic.StoreInt32(t.running, TRUE)
	go func() {
		// call the target function
		err := fn(func() bool {
			// this is the shouldStop() function available to the
			// target function
			shouldStop := atomic.LoadInt32(t.shouldStop)
			return shouldStop == TRUE
		})
		t.err.Store(err)
		atomic.StoreInt32(t.running, FALSE)
		var closeOnce sync.Once

		closeOnce.Do(t.closeOnceBody)
	}()
	return t
}

// Task represents an interrupted goroutine.
type Task struct {
	ID         string
	stopChan   chan struct{}
	shouldStop *int32
	running    *int32
	err        atomic.Value
}

func (t *Task) closeOnceBody() {
	// channel closing is actually a sending operation
	close(t.stopChan)
}
