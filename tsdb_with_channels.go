package main

type tsdbWithChannels struct {
	head     *head
	messages chan message
}

func newTsdbWithChannels(head *head) *tsdbWithChannels {
	h := &tsdbWithChannels{
		messages: make(chan message, 1000),
	}

	for k, v := range head.memSeries.series {
		go work(k, v, h.messages)
	}

	return h
}

func work(_ int, serie *serie, ch chan message) {
	for msg := range ch {
		serie.points.Add(msg.value)
	}
}

func (h *tsdbWithChannels) Append(seriesID int, value float64) error {
	h.messages <- message{seriesID: seriesID, value: value}
	return nil
}
