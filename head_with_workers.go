package main

type headWithWorkers struct {
	head       *head
	concurency int
	workers    []*worker
}

func newHeadWithWorkers(head *head, concurrency int) *headWithWorkers {
	h := &headWithWorkers{
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
