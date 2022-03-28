package async

import (
	"context"
	"sync"
	"time"

	"github.com/diptanw/go-toolkit/logger"
)

type Scheduler struct {
	cancels   []context.CancelFunc
	cancelsMu sync.Mutex
	logger    logger.Logger
}

func NewScheduler(log logger.Logger) *Scheduler {
	return &Scheduler{
		logger: log,
	}
}

// Schedule spawns configured number of parallel workers running in the pool.
func (p *Scheduler) Schedule(ctx context.Context, interval time.Duration, fn asyncJobFn) {
	p.cancelsMu.Lock()
	defer p.cancelsMu.Unlock()

	ctx, cancel := context.WithCancel(ctx)

	go func(ctx context.Context) {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := fn(ctx); err != nil {
					p.logger.Errorf("job: %s", err)
				}
			}
		}
	}(ctx)

	p.cancels = append(p.cancels, cancel)
}

// Close cancels all running scheduler jobs.
func (p *Scheduler) Close() {
	p.cancelsMu.Lock()
	defer p.cancelsMu.Unlock()

	for _, cancel := range p.cancels {
		cancel()
	}

	p.cancels = nil
}
