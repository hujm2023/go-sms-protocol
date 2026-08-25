package nioserver

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestSerialWriterSerializesWrites(t *testing.T) {
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	var mu sync.Mutex
	var calls, active, maxActive int
	var seen [][]byte

	w := newSerialWriter(2, func(data []byte) error {
		mu.Lock()
		calls++
		active++
		if active > maxActive {
			maxActive = active
		}
		seen = append(seen, append([]byte(nil), data...))
		call := calls
		mu.Unlock()

		if call == 1 {
			close(firstStarted)
			<-releaseFirst
		}

		mu.Lock()
		active--
		mu.Unlock()
		return nil
	})

	firstDone := make(chan error, 1)
	go func() { firstDone <- w.Write(context.Background(), []byte("first")) }()
	<-firstStarted

	secondDone := make(chan error, 1)
	go func() { secondDone <- w.Write(context.Background(), []byte("second")) }()
	waitSerialWriterPending(t, w, 2)

	select {
	case <-secondDone:
		t.Fatal("second write ran before the first write was released")
	default:
	}
	close(releaseFirst)

	if err := waitSerialWriterResult(t, firstDone); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := waitSerialWriterResult(t, secondDone); err != nil {
		t.Fatalf("second write: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if maxActive != 1 {
		t.Fatalf("writeFn ran concurrently: max active = %d", maxActive)
	}
	if len(seen) != 2 || string(seen[0]) != "first" || string(seen[1]) != "second" {
		t.Fatalf("write order = %q, want [first second]", seen)
	}
}

func TestSerialWriterQueueFullAndMaxPendingNormalization(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	w := newSerialWriter(0, func([]byte) error {
		close(started)
		<-release
		return nil
	})

	firstDone := make(chan error, 1)
	go func() { firstDone <- w.Write(context.Background(), []byte("first")) }()
	<-started

	if err := w.Write(context.Background(), []byte("second")); !errors.Is(err, ErrWriteQueueFull) {
		t.Fatalf("full admission error = %v, want %v", err, ErrWriteQueueFull)
	}
	close(release)
	if err := waitSerialWriterResult(t, firstDone); err != nil {
		t.Fatalf("first write: %v", err)
	}
}

func TestSerialWriterCanceledWaiterReleasesCapacity(t *testing.T) {
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	var mu sync.Mutex
	var calls int
	w := newSerialWriter(2, func([]byte) error {
		mu.Lock()
		calls++
		call := calls
		mu.Unlock()
		if call == 1 {
			close(firstStarted)
			<-releaseFirst
		}
		return nil
	})

	firstDone := make(chan error, 1)
	go func() { firstDone <- w.Write(context.Background(), []byte("first")) }()
	<-firstStarted

	ctx, cancel := context.WithCancel(context.Background())
	secondDone := make(chan error, 1)
	go func() { secondDone <- w.Write(ctx, []byte("second")) }()
	waitSerialWriterPending(t, w, 2)
	cancel()
	if err := waitSerialWriterResult(t, secondDone); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled waiter error = %v, want %v", err, context.Canceled)
	}

	thirdDone := make(chan error, 1)
	go func() { thirdDone <- w.Write(context.Background(), []byte("third")) }()
	waitSerialWriterPending(t, w, 2)
	close(releaseFirst)
	if err := waitSerialWriterResult(t, firstDone); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := waitSerialWriterResult(t, thirdDone); err != nil {
		t.Fatalf("third write after canceled waiter: %v", err)
	}
}

func TestSerialWriterCopiesQueuedInput(t *testing.T) {
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	secondSeen := make(chan []byte, 1)
	var mu sync.Mutex
	var calls int
	w := newSerialWriter(2, func(data []byte) error {
		mu.Lock()
		calls++
		call := calls
		mu.Unlock()
		if call == 1 {
			close(firstStarted)
			<-releaseFirst
			return nil
		}
		secondSeen <- append([]byte(nil), data...)
		return nil
	})

	firstDone := make(chan error, 1)
	go func() { firstDone <- w.Write(context.Background(), []byte("first")) }()
	<-firstStarted

	queued := []byte("queued")
	secondDone := make(chan error, 1)
	go func() { secondDone <- w.Write(context.Background(), queued) }()
	waitSerialWriterPending(t, w, 2)
	copy(queued, "changed")
	close(releaseFirst)

	if err := waitSerialWriterResult(t, firstDone); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := waitSerialWriterResult(t, secondDone); err != nil {
		t.Fatalf("second write: %v", err)
	}
	select {
	case got := <-secondSeen:
		if string(got) != "queued" {
			t.Fatalf("queued write data = %q, want queued", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for queued write data")
	}
}

func TestSerialWriterStopRejectsQueuedAndNewWrites(t *testing.T) {
	activeStarted := make(chan struct{})
	releaseActive := make(chan struct{})
	w := newSerialWriter(2, func([]byte) error {
		close(activeStarted)
		<-releaseActive
		return nil
	})

	activeDone := make(chan error, 1)
	go func() { activeDone <- w.Write(context.Background(), []byte("active")) }()
	<-activeStarted
	queuedDone := make(chan error, 1)
	go func() { queuedDone <- w.Write(context.Background(), []byte("queued")) }()
	waitSerialWriterPending(t, w, 2)

	// BeginStop is deliberately non-blocking even while the active write is held.
	w.BeginStop()
	w.BeginStop()
	select {
	case err := <-activeDone:
		t.Fatalf("active write completed during BeginStop: %v", err)
	default:
	}
	select {
	case err := <-queuedDone:
		if !errors.Is(err, ErrConnectionClosed) {
			t.Fatalf("queued write after BeginStop = %v, want %v", err, ErrConnectionClosed)
		}
	case <-time.After(time.Second):
		t.Fatal("queued write was not woken by BeginStop")
	}
	if err := w.Write(context.Background(), []byte("new")); !errors.Is(err, ErrConnectionClosed) {
		t.Fatalf("new write after BeginStop = %v, want %v", err, ErrConnectionClosed)
	}

	stopDone := make(chan error, 1)
	go func() { stopDone <- w.Stop(context.Background()) }()
	select {
	case err := <-stopDone:
		t.Fatalf("Stop returned while active write was blocked: %v", err)
	default:
	}
	close(releaseActive)
	if err := waitSerialWriterResult(t, activeDone); err != nil {
		t.Fatalf("active write: %v", err)
	}
	if err := waitSerialWriterResult(t, stopDone); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if err := w.Stop(context.Background()); err != nil {
		t.Fatalf("repeated Stop: %v", err)
	}
}

func TestSerialWriterStopContextCancellationAndWriteError(t *testing.T) {
	activeStarted := make(chan struct{})
	releaseActive := make(chan struct{})
	wantWriteErr := errors.New("write failed")
	w := newSerialWriter(1, func([]byte) error {
		close(activeStarted)
		<-releaseActive
		return wantWriteErr
	})

	writeDone := make(chan error, 1)
	go func() { writeDone <- w.Write(context.Background(), []byte("active")) }()
	<-activeStarted

	stopCtx, cancelStop := context.WithCancel(context.Background())
	stopDone := make(chan error, 1)
	go func() { stopDone <- w.Stop(stopCtx) }()
	waitSerialWriterStopped(t, w)
	cancelStop()
	if err := waitSerialWriterResult(t, stopDone); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled Stop = %v, want %v", err, context.Canceled)
	}

	close(releaseActive)
	if err := waitSerialWriterResult(t, writeDone); !errors.Is(err, wantWriteErr) {
		t.Fatalf("writeFn error = %v, want %v", err, wantWriteErr)
	}
	if err := w.Stop(context.Background()); err != nil {
		t.Fatalf("draining Stop after cancellation: %v", err)
	}
}

func waitSerialWriterPending(t *testing.T, w *serialWriter, want int) {
	t.Helper()
	deadline := time.NewTimer(time.Second)
	defer deadline.Stop()
	for {
		w.mu.Lock()
		if w.pending >= want {
			w.mu.Unlock()
			return
		}
		waitCh := w.notify
		w.mu.Unlock()
		select {
		case <-waitCh:
		case <-deadline.C:
			t.Fatalf("pending writes did not reach %d", want)
		}
	}
}

func waitSerialWriterStopped(t *testing.T, w *serialWriter) {
	t.Helper()
	// Stop calls BeginStop before it waits; this barrier gives that state
	// transition a deterministic observation point without polling or sleeps.
	deadline := time.NewTimer(time.Second)
	defer deadline.Stop()
	for {
		w.mu.Lock()
		if w.stopped {
			w.mu.Unlock()
			return
		}
		waitCh := w.notify
		w.mu.Unlock()
		select {
		case <-deadline.C:
			t.Fatal("writer did not enter stopped state")
		case <-waitCh:
		}
	}
}

func waitSerialWriterResult(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for serial writer operation")
		return nil
	}
}
