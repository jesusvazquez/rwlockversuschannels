package main

type head struct {
	memSeries *memSeries
}

type Appender interface {
	Append(seriesID int, value float64) error
}

type memSeries struct {
	series map[int]*serie
}

type serie struct {
	id     int
	points *pointsRepository
}

type pointsRepository struct {
	points []float64
	count  uint32
}

func newPointsRepository() *pointsRepository {
	return &pointsRepository{
		points: make([]float64, 1000),
	}
}

func (p *pointsRepository) Add(point float64) {
	if p.count == 1000 { // just keep track of 1k and start over to avoid this growing over time
		p.count = 0
	}
	p.points[p.count] = point
	p.count++
}
