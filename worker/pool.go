package worker

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/sync/semaphore"
)

// Pool is a simple worker group that runs a number of tasks at
// a configured concurrency.
type Pool struct {
	closed bool
	wg     *sync.WaitGroup
	sem    *semaphore.Weighted
}

// NewPool spawns all workers with the given concurrency and returns a new pool.
func NewPool(workersNum int) *Pool {
	p := Pool{
		wg:  &sync.WaitGroup{},
		sem: semaphore.NewWeighted(int64(workersNum)),
	}

	return &p
}

// Enqueue adds new task to the tasks queue.
func (p *Pool) Enqueue(ctx context.Context, task func()) error {
	if task == nil {
		return errors.New("task cannot be nil")
	}

	if p.closed {
		return errors.New("pool is closed")
	}

	if err := p.sem.Acquire(ctx, 1); err != nil {
		return err
	}

	p.wg.Add(1)

	go func() {
		task()
		p.sem.Release(1)
		p.wg.Done()
	}()

	return nil
}

// Close waits for all workers to finish.
func (p *Pool) Close() {
	p.closed = true
	p.wg.Wait()
}
