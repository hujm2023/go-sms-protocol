package nioserver

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestAdmissionGateLimitAndInFlight(t *testing.T) {
	gate := newAdmissionGate(2)
	if err := gate.Acquire(context.Background()); err != nil {
		t.Fatalf("first Acquire() error = %v", err)
	}
	if err := gate.Acquire(context.Background()); err != nil {
		t.Fatalf("second Acquire() error = %v", err)
	}
	if got := gate.InFlight(); got != 2 {
		t.Fatalf("InFlight() = %d, want 2", got)
	}

	started := make(chan struct{})
	waiterResult := make(chan error, 1)
	go func() {
		close(started)
		waiterResult <- gate.Acquire(context.Background())
	}()
	<-started
	select {
	case err := <-waiterResult:
		t.Fatalf("waiting Acquire() returned before a permit was released: %v", err)
	default:
	}
	if got := gate.InFlight(); got != 2 {
		t.Fatalf("waiting Acquire() changed InFlight() to %d, want 2", got)
	}

	gate.Release()
	if err := receiveAdmissionGateResult(waiterResult); err != nil {
		t.Fatalf("waiting Acquire() error = %v", err)
	}
	if got := gate.InFlight(); got != 2 {
		t.Fatalf("InFlight() after waiter admission = %d, want 2", got)
	}

	gate.Release()
	gate.Release()
	gate.Close()
	if err := gate.Wait(context.Background()); err != nil {
		t.Fatalf("Wait() after drain error = %v", err)
	}
}

func TestAdmissionGateCanceledAcquireDoesNotConsumeCapacity(t *testing.T) {
	gate := newAdmissionGate(1)
	if err := gate.Acquire(context.Background()); err != nil {
		t.Fatalf("initial Acquire() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		close(started)
		result <- gate.Acquire(ctx)
	}()
	<-started
	cancel()
	if err := receiveAdmissionGateResult(result); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled Acquire() error = %v, want context.Canceled", err)
	}
	if got := gate.InFlight(); got != 1 {
		t.Fatalf("InFlight() after canceled waiter = %d, want 1", got)
	}

	gate.Release()
	if err := gate.Acquire(context.Background()); err != nil {
		t.Fatalf("Acquire() after canceled waiter error = %v", err)
	}
	gate.Release()
	gate.Close()
}

func TestAdmissionGateCloseUnblocksAndPreventsAdmission(t *testing.T) {
	gate := newAdmissionGate(1)
	if err := gate.Acquire(context.Background()); err != nil {
		t.Fatalf("initial Acquire() error = %v", err)
	}

	started := make(chan struct{})
	waiterResult := make(chan error, 1)
	go func() {
		close(started)
		waiterResult <- gate.Acquire(context.Background())
	}()
	<-started

	gate.Close()
	gate.Close()
	if err := receiveAdmissionGateResult(waiterResult); !errors.Is(err, ErrAdmissionClosed) {
		t.Fatalf("waiting Acquire() error = %v, want ErrAdmissionClosed", err)
	}
	if err := gate.Acquire(context.Background()); !errors.Is(err, ErrAdmissionClosed) {
		t.Fatalf("new Acquire() error = %v, want ErrAdmissionClosed", err)
	}
	if got := gate.InFlight(); got != 1 {
		t.Fatalf("InFlight() before final Release = %d, want 1", got)
	}

	gate.Release()
	if err := gate.Wait(context.Background()); err != nil {
		t.Fatalf("Wait() after final Release error = %v", err)
	}
}

func TestAdmissionGateCloseRaceRejectsTokenAfterClose(t *testing.T) {
	gate := newAdmissionGate(1)
	if err := gate.Acquire(context.Background()); err != nil {
		t.Fatalf("initial Acquire() error = %v", err)
	}

	started := make(chan struct{})
	waiterResult := make(chan error, 1)
	go func() {
		close(started)
		waiterResult <- gate.Acquire(context.Background())
	}()
	<-started

	gate.Close()
	gate.Release()
	if err := receiveAdmissionGateResult(waiterResult); !errors.Is(err, ErrAdmissionClosed) {
		t.Fatalf("Acquire() racing with Close() error = %v, want ErrAdmissionClosed", err)
	}
	if got := gate.InFlight(); got != 0 {
		t.Fatalf("InFlight() after close race = %d, want 0", got)
	}
	if err := gate.Wait(context.Background()); err != nil {
		t.Fatalf("Wait() after close race error = %v", err)
	}
}

func TestAdmissionGateWaitCancellationAndConcurrentClose(t *testing.T) {
	gate := newAdmissionGate(1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := gate.Wait(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Wait() on an open canceled gate = %v, want context.Canceled", err)
	}

	if err := gate.Acquire(context.Background()); err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}

	const waiterCount = 8
	started := make(chan struct{}, waiterCount)
	results := make(chan error, waiterCount)
	var waiters sync.WaitGroup
	waiters.Add(waiterCount)
	for i := 0; i < waiterCount; i++ {
		go func() {
			defer waiters.Done()
			started <- struct{}{}
			results <- gate.Wait(context.Background())
		}()
	}
	for i := 0; i < waiterCount; i++ {
		<-started
	}

	var closers sync.WaitGroup
	closers.Add(waiterCount)
	for i := 0; i < waiterCount; i++ {
		go func() {
			defer closers.Done()
			gate.Close()
		}()
	}
	closers.Wait()
	select {
	case err := <-results:
		t.Fatalf("Wait() returned before the final Release: %v", err)
	default:
	}

	gate.Release()
	waiters.Wait()
	for i := 0; i < waiterCount; i++ {
		if err := <-results; err != nil {
			t.Fatalf("concurrent Wait() error = %v", err)
		}
	}

	ctx, cancel = context.WithCancel(context.Background())
	cancel()
	if err := gate.Wait(ctx); err != nil {
		t.Fatalf("Wait() on an already drained gate = %v, want nil", err)
	}
}

func TestAdmissionGateNormalizesLimitAndSentinel(t *testing.T) {
	if got := ErrAdmissionClosed.Error(); got != "nioserver: admission closed" {
		t.Fatalf("ErrAdmissionClosed.Error() = %q, want nioserver: admission closed", got)
	}

	gate := newAdmissionGate(0)
	if err := gate.Acquire(context.Background()); err != nil {
		t.Fatalf("Acquire() with normalized limit error = %v", err)
	}
	if got := gate.InFlight(); got != 1 {
		t.Fatalf("InFlight() with normalized limit = %d, want 1", got)
	}
	gate.Release()
	gate.Close()
}

func TestAdmissionGateReleaseWithoutPermitPanics(t *testing.T) {
	gate := newAdmissionGate(1)
	defer func() {
		if recover() == nil {
			t.Fatal("Release() without a permit did not panic")
		}
	}()
	gate.Release()
}

func receiveAdmissionGateResult(results <-chan error) error {
	select {
	case err := <-results:
		return err
	case <-time.After(time.Second):
		return errors.New("timed out waiting for admission result")
	}
}
