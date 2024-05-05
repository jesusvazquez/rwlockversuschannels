package main

import (
	"context"
	"time"

	"github.com/grafana/dskit/services"
	"github.com/pkg/errors"
)

type tsdbWithWorkers struct {
	services.Service

	head        *head
	concurency  int
	workers     []*worker
	messages    chan message
	subServices *services.Manager
}

func newTsdbWithWorkers(head *head, concurrency int) *tsdbWithWorkers {
	h := &tsdbWithWorkers{
		head:       head,
		concurency: concurrency,
		messages:   make(chan message, 10000),
	}

	var subservices []services.Service
	h.workers = make([]*worker, concurrency)
	for workerID := 0; workerID < concurrency; workerID++ {
		worker := newWorker(workerID, head)
		h.workers[workerID] = worker
		subservices = append(subservices, worker)
	}

	h.subServices, _ = services.NewManager(subservices...)

	h.Service = services.NewBasicService(h.starting, h.running, h.stopping)

	return h
}

func (h *tsdbWithWorkers) Append(seriesID int, value float64) error {
	h.messages <- message{seriesID: seriesID, value: value}
	return nil
}

func (h *tsdbWithWorkers) starting(_ context.Context) error {
	if err := services.StartManagerAndAwaitHealthy(context.Background(), h.subServices); err != nil {
		return errors.Wrap(err, "unable to start tsdb subservices")
	}
	return nil
}

func (h *tsdbWithWorkers) running(svcCtx context.Context) error {
	var w *worker
	for svcCtx.Err() == nil {
		select {
		case msg := <-h.messages:
			if msg.seriesID <= 5000 {
				w = h.workers[0]
			} else {
				w = h.workers[1]
			}
			w.dispatch(msg)
		}
	}

	return nil
}

func (h *tsdbWithWorkers) stopping(_ error) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	return errors.Wrap(services.StopManagerAndAwaitStopped(ctx, h.subServices), "unable to stop tsdb subservices")
}
