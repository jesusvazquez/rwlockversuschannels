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
	}

	locks := make(map[int]*stripeLock) // initializing locks here to avoi concurrent map writes
	for i := 0; i < 10000; i++ {
		memSeries.series[i] = &serie{
			id: i,
		}
		locks[i] = &stripeLock{}
	}

	h := head{
		memSeries: &memSeries,
	}

	t1 := tsdbWithLocks{
		head:  &h,
		locks: locks,
	}

	var wg sync.WaitGroup
	wg.Add(3)
	var writesLoop1, writesLoop2 int
	go func() {
		defer wg.Done()

		for {
			writesLoop1 += 1
			t1.appendSamplesToSeriesBetwenRange(0, 5000, float64(writesLoop1))
		}
	}()
	go func() {
		defer wg.Done()

		for {
			writesLoop2 += 1
			t1.appendSamplesToSeriesBetwenRange(5001, 9999, float64(writesLoop1))
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
