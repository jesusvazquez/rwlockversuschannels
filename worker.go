package main

import (
	"context"
	"sync"
	"time"

	"github.com/grafana/dskit/services"
)

type message struct {
	seriesID int
	value    float64
}

type worker struct {
	services.Service
	workerID int
	ctx      context.Context
	wg       sync.WaitGroup

	head *head

	now time.Time

	messages chan message
}

func newWorker(workerID int, head *head) *worker {
	w := &worker{
		workerID: workerID,
		messages: make(chan message, 1000),
		head:     head,
	}
	w.Service = services.NewBasicService(nil, w.running, w.stopping)
	return w
}

func (w *worker) dispatch(msg message) {
	w.messages <- msg
}

func (w *worker) running(ctx context.Context) error {
	w.ctx = ctx

	w.wg.Add(1)
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-w.messages:
			w.handleMessage(msg)
		}
	}
}

func (w *worker) stopping(_ error) error {
	close(w.messages)
	w.wg.Wait()

	return nil
}

func (w *worker) handleMessage(msg message) {
	s, _ := w.head.memSeries.series[msg.seriesID]
	s.points = append(s.points, msg.value)
}
