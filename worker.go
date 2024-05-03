package main

import (
	"context"
	"sync"
	"time"
)

type message struct {
	seriesID int
	value    float64
}

type worker struct {
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
		messages: make(chan message, 1),
		head:     head,
	}
	return w
}

func (w *worker) dispatch(msg message) {
	w.messages <- msg
}

func (w *worker) running(ctx context.Context) error {
	w.ctx = ctx

	w.wg.Add(1)
	defer w.wg.Done()

	// optimizationto not have to call time.Now() as often
	wallclockTicker := time.NewTicker(time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil
		case w.now = <-wallclockTicker.C:
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
