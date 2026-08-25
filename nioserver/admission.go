package nioserver

import (
	"context"
	"errors"
	"sync"
)

// ErrAdmissionClosed indicates that a gate no longer accepts new work.
var ErrAdmissionClosed = errors.New("nioserver: admission closed")

// admissionGate bounds the number of admitted tasks and can be closed once
// the owner begins draining. A token is held from a successful Acquire until
// its matching Release; tasks waiting for a token are not included in
// inFlight.
type admissionGate struct {
	mu sync.Mutex

	tokens chan struct{}

	closedCh  chan struct{}
	drainedCh chan struct{}

	closed   bool
	drained  bool
	inFlight int
}

// newAdmissionGate creates a gate with a bounded number of permits. A
// non-positive limit is normalized to one so that the gate always has a
// usable capacity.
func newAdmissionGate(limit int) *admissionGate {
	if limit < 1 {
		limit = 1
	}

	tokens := make(chan struct{}, limit)
	for i := 0; i < limit; i++ {
		tokens <- struct{}{}
	}

	return &admissionGate{
		tokens:    tokens,
		closedCh:  make(chan struct{}),
		drainedCh: make(chan struct{}),
	}
}

// Acquire waits for one permit, or returns when the context is canceled or
// the gate is closed. The state check after receiving a token is the
// synchronization point that prevents admission from crossing Close.
func (g *admissionGate) Acquire(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if g.isClosed() {
		return ErrAdmissionClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	select {
	case <-g.closedCh:
		return ErrAdmissionClosed
	case <-ctx.Done():
		// Prefer a close result if Close won the race with cancellation.
		if g.isClosed() {
			return ErrAdmissionClosed
		}
		return ctx.Err()
	case <-g.tokens:
		g.mu.Lock()
		if g.closed {
			g.mu.Unlock()
			g.returnToken()
			return ErrAdmissionClosed
		}
		if err := ctx.Err(); err != nil {
			g.mu.Unlock()
			g.returnToken()
			return err
		}
		g.inFlight++
		g.mu.Unlock()
		return nil
	}
}

// Release returns one previously acquired permit. Calling Release without a
// matching successful Acquire is a programmer error and panics.
func (g *admissionGate) Release() {
	g.mu.Lock()
	if g.inFlight == 0 {
		g.mu.Unlock()
		panic("nioserver: admission release without acquired permit")
	}
	g.inFlight--
	if g.closed && g.inFlight == 0 {
		g.markDrainedLocked()
	}
	g.mu.Unlock()

	// A valid permit always leaves room in the buffered channel. Keep this
	// send outside the mutex so a broken caller cannot block gate state.
	g.returnToken()
}

// Close stops admission and wakes all current and future Acquire calls. It is
// safe to call concurrently and repeatedly.
func (g *admissionGate) Close() {
	g.mu.Lock()
	if !g.closed {
		g.closed = true
		close(g.closedCh)
		if g.inFlight == 0 {
			g.markDrainedLocked()
		}
	}
	g.mu.Unlock()
}

// Wait waits for Close and for all successfully acquired permits to be
// released. Cancellation only affects a wait that has not already observed a
// drained gate.
func (g *admissionGate) Wait(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// Preserve the immediate-completion contract when the gate is already
	// drained, even if the supplied context is already canceled.
	select {
	case <-g.drainedCh:
		return nil
	default:
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	select {
	case <-g.drainedCh:
		return nil
	case <-ctx.Done():
		// If draining completed at the same time as cancellation, preserve
		// the already-drained result.
		select {
		case <-g.drainedCh:
			return nil
		default:
		}
		return ctx.Err()
	}
}

// InFlight reports the number of permits currently held by admitted tasks.
func (g *admissionGate) InFlight() int {
	g.mu.Lock()
	inFlight := g.inFlight
	g.mu.Unlock()
	return inFlight
}

func (g *admissionGate) isClosed() bool {
	g.mu.Lock()
	closed := g.closed
	g.mu.Unlock()
	return closed
}

func (g *admissionGate) markDrainedLocked() {
	if g.drained {
		return
	}
	g.drained = true
	close(g.drainedCh)
}

func (g *admissionGate) returnToken() {
	select {
	case g.tokens <- struct{}{}:
	default:
		panic("nioserver: admission token returned without acquisition")
	}
}
