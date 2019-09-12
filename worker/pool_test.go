package worker

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	p := NewPool(1)

	assert.NotNil(t, p)
	assert.NotNil(t, p.sem)
	assert.NotNil(t, p.wg)
	assert.False(t, p.closed)
}

func TestPool_Start_ctxCancellsAllTasks(t *testing.T) {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.TODO())
	p := NewPool(3)

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		p.Enqueue(ctx, func() {
			<-ctx.Done()
			wg.Done()
		})
	}

	cancel()
	wg.Wait()
}

func TestPool_Start_executesMaxParallelTasks(t *testing.T) {
	var count int32

	wg := sync.WaitGroup{}
	p := NewPool(3)

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		p.Enqueue(context.TODO(), func() {
			atomic.AddInt32(&count, 1)
			wg.Done()
			wg.Wait()
		})
	}

	wg.Wait()

	assert.Equal(t, int32(3), count)
}

func TestPool_Enqueue(t *testing.T) {
	p := NewPool(2)

	err := p.Enqueue(context.TODO(), func() {})

	assert.NoError(t, err)

	err = p.Enqueue(context.TODO(), nil)

	assert.Errorf(t, err, "task cannot be nil")
}

func TestPool_Close(t *testing.T) {
	p := NewPool(1)

	p.Close()
	err := p.Enqueue(context.TODO(), func() {})

	assert.Errorf(t, err, "pool is closed")
	assert.True(t, p.closed)
}

func TestPool_Close_waitsGroup(t *testing.T) {
	p := NewPool(1)
	syncCh := make(chan string)

	p.Enqueue(context.TODO(), func() {
		syncCh <- "started"
		syncCh <- "finished"
	})

	// wait for task to start
	assert.Equal(t, "started", <-syncCh)

	go func() {
		assert.Equal(t, "finished", <-syncCh)
	}()

	// close the pool, blocks until task finished
	p.Close()
}
