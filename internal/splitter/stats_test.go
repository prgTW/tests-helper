package splitter_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/rs/zerolog"

	"github.com/prgtw/tests-helper/internal/splitter"
	"github.com/prgtw/tests-helper/internal/worker"
)

func TestPercentileCalculator_Calculate(t *testing.T) {
	calc := splitter.NewPercentileCalculator()

	tests := []struct {
		name        string
		times       []float64
		percentiles []int
		want        map[int]float64
	}{
		{
			name:        "simple case",
			times:       []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			percentiles: []int{50, 100},
			want:        map[int]float64{50: 3.0, 100: 5.0},
		},
		{
			name:        "single value",
			times:       []float64{42.0},
			percentiles: []int{50, 75, 95, 99, 100},
			want:        map[int]float64{50: 42.0, 75: 42.0, 95: 42.0, 99: 42.0, 100: 42.0},
		},
		{
			name:        "two values",
			times:       []float64{10.0, 20.0},
			percentiles: []int{50, 100},
			want:        map[int]float64{50: 15.0, 100: 20.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.Calculate(tt.times, tt.percentiles)

			for p, wantVal := range tt.want {
				gotVal, exists := got[p]
				if !exists {
					t.Errorf("Percentile %d not found in result", p)
					continue
				}
				if !floatEqual(gotVal, wantVal, 0.001) {
					t.Errorf("P%d: got %.3f, want %.3f", p, gotVal, wantVal)
				}
			}
		})
	}

	t.Run("empty times", func(t *testing.T) {
		result := calc.Calculate([]float64{}, []int{50, 75, 95})
		if len(result) != 0 {
			t.Errorf("Expected empty result for empty times, got %v", result)
		}
	})
}

func TestStatsReporter_PrintSummary(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)
	reporter := splitter.NewStatsReporter(logger)

	stats := worker.Distribution{
		TotalTime: 100.0,
		AvgTime:   50.0,
		Workers: []worker.Stats{
			{
				Index:     0,
				Total:     50.0,
				TestCount: 5,
				MinTime:   2.0,
				MaxTime:   15.0,
				TestTimes: []float64{2.0, 5.0, 10.0, 15.0, 18.0},
			},
			{
				Index:     1,
				Total:     50.0,
				TestCount: 3,
				MinTime:   10.0,
				MaxTime:   20.0,
				TestTimes: []float64{10.0, 20.0, 20.0},
			},
		},
	}

	t.Run("without percentiles", func(t *testing.T) {
		buf.Reset()
		reporter.PrintSummary(stats, false)

		if buf.Len() == 0 {
			t.Error("PrintSummary produced no output")
		}

		// Check for key information in output
		checks := []string{
			"Distribution Summary",
			"100.000s", // total time
			"50.000s",  // avg time
			"Worker 0",
			"Worker 1",
		}

		for _, check := range checks {
			if !bytes.Contains(buf.Bytes(), []byte(check)) {
				t.Errorf("Output missing %q", check)
			}
		}

		// Should not contain percentile markers
		if bytes.Contains(buf.Bytes(), []byte("P50")) {
			t.Error("Output contains P50 when percentiles disabled")
		}
	})

	t.Run("with percentiles", func(t *testing.T) {
		buf.Reset()
		reporter.PrintSummary(stats, true)

		// Should contain percentile markers
		percentileMarkers := []string{"P50", "P75", "P95", "P99", "P100"}
		for _, marker := range percentileMarkers {
			if !bytes.Contains(buf.Bytes(), []byte(marker)) {
				t.Errorf("Output missing percentile marker %q", marker)
			}
		}
	})

	t.Run("empty workers", func(t *testing.T) {
		buf.Reset()
		emptyStats := worker.Distribution{
			TotalTime: 0,
			AvgTime:   0,
			Workers: []worker.Stats{
				{Index: 0, Total: 0, TestCount: 0},
			},
		}

		reporter.PrintSummary(emptyStats, false)

		if !bytes.Contains(buf.Bytes(), []byte("0 test files")) {
			t.Error("Output should mention '0 test files' for empty worker")
		}
	})
}

func TestStatsReporter_PrintWorkerDetails(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)
	reporter := splitter.NewStatsReporter(logger)

	// Create a simple allocator with test data using splitter
	s := splitter.NewSplitter(zerolog.New(os.Stderr).Level(zerolog.Disabled))
	input := "test1.go\ntest2.go\n"
	times := map[string]float64{"test1.go": 10.0, "test2.go": 5.0}
	testList, _ := s.ReadTests(bytes.NewReader([]byte(input)), times)
	allocator := s.Split(testList, 2)

	t.Run("valid worker", func(t *testing.T) {
		buf.Reset()
		reporter.PrintWorkerDetails(allocator, 0)

		if buf.Len() == 0 {
			t.Error("PrintWorkerDetails produced no output")
		}

		// Should contain rendering message
		if !bytes.Contains(buf.Bytes(), []byte("Rendering test files")) {
			t.Error("Output missing rendering message")
		}
	})

	t.Run("invalid worker index", func(t *testing.T) {
		buf.Reset()
		reporter.PrintWorkerDetails(allocator, 99)

		// Should log error about invalid index
		if !bytes.Contains(buf.Bytes(), []byte("Invalid worker index")) {
			t.Error("Output should mention invalid worker index")
		}
	})
}
