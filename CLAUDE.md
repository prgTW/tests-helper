# Test Splitter - AI Agent Context

This document provides context for AI agents working with the tests-helper codebase.

## Project Overview

**tests-helper** is a CLI tool that distributes test files across parallel workers based on historical execution times from JUnit XML reports. It uses a greedy algorithm to balance test execution time across workers, optimizing parallel test execution in CI/CD environments.

## Architecture

### Technology Stack
- **Language**: Go 1.x
- **CLI Framework**: spf13/cobra (command-line interface)
- **Logging**: rs/zerolog (structured JSON logging)
- **Configuration**: caarlos0/env (environment variable parsing)

### Directory Structure

```
tests-helper/
├── main.go                    # Application entry point
├── cmd/                       # Cobra CLI commands
│   ├── root.go               # Root command (empty, shows help)
│   └── split.go              # Split subcommand (main logic)
├── internal/                 # Private application code
│   ├── config/
│   │   └── config.go         # Env var configuration (CircleCI support)
│   ├── junit/
│   │   ├── types.go          # JUnit XML data structures
│   │   └── parser.go         # JUnit XML parsing logic
│   ├── splitter/
│   │   ├── splitter.go       # Test splitting orchestration
│   │   └── stats.go          # Statistics and percentile calculation
│   └── worker/
│       └── worker.go         # Worker allocation and distribution
├── old.go                    # Original implementation (reference)
├── go.mod                    # Go module definition
└── go.sum                    # Dependency checksums
```

## Core Concepts

### 1. Test Distribution Algorithm
- **Algorithm**: Greedy bin packing
- **Strategy**: Always assign the next test to the worker with the minimum total time
- **Input**: List of tests sorted by execution time (descending)
- **Output**: Balanced distribution across N workers

### 2. Historical Time Data
- **Source**: JUnit XML reports from previous test runs
- **Format**: `<testsuite file="test/path.go" time="12.345">`
- **Fallback**: Tests without historical data get a default time of 1.0 seconds

### 3. Worker Assignment
- Workers are identified by index (0 to N-1)
- The tool outputs only the tests assigned to the specified worker index
- Statistics are printed to stderr, test list to stdout

## Key Components

### Configuration (`internal/config`)
- Parses `CIRCLE_NODE_INDEX` and `CIRCLE_NODE_TOTAL` environment variables for compatibility with CircleCI
- Supports CLI flag overrides
- Used for test splitting

### JUnit Parser (`internal/junit`)
- Parses JUnit XML files with nested `<testsuite>` elements
- Supports glob patterns for multiple input files
- Accumulates test times across multiple reports
- Handles locale-specific decimal separators (comma vs dot)

### Worker Allocator (`internal/worker`)
- Implements greedy distribution algorithm
- Calculates distribution statistics (min, max, avg, percentiles)
- Maintains worker load balance

### Splitter (`internal/splitter`)
- Orchestrates the splitting workflow
- Reads test names from stdin
- Sorts tests by execution time
- Coordinates worker allocation
- Generates statistics reports

## Usage Examples

```bash
# Basic usage with 4 workers, get tests for worker 0
cat test-list.txt | tests-helper split --stats "junit-*.xml" --index 0 --total 4

# Use CircleCI environment variables (CIRCLE_NODE_INDEX, CIRCLE_NODE_TOTAL)
cat test-list.txt | tests-helper split --stats "reports/*.xml"

# Enable debug logging
cat test-list.txt | tests-helper split --stats "*.xml" --debug --index 0 --total 2

# Disable percentile statistics
cat test-list.txt | tests-helper split --stats "*.xml" --no-percentiles

# Show help
tests-helper split --help
```

## Input/Output

### Input
- **stdin**: Newline-separated list of test file paths
- **--stats flag**: Glob pattern(s) for JUnit XML files

### Output
- **stdout**: Test files assigned to the selected worker (one per line)
- **stderr**: Structured logs and statistics summary

### Statistics Output (stderr - structured logging)
All statistics are logged using zerolog with structured fields:
```
7:10PM INF Starting test split index=0 total=4
7:10PM INF No stats files provided, using default test times
7:10PM INF Read tests from input count=120
7:10PM INF Split tests across workers tests=120 workers=4
7:10PM INF === Distribution Summary ===
7:10PM INF Total time: 120.456s, Avg per bucket: 30.114s avg_per_bucket=30.114 total_time=120.456
7:10PM INF Worker 0: 30.234s (15 test files, min 0.123s, max 5.678s) max_time=5.678 min_time=0.123 test_count=15 total_time=30.234 worker=0
7:10PM INF P50  = 1.234s percentile=50 value=1.234
7:10PM INF P75  = 2.345s percentile=75 value=2.345
7:10PM INF P95  = 4.567s percentile=95 value=4.567
7:10PM INF P99  = 5.234s percentile=99 value=5.234
7:10PM INF P100 = 5.678s percentile=100 value=5.678
7:10PM INF Worker 1: 29.876s (14 test files, min 0.145s, max 5.234s) max_time=5.234 min_time=0.145 test_count=14 total_time=29.876 worker=1
7:10PM INF P50  = 1.456s percentile=50 value=1.456
...
7:10PM INF Rendering test files test_count=15 total_time=30.234 worker=0
7:10PM INF Split completed successfully tests_assigned=15 total_time=30.234
```

All log messages include structured fields for easy parsing and analysis.

**Percentiles**: By default, percentile statistics (P50, P75, P95, P99, P100) are shown for each worker's test distribution. Use `--no-percentiles` to disable this output.

## Common Development Tasks

### Building
```bash
go build -o tests-helper
```

### Running Tests
```bash
go test ./...
```

### Linting
The project uses [golangci-lint](https://golangci-lint.run/) with a strict configuration (`.golangci.yaml`).

```bash
# Run linter
golangci-lint run

# Run linter with auto-fix (recommended)
golangci-lint run --fix
```

The configuration includes 50+ linters and enforces:
- No global variables or `init()` functions
- Proper error handling and wrapping
- Structured logging best practices
- Code complexity limits
- Consistent code formatting
- Security checks (gosec)
- Performance optimizations

**Always run `golangci-lint run --fix` before committing code.**

### Adding a New Command
1. Create new file in `cmd/` directory
2. Define cobra command with `cobra.Command{}`
3. Register with `rootCmd.AddCommand()` in `init()`
4. Implement `RunE` function with business logic

### Adding New Features
- **New stat calculation**: Extend `internal/splitter/stats.go`
- **New input format**: Add parser in `internal/junit/`
- **New distribution algorithm**: Modify `internal/worker/worker.go`
- **New configuration**: Extend `internal/config/config.go`

## Error Handling

- Use `fmt.Errorf()` with `%w` for error wrapping
- Return errors up to the command level
- Cobra automatically prints errors and exits with status 1
- Log warnings for non-fatal issues (missing stats files, unparseable XML)

## Logging

The application uses zerolog for structured JSON logging with console output. All logs are written to stderr, while test file names are written to stdout.

- **Debug**: Detailed trace information (per-test assignments, parsing details)
- **Info**: Key operations (files loaded, distribution stats, worker assignments)
- **Warn**: Recoverable errors (missing stats files, malformed XML)
- **Error**: Fatal errors that prevent execution

Enable debug logging with `--debug` flag.

### Structured Fields
- Distribution stats: `total_time`, `avg_per_bucket`, `worker`, `test_count`, `min_time`, `max_time`
- Worker assignment: `tests_assigned`, `total_time`, `worker`
- Parsing: `count`, `file`, `pattern`

All log messages include relevant structured fields for easy parsing, filtering, and analysis.

## Testing Strategy

### Test Coverage
Current test coverage (as of latest run):
- `internal/config`: **100.0%** - Full coverage of configuration management
- `internal/worker`: **96.6%** - Core test distribution algorithm
- `internal/junit`: **91.2%** - JUnit XML parsing
- `internal/splitter`: **88.0%** - Test splitting orchestration

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run tests in short mode (skips E2E tests)
go test ./... -short

# Run tests with verbose output
go test ./... -v

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Structure

**Unit Tests**:
- `internal/junit/parser_test.go`: XML parsing with various formats, nested suites, locale handling
- `internal/worker/worker_test.go`: Distribution algorithm, edge cases, balanced distribution
- `internal/splitter/splitter_test.go`: Test reading, sorting, complete splitting workflow
- `internal/config/config_test.go`: Environment variable parsing, priority handling

**Integration Tests**:
- `internal/splitter/splitter_test.go::TestSplitter_Integration`: End-to-end with fixture files
- `cmd/split_test.go`: Command structure validation (E2E tests marked as skipped)

### Test Fixtures
All test fixtures are located in `testdata/`:
- `testdata/junit/*.xml`: Sample JUnit XML files (nested, comma decimals, multiple files)
- `testdata/testlists/*.txt`: Sample test file lists

### Test Data Patterns
- **Simple cases**: Basic distribution across 2-4 workers
- **Edge cases**: Empty tests, more workers than tests, single test
- **Real-world scenarios**: Using actual JUnit timing data
- **Error handling**: Invalid XML, missing files, empty input

## CI/CD Integration

### GitHub Actions

The project uses GitHub Actions for automated testing and releases:

**Tests Workflow** (`.github/workflows/tests.yml`):
- Triggers on all pull requests to master
- Reusable workflow that can be called by other workflows
- Runs two jobs in parallel:
  - **test**: Executes full test suite with coverage reporting
  - **lint**: Runs golangci-lint with 50+ linters

**Release Workflow** (`.github/workflows/release.yml`):
- Triggers on pushes to master and version tags
- Calls the Tests workflow as a nested workflow
- Only proceeds with release if tests pass
- Uses GoReleaser to build binaries for multiple platforms
- Creates GitHub releases with binaries and checksums

### CircleCI Integration

The tool is designed for CircleCI's parallel test execution:
- `CIRCLE_NODE_TOTAL`: Total number of parallel containers
- `CIRCLE_NODE_INDEX`: Current container index (0-based)

Example CircleCI config:
```yaml
jobs:
  test:
    parallelism: 4
    steps:
      - run: |
          go list ./... | \
          tests-helper split --stats "junit-*.xml" | \
          xargs go test -v
```

## Future Enhancements

Potential areas for expansion:
- Support for other test report formats (TAP, Subunit)
- Visualization of distribution balance
- Machine learning-based time prediction
- Caching of historical data
- Subcommands: `analyze`, `report`, `merge`

## Dependencies

- `github.com/spf13/cobra`: CLI framework
- `github.com/rs/zerolog`: Structured logging
- `github.com/caarlos0/env/v11`: Environment variable parsing
- Standard library: `encoding/xml`, `bufio`, `sort`, etc.

## Version

Current version: 1.0.0

## References

- Original implementation: `old.go`
- Cobra docs: https://github.com/spf13/cobra
- Zerolog docs: https://github.com/rs/zerolog
- CircleCI parallel docs: https://circleci.com/docs/parallelism-faster-jobs/
