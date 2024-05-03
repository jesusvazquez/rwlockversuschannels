package main

type headWithLocks struct {
	head  *head
	locks map[int]*stripeLock
}

func (h *headWithLocks) Append(seriesID int, value float64) error {
	s, _ := h.head.memSeries.series[seriesID]
	l, ok := h.locks[s.id]
	if !ok {
		h.locks[s.id] = &stripeLock{}
	}
	l.Lock()
	s.points = append(s.points, value)
	l.Unlock()

	return nil
}
