package main

import (
	"fmt"
	"sync"
	"time"
)

type head struct {
	memSeries *memSeries
}

type memSeries struct {
	series map[int]*serie

	locks map[int]*stripeLock // 24 bytes
}

type stripeLock struct {
	sync.RWMutex          // 24 bytes
	_            [40]byte // 24+40 = 64 bytes. The extra padding makes sure locks go in different cache lines. Reduces lock contention and reader starvation
}

type serie struct {
	id     int
	points []float64
}

type Appender interface {
	Append(seriesID int, value float64) error
}

func main() {
	memSeries := memSeries{
		series: make(map[int]*serie),
		locks:  make(map[int]*stripeLock),
	}
	for i := 0; i < 10000; i++ {
		memSeries.series[i] = &serie{
			id: i,
		}
		memSeries.locks[i] = &stripeLock{}
	}

	var wg sync.WaitGroup
	wg.Add(3)
	var writesLoop1, writesLoop2 int
	go func() {
		defer wg.Done()

		for {
			writesLoop1 += 1
			memSeries.addPointUsingLocksBetweenIntervals(0, 5000)
		}
	}()
	go func() {
		defer wg.Done()

		for {
			writesLoop2 += 1
			memSeries.addPointUsingLocksBetweenIntervals(5001, 9999)
		}
	}()

	go func() {
		defer wg.Done()
		for {
			fmt.Printf("Writes loop 1: %d, Writes Loop 2: %d\n", writesLoop1, writesLoop2)
			time.Sleep(5 * time.Second)
		}
	}()
	wg.Wait()
}

// Create a function for memSeries that iterates its series and adds a point to each one of them
// using locks
func (m memSeries) addPointUsingLocksBetweenIntervals(start, end int) {
	for i := start; i < end; i++ {
		s, _ := m.series[i]
		lock, ok := m.locks[s.id]
		if !ok {
			m.locks[s.id] = &stripeLock{}
		}
		lock.Lock()
		if len(s.points) == 1000 {
			s.points = s.points[:0] // Reset the slice to avoid memory leak
		}
		s.points = append(s.points, 1.0)
		lock.Unlock()
	}
}
