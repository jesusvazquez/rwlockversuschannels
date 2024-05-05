package main

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/grafana/dskit/services"
	"github.com/pkg/errors"
)

type tsdbWithWorkers struct {
	services.Service

	head           *head
	concurency     int
	workers        []*worker
	workerBalancer WorkerBalancer
	messages       chan message
	subServices    *services.Manager
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

	h.workerBalancer = newRoundRobinBalancer(h.workers...)

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
			w = h.workerBalancer.Next()
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

type WorkerBalancer interface {
	Next() *worker
}

type roundrobin struct {
	workers []*worker
	next    uint32
}

func newRoundRobinBalancer(workers ...*worker) *roundrobin {
	return &roundrobin{workers: workers}
}

func (r *roundrobin) Next() *worker {
	n := atomic.AddUint32(&r.next, 1)
	return r.workers[(int(n)-1)%len(r.workers)]
}
