package worker_test

import (
	"testing"

	"github.com/tomaszwojcik/tests-helper/internal/junit"
	"github.com/tomaszwojcik/tests-helper/internal/worker"
)

func TestAllocator_Distribute(t *testing.T) {
	tests := []struct {
		name       string
		tests      []junit.Test
		numWorkers int
		want       []float64 // expected total time per worker
	}{
		{
			name: "equal distribution",
			tests: []junit.Test{
				{Name: "test1", Time: 10.0},
				{Name: "test2", Time: 10.0},
				{Name: "test3", Time: 10.0},
				{Name: "test4", Time: 10.0},
			},
			numWorkers: 2,
			want:       []float64{20.0, 20.0},
		},
		{
			name: "unequal tests balanced",
			tests: []junit.Test{
				{Name: "test1", Time: 5.0},
				{Name: "test2", Time: 3.0},
				{Name: "test3", Time: 2.0},
			},
			numWorkers: 2,
			want:       []float64{5.0, 5.0},
		},
		{
			name: "greedy algorithm test",
			tests: []junit.Test{
				{Name: "large", Time: 100.0},
				{Name: "medium", Time: 50.0},
				{Name: "small1", Time: 25.0},
				{Name: "small2", Time: 25.0},
			},
			numWorkers: 2,
			want:       []float64{100.0, 100.0},
		},
		{
			name: "single test",
			tests: []junit.Test{
				{Name: "only", Time: 42.0},
			},
			numWorkers: 3,
			want:       []float64{42.0, 0.0, 0.0},
		},
		{
			name: "more workers than tests",
			tests: []junit.Test{
				{Name: "test1", Time: 10.0},
				{Name: "test2", Time: 20.0},
			},
			numWorkers: 5,
			want:       []float64{10.0, 20.0, 0.0, 0.0, 0.0}, // greedy assigns to worker with min load
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allocator := worker.NewAllocator(tt.numWorkers)
			allocator.Distribute(tt.tests)

			for i := range tt.numWorkers {
				w := allocator.GetWorker(i)
				if w == nil {
					t.Fatalf("GetWorker(%d) returned nil", i)
				}
				if w.Total != tt.want[i] {
					t.Errorf("Worker %d: got total=%.1f, want %.1f", i, w.Total, tt.want[i])
				}
			}
		})
	}
}

func TestAllocator_GetWorker(t *testing.T) {
	tests := []junit.Test{
		{Name: "test1", Time: 10.0},
		{Name: "test2", Time: 20.0},
	}

	allocator := worker.NewAllocator(2)
	allocator.Distribute(tests)

	t.Run("valid index", func(t *testing.T) {
		w := allocator.GetWorker(0)
		if w == nil {
			t.Fatal("GetWorker(0) returned nil")
		}
		if len(w.Tests) == 0 {
			t.Error("Worker 0 has no tests")
		}
	})

	t.Run("negative index", func(t *testing.T) {
		w := allocator.GetWorker(-1)
		if w != nil {
			t.Error("GetWorker(-1) should return nil")
		}
	})

	t.Run("index out of bounds", func(t *testing.T) {
		w := allocator.GetWorker(10)
		if w != nil {
			t.Error("GetWorker(10) should return nil")
		}
	})
}

func TestAllocator_GetStats(t *testing.T) {
	tests := []junit.Test{
		{Name: "test1", Time: 10.0},
		{Name: "test2", Time: 5.0},
		{Name: "test3", Time: 3.0},
		{Name: "test4", Time: 2.0},
	}

	allocator := worker.NewAllocator(2)
	allocator.Distribute(tests)

	stats := allocator.GetStats()

	t.Run("total time", func(t *testing.T) {
		expectedTotal := 20.0
		if stats.TotalTime != expectedTotal {
			t.Errorf("TotalTime: got %.1f, want %.1f", stats.TotalTime, expectedTotal)
		}
	})

	t.Run("average time", func(t *testing.T) {
		expectedAvg := 10.0
		if stats.AvgTime != expectedAvg {
			t.Errorf("AvgTime: got %.1f, want %.1f", stats.AvgTime, expectedAvg)
		}
	})

	t.Run("worker count", func(t *testing.T) {
		if len(stats.Workers) != 2 {
			t.Errorf("Workers count: got %d, want 2", len(stats.Workers))
		}
	})

	t.Run("worker stats", func(t *testing.T) {
		for i, ws := range stats.Workers {
			if ws.Index != i {
				t.Errorf("Worker %d: Index got %d, want %d", i, ws.Index, i)
			}
			if ws.TestCount == 0 {
				t.Errorf("Worker %d: TestCount is 0", i)
			}
			if ws.Total == 0 {
				t.Errorf("Worker %d: Total is 0", i)
			}
			if len(ws.TestTimes) != ws.TestCount {
				t.Errorf("Worker %d: TestTimes length %d != TestCount %d",
					i, len(ws.TestTimes), ws.TestCount)
			}
		}
	})

	t.Run("min and max times", func(t *testing.T) {
		for i, ws := range stats.Workers {
			if ws.TestCount > 0 {
				if ws.MinTime <= 0 {
					t.Errorf("Worker %d: MinTime should be positive, got %.1f", i, ws.MinTime)
				}
				if ws.MaxTime <= 0 {
					t.Errorf("Worker %d: MaxTime should be positive, got %.1f", i, ws.MaxTime)
				}
				if ws.MinTime > ws.MaxTime {
					t.Errorf("Worker %d: MinTime %.1f > MaxTime %.1f", i, ws.MinTime, ws.MaxTime)
				}
			}
		}
	})
}

func TestAllocator_EmptyTests(t *testing.T) {
	allocator := worker.NewAllocator(3)
	allocator.Distribute([]junit.Test{})

	stats := allocator.GetStats()

	if stats.TotalTime != 0 {
		t.Errorf("TotalTime: got %.1f, want 0", stats.TotalTime)
	}

	for i, ws := range stats.Workers {
		if ws.TestCount != 0 {
			t.Errorf("Worker %d: TestCount got %d, want 0", i, ws.TestCount)
		}
		if ws.Total != 0 {
			t.Errorf("Worker %d: Total got %.1f, want 0", i, ws.Total)
		}
		if ws.MinTime != 0 {
			t.Errorf("Worker %d: MinTime got %.1f, want 0", i, ws.MinTime)
		}
	}
}

func TestAllocator_BalancedDistribution(t *testing.T) {
	// Test that the greedy algorithm produces reasonable balance
	tests := []junit.Test{
		{Name: "t1", Time: 100.0},
		{Name: "t2", Time: 90.0},
		{Name: "t3", Time: 80.0},
		{Name: "t4", Time: 70.0},
		{Name: "t5", Time: 60.0},
		{Name: "t6", Time: 50.0},
		{Name: "t7", Time: 40.0},
		{Name: "t8", Time: 30.0},
		{Name: "t9", Time: 20.0},
		{Name: "t10", Time: 10.0},
	}

	allocator := worker.NewAllocator(3)
	allocator.Distribute(tests)

	stats := allocator.GetStats()

	// Calculate the maximum difference between workers
	minWorkerTime := stats.Workers[0].Total
	maxWorkerTime := stats.Workers[0].Total

	for _, ws := range stats.Workers {
		if ws.Total < minWorkerTime {
			minWorkerTime = ws.Total
		}
		if ws.Total > maxWorkerTime {
			maxWorkerTime = ws.Total
		}
	}

	difference := maxWorkerTime - minWorkerTime
	tolerance := stats.AvgTime * 0.3 // Allow 30% variance

	if difference > tolerance {
		t.Errorf("Distribution is not balanced: max-min=%.1f exceeds tolerance %.1f",
			difference, tolerance)
		for i, ws := range stats.Workers {
			t.Logf("Worker %d: %.1f (%d tests)", i, ws.Total, ws.TestCount)
		}
	}
}
