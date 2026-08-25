package nioserver

import (
	"context"
	"errors"
	"sync"
)

var (
	// ErrConnectionClosed indicates that a write was attempted after stopping admission.
	ErrConnectionClosed = errors.New("nioserver: connection closed")
	// ErrWriteQueueFull indicates that all pending write slots are occupied.
	ErrWriteQueueFull = errors.New("nioserver: write queue full")
)

// serialWriter bounds and serializes calls to a synchronous write function.
// pending includes the currently executing call and calls waiting for their turn.
type serialWriter struct {
	mu         sync.Mutex
	maxPending int
	writeFn    func([]byte) error

	pending int
	active  bool
	stopped bool
	notify  chan struct{}
}

func newSerialWriter(maxPending int, writeFn func([]byte) error) *serialWriter {
	if maxPending < 1 {
		maxPending = 1
	}
	return &serialWriter{
		maxPending: maxPending,
		writeFn:    writeFn,
		notify:     make(chan struct{}),
	}
}

// Write admits one bounded write, waits for its serialized turn, and invokes
// the configured write function synchronously.
func (w *serialWriter) Write(ctx context.Context, data []byte) error {
	if ctx == nil {
		ctx = context.Background()
	}
	w.mu.Lock()
	if w.stopped {
		w.mu.Unlock()
		return ErrConnectionClosed
	}
	if w.pending >= w.maxPending {
		w.mu.Unlock()
		return ErrWriteQueueFull
	}
	w.pending++
	// Keep a private copy for the entire lifetime of this admitted write. The
	// caller may reuse or mutate data while this call waits for its turn.
	dataCopy := append([]byte(nil), data...)
	w.signalLocked()
	w.mu.Unlock()

	for {
		w.mu.Lock()
		if w.stopped {
			w.pending--
			w.signalLocked()
			w.mu.Unlock()
			return ErrConnectionClosed
		}
		if !w.active {
			// A cancellation observed before claiming the turn releases this
			// call's admission instead of starting a write after cancellation.
			if err := ctx.Err(); err != nil {
				w.pending--
				w.signalLocked()
				w.mu.Unlock()
				return err
			}
			w.active = true
			w.mu.Unlock()

			// Keep the lock-free I/O boundary explicit. The deferred cleanup also
			// restores the writer state if a custom write function panics.
			defer w.finish()
			return w.writeFn(dataCopy)
		}
		waitCh := w.notify
		w.mu.Unlock()

		select {
		case <-waitCh:
			continue
		case <-ctx.Done():
			w.mu.Lock()
			if w.stopped {
				w.pending--
				w.signalLocked()
				w.mu.Unlock()
				return ErrConnectionClosed
			}
			w.pending--
			w.signalLocked()
			w.mu.Unlock()
			return ctx.Err()
		}
	}
}

// BeginStop closes admission and wakes all currently waiting writes. It does
// not wait for an active write or for queued calls to observe the stop.
func (w *serialWriter) BeginStop() {
	w.mu.Lock()
	if !w.stopped {
		w.stopped = true
		w.signalLocked()
	}
	w.mu.Unlock()
}

// Stop closes admission and waits for all admitted writes to drain. It is
// safe to call repeatedly; every call observes the same stopped state and
// drain condition.
func (w *serialWriter) Stop(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	w.BeginStop()

	for {
		w.mu.Lock()
		if w.pending == 0 {
			w.mu.Unlock()
			return nil
		}
		waitCh := w.notify
		w.mu.Unlock()

		select {
		case <-waitCh:
			continue
		case <-ctx.Done():
			w.mu.Lock()
			drained := w.pending == 0
			w.mu.Unlock()
			if drained {
				return nil
			}
			return ctx.Err()
		}
	}
}

func (w *serialWriter) finish() {
	w.mu.Lock()
	w.active = false
	w.pending--
	w.signalLocked()
	w.mu.Unlock()
}

// signalLocked wakes all waiters and gives subsequent waiters a fresh channel.
// The channel is only closed while mu is held, so no sender can race a close.
func (w *serialWriter) signalLocked() {
	close(w.notify)
	w.notify = make(chan struct{})
}
