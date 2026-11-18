package worker

import (
	"sync"
	"context"

)

type WorkerPool struct {
	conn	*ampq.Connection
	ch 		*ampq.Channel
	wg sync.WaitGroup
}

func NewWorkerPool(conn *ampq.Connection) *WorkerPool {
	return &WorkerPool{}
}

func (wp *WorkerPool) Start(ctx context.Context, numWorkers int) {
	for i := 0; i < numWorkers; i ++ {
		wp.wg.Add(1)
		go wp.
	}
}

func (wp *WorkerPool) worker(ctx context.Context) {
	defer wp.wg.Done()
	log.Print()
}