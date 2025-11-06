package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/prgtw/tests-helper/internal/config"
	"github.com/prgtw/tests-helper/internal/junit"
	"github.com/prgtw/tests-helper/internal/splitter"
)

type splitOptions struct {
	statsFiles    []string
	indexFlag     int
	totalFlag     int
	noPercentiles bool
	debugFlag     bool
}

// newSplitCmd creates the split command.
func newSplitCmd(logger zerolog.Logger) *cobra.Command {
	opts := &splitOptions{}

	cmd := &cobra.Command{
		Use:   "split",
		Short: "Split tests across parallel workers",
		Long: `Split reads a list of test files from stdin and distributes them across
parallel workers based on historical execution times from JUnit XML reports.

The command outputs the test files assigned to the specified worker index.

Examples:
  # Split tests across 4 workers, get tests for worker 0
  cat test-list.txt | tests-helper split --stats "junit-*.xml" --index 0 --total 4

  # Use CircleCI environment variables
  cat test-list.txt | tests-helper split --stats "reports/*.xml"

  # Enable debug logging
  cat test-list.txt | tests-helper split --stats "*.xml" --debug --index 0 --total 2`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runSplit(logger, opts, os.Stdin, os.Stdout)
		},
	}

	cmd.Flags().
		StringSliceVar(&opts.statsFiles, "stats", []string{}, "Path(s) to JUnit XML stats files (supports glob patterns)")
	cmd.Flags().IntVar(&opts.indexFlag, "index", -1, "Worker index (overrides CIRCLE_NODE_INDEX)")
	cmd.Flags().IntVar(&opts.totalFlag, "total", -1, "Total number of workers (overrides CIRCLE_NODE_TOTAL)")
	cmd.Flags().BoolVar(&opts.noPercentiles, "no-percentiles", false, "Disable percentile statistics")
	cmd.Flags().BoolVar(&opts.debugFlag, "debug", false, "Enable debug logging")

	return cmd
}

func runSplit(logger zerolog.Logger, opts *splitOptions, stdin io.Reader, stdout io.Writer) error {
	// Configure logger level
	if opts.debugFlag {
		logger = logger.Level(zerolog.DebugLevel)
	} else {
		logger = logger.Level(zerolog.InfoLevel)
	}

	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get worker index and total
	total := cfg.GetNodeTotal(opts.totalFlag, 1)
	index := cfg.GetNodeIndex(opts.indexFlag, 0)

	// Validate index
	if index < 0 || index >= total {
		return fmt.Errorf("invalid node index: %d (must be between 0 and %d)", index, total-1)
	}

	logger.Info().
		Int("index", index).
		Int("total", total).
		Msg("Starting test split")

	// Parse JUnit XML files
	var times map[string]float64
	if len(opts.statsFiles) > 0 {
		parser := junit.NewParser(logger)
		times, err = parser.LoadFiles(opts.statsFiles)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to load stats files, continuing with defaults")
			times = make(map[string]float64)
		}
	} else {
		logger.Info().Msg("No stats files provided, using default test times")
		times = make(map[string]float64)
	}

	// Read tests from stdin
	testSplitter := splitter.NewSplitter(logger)
	tests, err := testSplitter.ReadTests(stdin, times)
	if err != nil {
		return fmt.Errorf("failed to read tests: %w", err)
	}

	// Split tests across workers
	allocator := testSplitter.Split(tests, total)

	// Print distribution summary using logger
	stats := allocator.GetStats()
	reporter := splitter.NewStatsReporter(logger)
	reporter.PrintSummary(stats, !opts.noPercentiles)

	// Print selected worker details using logger
	reporter.PrintWorkerDetails(allocator, index)

	// Print selected worker's tests to stdout
	worker := allocator.GetWorker(index)
	if worker == nil {
		return fmt.Errorf("failed to get worker %d", index)
	}

	for _, test := range worker.Tests {
		_, _ = fmt.Fprintln(stdout, test.Name)
	}

	logger.Info().
		Int("tests_assigned", len(worker.Tests)).
		Float64("total_time", worker.Total).
		Msg("Split completed successfully")

	return nil
}
