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

func (h *tsdbWithLocks) Append(seriesID int, value float64) error {
	s, _ := h.head.memSeries.series[seriesID]
	l, _ := h.locks[s.id]
	l.Lock()
	s.points = append(s.points, value)
	l.Unlock()

	return nil
}

func (h *tsdbWithLocks) appendSamplesToSeriesBetwenRange(start, end int, value float64) {
	for i := start; i < end; i++ {
		_ = h.Append(i, value)
	}
}
