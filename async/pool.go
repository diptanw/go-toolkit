package async

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type asyncJobFn func(context.Context) error

// Pool is a simple worker group that runs a number of tasks at a configured
// concurrency.
type Pool struct {
	taskCh chan asyncJobFn
}

// NewPool initializes a new pool with a given concurrency.
func NewPool() Pool {
	return Pool{
		taskCh: make(chan asyncJobFn),
	}
}

// Run spawns configured number of parallel workers running in the pool.
func (p Pool) Run(ctx context.Context, workersNum int) error {
	errG, errCtx := errgroup.WithContext(ctx)

	for i := 0; i < workersNum; i++ {
		errG.Go(func() error {
			for t := range p.taskCh {
				select {
				case <-errCtx.Done():
					return nil
				default:
					if err := t(errCtx); err != nil {
						return err
					}
				}
			}

			return nil
		})
	}

	return errG.Wait()
}

// Enqueue adds new task to the tasks queue.
func (p Pool) Enqueue(task asyncJobFn) {
	p.taskCh <- task
}
