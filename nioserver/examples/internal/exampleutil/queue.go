package exampleutil

import (
	"context"
	cryptorand "crypto/rand"
	"math/big"
	"sync"
	"time"
)

// Queue bounds queued submissions and worker goroutines. A job owns all recipients
// of one submission, so admission either accepts or rejects the complete batch.
type Queue[T any] struct {
	ctx  context.Context
	jobs chan T
	done chan struct{}
}

func NewQueue[T any](ctx context.Context, workers, capacity int, handle func(context.Context, T)) *Queue[T] {
	q := &Queue[T]{ctx: ctx, jobs: make(chan T, capacity), done: make(chan struct{})}
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for ctx.Err() == nil {
				select {
				case <-ctx.Done():
					return
				case job := <-q.jobs:
					handle(ctx, job)
				}
			}
		}()
	}
	go func() { wg.Wait(); close(q.done) }()
	return q
}

func (q *Queue[T]) TryEnqueue(job T) bool {
	if q.ctx.Err() != nil {
		return false
	}
	select {
	case q.jobs <- job:
		return true
	default:
		return false
	}
}

func (q *Queue[T]) Wait(ctx context.Context) error {
	select {
	case <-q.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func Delay(ctx context.Context, minimum, maximum time.Duration) bool {
	delay := minimum
	if span := maximum - minimum; span > 0 {
		if n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(span)+1)); err == nil {
			delay += time.Duration(n.Int64())
		}
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return ctx.Err() == nil
	}
}
