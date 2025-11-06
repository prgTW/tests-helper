package splitter

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/rs/zerolog"

	"github.com/prgtw/tests-helper/internal/junit"
	"github.com/prgtw/tests-helper/internal/worker"
)

const (
	DefaultTestTime = 1.0 // Default time for tests without historical data
)

// Splitter handles the test splitting logic.
type Splitter struct {
	logger zerolog.Logger
}

// NewSplitter creates a new test splitter.
func NewSplitter(logger zerolog.Logger) *Splitter {
	return &Splitter{logger: logger}
}

// ReadTests reads test names from a reader and assigns times based on historical data.
func (s *Splitter) ReadTests(r io.Reader, times map[string]float64) ([]junit.Test, error) {
	var tests []junit.Test
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name == "" {
			continue
		}

		time := times[name]
		if time == 0 {
			time = DefaultTestTime
			s.logger.Debug().
				Str("test", name).
				Float64("time", time).
				Msg("No historical data, using default time")
		}

		tests = append(tests, junit.Test{
			Name: name,
			Time: time,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading tests: %w", err)
	}

	if len(tests) == 0 {
		return nil, errors.New("no tests provided")
	}

	s.logger.Info().
		Int("count", len(tests)).
		Msg("Read tests from input")

	return tests, nil
}

// SortTests sorts tests by descending execution time.
func (s *Splitter) SortTests(tests []junit.Test) {
	sort.Slice(tests, func(i, j int) bool {
		return tests[i].Time > tests[j].Time
	})
	s.logger.Debug().Msg("Sorted tests by descending time")
}

// Split performs the complete test splitting operation.
func (s *Splitter) Split(tests []junit.Test, numWorkers int) *worker.Allocator {
	// Sort tests by descending time for optimal distribution
	s.SortTests(tests)

	// Create allocator and distribute tests
	allocator := worker.NewAllocator(numWorkers)
	allocator.Distribute(tests)

	s.logger.Info().
		Int("workers", numWorkers).
		Int("tests", len(tests)).
		Msg("Split tests across workers")

	return allocator
}
