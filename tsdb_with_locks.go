package main

import "sync"

type tsdbWithLocks struct {
	head  *head
	locks map[int]*stripeLock
}

type stripeLock struct {
	sync.RWMutex          // 24 bytes
	_            [40]byte // 24+40 = 64 bytes. The extra padding makes sure locks go in different cache lines. Reduces lock contention and reader starvation
}

func newTsdbWithLocks(head *head, locks map[int]*stripeLock) *tsdbWithLocks {
	return &tsdbWithLocks{
		head:  head,
		locks: locks,
	}
}

func (h *tsdbWithLocks) Append(seriesID int, value float64) error {
	s, _ := h.head.memSeries.series[seriesID]
	l, _ := h.locks[s.id]
	l.Lock()
	s.points.Add(value)
	l.Unlock()

	return nil
}
