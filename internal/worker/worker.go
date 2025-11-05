package worker

import (
	"math"

	"github.com/tomaszwojcik/tests-helper/internal/junit"
)

// Worker represents a worker with assigned tests.
type Worker struct {
	Tests []junit.Test
	Total float64
}

// Allocator handles distribution of tests across workers.
type Allocator struct {
	workers []Worker
}

// NewAllocator creates a new worker allocator.
func NewAllocator(numWorkers int) *Allocator {
	return &Allocator{
		workers: make([]Worker, numWorkers),
	}
}

// Distribute distributes tests across workers using a greedy algorithm.
// Tests should be sorted by time in descending order for best results.
func (a *Allocator) Distribute(tests []junit.Test) {
	for _, test := range tests {
		// Find worker with minimum total time
		minIdx := 0
		for i := 1; i < len(a.workers); i++ {
			if a.workers[i].Total < a.workers[minIdx].Total {
				minIdx = i
			}
		}

		// Assign test to worker with minimum load
		a.workers[minIdx].Tests = append(a.workers[minIdx].Tests, test)
		a.workers[minIdx].Total += test.Time
	}
}

// GetWorker returns the worker at the specified index.
func (a *Allocator) GetWorker(index int) *Worker {
	if index < 0 || index >= len(a.workers) {
		return nil
	}
	return &a.workers[index]
}

// GetWorkers returns all workers.
func (a *Allocator) GetWorkers() []Worker {
	return a.workers
}

// Distribution returns statistics about worker distribution.
type Distribution struct {
	TotalTime float64
	AvgTime   float64
	Workers   []Stats
}

// Stats represents statistics for a single worker.
type Stats struct {
	Index     int
	Total     float64
	TestCount int
	MinTime   float64
	MaxTime   float64
	TestTimes []float64
}

// GetStats calculates distribution statistics.
func (a *Allocator) GetStats() Distribution {
	var totalTime float64
	workerStats := make([]Stats, len(a.workers))

	for i, w := range a.workers {
		totalTime += w.Total

		minTime := math.MaxFloat64
		maxTime := 0.0
		testTimes := make([]float64, len(w.Tests))

		for j, t := range w.Tests {
			testTimes[j] = t.Time
			if t.Time < minTime {
				minTime = t.Time
			}
			if t.Time > maxTime {
				maxTime = t.Time
			}
		}

		// Handle case where worker has no tests
		if len(w.Tests) == 0 {
			minTime = 0
		}

		workerStats[i] = Stats{
			Index:     i,
			Total:     w.Total,
			TestCount: len(w.Tests),
			MinTime:   minTime,
			MaxTime:   maxTime,
			TestTimes: testTimes,
		}
	}

	return Distribution{
		TotalTime: totalTime,
		AvgTime:   totalTime / float64(len(a.workers)),
		Workers:   workerStats,
	}
}
