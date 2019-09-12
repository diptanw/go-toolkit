/*
Package worker provides a pool of asynchronous workers for executing
concurrent tasks in background with a certain parallelism.

Basic Usage

	p := NewPool(3)

	ctx, cancel := context.WithCancel(context.Background())

	for _, item := range items {
		p.Enqueue(ctx, func() {
			// some useful job here
			<-ctx.Done()
		})
	}

	// when exiting an application
	p.Close()
	cancel()
*/
package worker
