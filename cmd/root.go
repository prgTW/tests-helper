package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	version = "snapshot"
	commit  = "<commit-unknown>"              //nolint:gochecknoglobals // replaced by GoReleaser
	date    = time.Now().Format(time.RFC3339) //nolint:gochecknoglobals // replaced by GoReleaser
)

// newRootCmd creates the root command.
func newRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tests-helper",
		Short: "A tool for splitting test suites across parallel workers",
		Long: fmt.Sprintf(`tests-helper is a CLI tool that distributes test files across parallel workers
based on historical execution times from JUnit XML reports.

It uses a greedy algorithm to balance test execution time across workers,
helping to optimize parallel test execution in CI/CD environments.

Version: %s
Commit:  %s
Built:   %s`, version, commit, date),
		Version: version,
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Initialize logger with console output
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Logger()

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newSplitCmd(logger))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
