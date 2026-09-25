// Package worker provides the background job framework.
package worker

import (
	"context"
	"time"
)

// Worker runs job handlers on a fixed interval.
type Worker interface {
	Run(handlers ...Handler)
	RunAsync(handlers ...Handler)
	Close()
}

// Handler is a single background job.
type Handler func(ctx context.Context)

type worker struct {
	ticker *time.Ticker
	quit   chan struct{}
}

// NewWorker returns a Worker that fires every interval.
func NewWorker(interval time.Duration) Worker {
	return &worker{
		ticker: time.NewTicker(interval),
		quit:   make(chan struct{}),
	}
}

func (w *worker) Run(handlers ...Handler) {
	for {
		select {
		case <-w.ticker.C:
			for _, handler := range handlers {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
				handler(ctx)
				cancel()
			}
		case <-w.quit:
			return
		}
	}
}

func (w *worker) RunAsync(handlers ...Handler) {
	go w.Run(handlers...)
}

func (w *worker) Close() {
	w.ticker.Stop()
	close(w.quit)
}
