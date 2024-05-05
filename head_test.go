package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testHead(seriesCount int) head {
	memSeries := memSeries{
		series: make(map[int]*serie),
	}
	for i := 0; i < seriesCount; i++ {
		memSeries.series[i] = &serie{
			id:     i,
			points: newPointsRepository(),
		}
	}
	h := head{
		memSeries: &memSeries,
	}
	return h
}

func TestTsdbWithLocks_Append(t *testing.T) {
	seriesCount := 10000
	head := testHead(seriesCount)
	tsdb := newTsdbWithLocks(&head)

	for i := 0; i < seriesCount; i++ {
		err := tsdb.Append(i, 1)
		assert.NoError(t, err)
	}
}

func TestTsdbWithWorkers_Append(t *testing.T) {
	seriesCount := 10000
	head := testHead(seriesCount)
	tsdb := newTsdbWithWorkers(&head, 2)
	for i := 0; i < seriesCount; i++ {
		err := tsdb.Append(i, 1)
		assert.NoError(t, err)
	}
}

var inputTable = []struct {
	series int
}{
	{series: 10},
	{series: 100},
	{series: 1000},
	{series: 10000},
}

func BenchmarkTsdbWithLocks_Append(b *testing.B) {
	for _, v := range inputTable {
		b.Run(fmt.Sprintf("input_series_%d", v.series), func(b *testing.B) {
			head := testHead(v.series)
			tsdb := newTsdbWithLocks(&head)
			for i := 0; i < v.series; i++ {
				_ = tsdb.Append(i, float64(i))
			}
		})
	}
}

var workerConcurrency = []struct {
	concurrency int
}{
	{concurrency: 1},
	{concurrency: 2},
	{concurrency: 5},
	{concurrency: 10},
}

func BenchmarkTsdbWithWorkers_Append(b *testing.B) {
	for _, v := range inputTable {
		for _, c := range workerConcurrency {
			b.Run(fmt.Sprintf("input_series_%d_concurrency_%d", v.series, c.concurrency), func(b *testing.B) {
				head := testHead(v.series)
				tsdb := newTsdbWithWorkers(&head, c.concurrency)
				for i := 0; i < v.series; i++ {
					_ = tsdb.Append(i, float64(i))
				}
			})
		}
	}
}
