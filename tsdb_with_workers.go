package main

type tsdbWithWorkers struct {
	head       *head
	concurency int
	workers    []*worker
}

func newTsdbWithWorkers(head *head, concurrency int) *tsdbWithWorkers {
	h := &tsdbWithWorkers{
		head:       head,
		concurency: concurrency,
	}

	h.workers = make([]*worker, concurrency)
	for workerID := 0; workerID < concurrency; workerID++ {
		worker := newWorker(workerID, head)
		h.workers[workerID] = worker
	}

	return h
}

func (h *tsdbWithWorkers) Append(seriesID int, value float64) error {
	w := h.workers[0] // TODO choose right worker
	w.messages <- message{seriesID: seriesID, value: value}
	return nil
}
