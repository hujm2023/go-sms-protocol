package exampleutil

import (
	"context"
	"testing"
	"time"
)

func TestQueueBoundsWorkAndJoinsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{}, 1)
	q := NewQueue(ctx, 1, 1, func(ctx context.Context, _ int) { started <- struct{}{}; <-ctx.Done() })
	if !q.TryEnqueue(1) {
		t.Fatal("first job rejected")
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("worker did not start")
	}
	if !q.TryEnqueue(2) || q.TryEnqueue(3) {
		t.Fatal("queue capacity was not enforced")
	}
	cancel()
	waitCtx, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	if err := q.Wait(waitCtx); err != nil {
		t.Fatal(err)
	}
	if q.TryEnqueue(4) {
		t.Fatal("cancelled queue accepted work")
	}
}

func TestDelayCanBeCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if Delay(ctx, time.Hour, time.Hour) {
		t.Fatal("cancelled delay succeeded")
	}
	if !Delay(context.Background(), time.Nanosecond, time.Nanosecond) {
		t.Fatal("active delay failed")
	}
}
