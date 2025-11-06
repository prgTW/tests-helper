package splitter

import (
	"fmt"
	"sort"

	"github.com/rs/zerolog"

	"github.com/prgtw/tests-helper/internal/worker"
)

// StatsReporter handles printing of distribution statistics.
type StatsReporter struct {
	logger zerolog.Logger
}

// NewStatsReporter creates a new statistics reporter.
func NewStatsReporter(logger zerolog.Logger) *StatsReporter {
	return &StatsReporter{logger: logger}
}

// PrintSummary prints the overall distribution summary.
func (r *StatsReporter) PrintSummary(stats worker.Distribution, showPercentiles bool) {
	r.logger.Info().Msg("=== Distribution Summary ===")
	r.logger.Info().
		Float64("total_time", stats.TotalTime).
		Float64("avg_per_bucket", stats.AvgTime).
		Msgf("Total time: %.3fs, Avg per bucket: %.3fs", stats.TotalTime, stats.AvgTime)

	for _, ws := range stats.Workers {
		if ws.TestCount == 0 {
			r.logger.Info().
				Int("worker", ws.Index).
				Msgf("Worker %d: 0 test files", ws.Index)
			continue
		}

		r.logger.Info().
			Int("worker", ws.Index).
			Float64("total_time", ws.Total).
			Int("test_count", ws.TestCount).
			Float64("min_time", ws.MinTime).
			Float64("max_time", ws.MaxTime).
			Msgf("Worker %d: %.3fs (%d test files, min %.3fs, max %.3fs)",
				ws.Index, ws.Total, ws.TestCount, ws.MinTime, ws.MaxTime)

		if showPercentiles && len(ws.TestTimes) > 0 {
			r.printWorkerPercentiles(ws.TestTimes)
		}
	}
}

// printWorkerPercentiles prints percentile statistics for a worker.
func (r *StatsReporter) printWorkerPercentiles(times []float64) {
	// Sort times for percentile calculation
	sorted := make([]float64, len(times))
	copy(sorted, times)
	sort.Float64s(sorted)

	calc := NewPercentileCalculator()
	percentiles := []int{50, 75, 95, 99, 100}
	results := calc.Calculate(sorted, percentiles)

	for _, p := range percentiles {
		label := fmt.Sprintf("P%-3d", p)
		r.logger.Info().
			Int("percentile", p).
			Float64("value", results[p]).
			Msgf("%4s = %.3fs", label, results[p])
	}
}

// PrintWorkerDetails prints detailed information about a specific worker.
func (r *StatsReporter) PrintWorkerDetails(allocator *worker.Allocator, index int) {
	w := allocator.GetWorker(index)
	if w == nil {
		r.logger.Error().
			Int("worker_index", index).
			Msg("Invalid worker index")
		return
	}

	r.logger.Info().
		Int("worker", index).
		Float64("total_time", w.Total).
		Int("test_count", len(w.Tests)).
		Msg("Rendering test files")
}

// PercentileCalculator calculates percentiles for test time distributions.
type PercentileCalculator struct{}

// NewPercentileCalculator creates a new percentile calculator.
func NewPercentileCalculator() *PercentileCalculator {
	return &PercentileCalculator{}
}

// Calculate calculates percentiles for a set of test times.
func (pc *PercentileCalculator) Calculate(times []float64, percentiles []int) map[int]float64 {
	if len(times) == 0 {
		return make(map[int]float64)
	}

	// Sort times for percentile calculation
	sorted := make([]float64, len(times))
	copy(sorted, times)
	sort.Float64s(sorted)

	results := make(map[int]float64)
	const percentageDivisor = 100.0
	for _, p := range percentiles {
		pos := float64(p) / percentageDivisor * float64(len(sorted)-1)
		results[p] = interpolate(sorted, pos)
	}

	return results
}

// interpolate performs linear interpolation for percentile calculation.
func interpolate(sorted []float64, pos float64) float64 {
	i := int(pos)
	if i >= len(sorted)-1 {
		return sorted[len(sorted)-1]
	}
	frac := pos - float64(i)
	return sorted[i]*(1-frac) + sorted[i+1]*frac
}

// PrintPercentiles prints percentile statistics.
func (r *StatsReporter) PrintPercentiles(times []float64) {
	if len(times) == 0 {
		return
	}

	calc := NewPercentileCalculator()
	percentiles := []int{50, 75, 95, 99, 100}
	results := calc.Calculate(times, percentiles)

	for _, p := range percentiles {
		label := fmt.Sprintf("P%-3d", p)
		r.logger.Info().
			Int("percentile", p).
			Float64("value", results[p]).
			Msgf("%4s = %.3fs", label, results[p])
	}
}
