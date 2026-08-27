// Package workers provides a fixed-size goroutine pool with a bounded task
// queue, per-task timeout and retries. All sizing knobs come from
// config.C.Workers.
package workers

import (
	"context"
	"errors"
	"sync"

	"go.uber.org/zap"

	"github.com/blkst8/invoice-service/internal/config"
	"github.com/blkst8/invoice-service/internal/log"
)

// Task is a unit of work executed by the pool.
type Task interface {
	Name() string
	Execute(ctx context.Context) error
}

var ErrQueueFull = errors.New("workers queue is full")

// Pool is a fixed-size worker pool with a bounded queue.
type Pool struct {
	cfg   config.Workers
	tasks chan Task
	wg    sync.WaitGroup
}

func NewPool(cfg config.Workers) *Pool {
	return &Pool{
		cfg:   cfg,
		tasks: make(chan Task, cfg.QueueSize),
	}
}

// Start launches the configured number of worker goroutines.
func (p *Pool) Start() {
	for i := 0; i < p.cfg.Count; i++ {
		p.wg.Add(1)
		go p.loop(i)
	}
}

func (p *Pool) loop(id int) {
	defer p.wg.Done()

	for task := range p.tasks {
		for attempt := 0; attempt <= p.cfg.Retries; attempt++ {
			ctx, cancel := context.WithTimeout(context.Background(), p.cfg.Timeout)
			err := task.Execute(ctx)
			cancel()
			if err == nil {
				break
			}
			log.Logger.Error("task failed",
				zap.Int("worker_id", id),
				zap.String("task", task.Name()),
				zap.Int("attempt", attempt+1),
				zap.Error(err),
			)
		}
	}
}

// Submit enqueues a task without blocking.
// Returns ErrQueueFull when the queue has no capacity left.
// Must not be called after Stop.
func (p *Pool) Submit(task Task) error {
	select {
	case p.tasks <- task:
		return nil
	default:
		log.Logger.Error("submit failed", zap.String("task", task.Name()), zap.Error(ErrQueueFull))
		return ErrQueueFull
	}
}

// Stop closes the queue and waits until running/pending tasks complete.
func (p *Pool) Stop() {
	close(p.tasks)
	p.wg.Wait()
}
