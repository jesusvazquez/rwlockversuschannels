package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type memSeries struct {
	series map[int]*serie

	locks map[int]stripeLock // 24 bytes
}

type stripeLock struct {
	sync.RWMutex
	_ [40]byte
}

type serie struct {
	id     int
	points []float64
}

func main() {
	// initialize series with 10k series
	series := make([]memSeries, 10000)
	for i := range series {
		series[i] = memSeries{
			series: make(map[int]*serie),
			locks:  make(map[int]stripeLock),
		}
		for j := 0; j < 1000; j++ {
			series[i].series[j] = &serie{
				id: j,
			}
		}
	}

	runtime.GOMAXPROCS(3)
	// Create 2 goroutines where the first one writes to the first 500 series and the second one writes to the last 500 series
	var wg sync.WaitGroup
	wg.Add(3)
	var writesLoop1, writesLoop2 int
	go func() {
		defer wg.Done()

		for {
			writesLoop1 += 1
			addPointUsingLocks(series[:500])
			time.Sleep(1 * time.Millisecond)
		}
	}()
	go func() {
		defer wg.Done()

		for {
			writesLoop2 += 1
			addPointUsingLocks(series[500:])
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

	// write one point to all series
	addPointUsingLocks(series) // i = 0
}

func addPointUsingLocks(series []memSeries) {
	for i := range series {
		for _, s := range series[i].series {
			lock, ok := series[i].locks[s.id]
			if !ok {
				series[i].locks[s.id] = stripeLock{}
			}
			lock.Lock()
			if len(s.points) == 1000 {
				s.points = s.points[:0] // Reset the slice to avoid memory leak
			}
			s.points = append(s.points, 1.0)
			lock.Unlock()
		}
	}
}
