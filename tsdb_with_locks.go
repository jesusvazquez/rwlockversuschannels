package main

import (
	"sync"
)

type tsdbWithLocks struct {
	head       *head
	locks      map[int]*stripeLock
	locksMutex sync.Mutex
}

type stripeLock struct {
	sync.RWMutex          // 24 bytes
	_            [40]byte // 24+40 = 64 bytes. The extra padding makes sure locks go in different cache lines. Reduces lock contention and reader starvation
}

func newTsdbWithLocks(head *head) *tsdbWithLocks {
	// Generating locks before appending so we don't need a lock to add new locks to the map
	locks := make(map[int]*stripeLock)
	for i := 0; i < len(head.memSeries.series); i++ {
		locks[i] = &stripeLock{}
	}
	return &tsdbWithLocks{
		head:  head,
		locks: locks,
	}
}

func (h *tsdbWithLocks) Append(seriesID int, value float64) error {
	s, _ := h.head.memSeries.series[seriesID]
	l, ok := h.locks[s.id]
	if !ok {
		h.locksMutex.Lock()
		l = &stripeLock{}
		h.locks[s.id] = l
		h.locksMutex.Unlock()
	}
	l.Lock()
	s.points.Add(value)
	l.Unlock()

	return nil
}
