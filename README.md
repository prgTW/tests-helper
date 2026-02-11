# tests-helper

[![Release](https://github.com/prgtw/tests-helper/actions/workflows/release.yml/badge.svg)](https://github.com/prgtw/tests-helper/actions/workflows/release.yml)

A CLI tool that distributes a list of tests across parallel workers based on historical execution times from JUnit XML reports. It uses a greedy bin‑packing algorithm to keep workers balanced and reduce overall wall time.

## Features

- **Smart Test Distribution**: Uses historical timing data to balance test execution across workers
- **JUnit XML Support**: Parses JUnit test reports to extract execution times
- **CircleCI Integration**: Built-in support for CircleCI environment variables for parallel execution
- **Flexible Input**: Accepts a newline-separated test list via stdin and glob patterns for stats files
- **Detailed Statistics**: Provides comprehensive distribution metrics with percentiles
- **Structured Logging**: Uses zerolog with a human-readable console writer (logs to stderr)

## Installation

### Install via Go

```bash
go install github.com/prgtw/tests-helper@latest
```

### Download a Pre-built Binary (GitHub Releases)

Download the asset for your OS/arch from the [Releases page](https://github.com/prgtw/tests-helper/releases).

Release assets are published as **raw binaries** (not `.tar.gz`). Their names follow:

```text
tests-helper_<version>_<os>_<arch>
```

Example (macOS arm64):

```bash
VERSION="v1.2.3" # replace with an actual tag from Releases
curl -L -o tests-helper "https://github.com/prgtw/tests-helper/releases/download/${VERSION}/tests-helper_${VERSION#v}_darwin_arm64"
chmod +x tests-helper
sudo mv tests-helper /usr/local/bin/tests-helper
```

### Build from Source

```bash
git clone https://github.com/prgtw/tests-helper.git
cd tests-helper
go build -o tests-helper
```

## Quick Start

```bash
# Split tests across 4 workers, print the slice for worker 0
cat test-list.txt | tests-helper split --stats "junit-*.xml" --index 0 --total 4
```

## Usage

### Command Structure

```bash
tests-helper split [flags]
```

### Input contract

`tests-helper split` reads **one test identifier per line** from stdin. A “test identifier” is whatever your test runner uses as the unit you want to distribute (commonly **file paths** or **package names**).

Timing data is pulled from JUnit `<testsuite>` elements:

- Prefer matching the suite `file` attribute (e.g. `internal/foo/bar_test.go`)
- Also matches the suite `name` attribute (useful when your JUnit producer reports suites by package/module)

To get useful balancing, your stdin lines should match whatever your JUnit generator uses for suite `file` or `name`.

### Flags

| Flag | Description | Default |
| ---- | ----------- | ------- |
| `--stats` | Glob pattern(s) for JUnit XML files | - |
| `--index` | Worker index (0-based) | `$CIRCLE_NODE_INDEX` |
| `--total` | Total number of workers | `$CIRCLE_NODE_TOTAL` |
| `--debug` | Enable debug logging | `false` |
| `--no-percentiles` | Disable percentile statistics output | `false` |

### Examples

**Basic test splitting:**

```bash
cat test-list.txt | tests-helper split --stats "reports/junit-*.xml" --index 0 --total 4
```

**With CircleCI (automatic worker selection):**

```bash
# Uses CIRCLE_NODE_INDEX and CIRCLE_NODE_TOTAL environment variables
cat test-list.txt | tests-helper split --stats "previous-run/*.xml"
```

**Enable debug logging:**

```bash
cat tests.txt | tests-helper split --stats "*.xml" --debug --index 0 --total 2
```

**Without historical data:**

```bash
# Tests without historical data get a default time of 1.0 seconds (warn-logged per test)
cat tests.txt | tests-helper split --index 0 --total 3
```

## How It Works

1. **Read Input**: Reads test identifiers from stdin (one per line)
2. **Parse Stats**: Parses JUnit XML files to extract historical execution times
3. **Sort Tests**: Sorts tests by execution time (descending)
4. **Distribute**: Uses greedy algorithm to assign tests to workers
   - Always assigns next test to worker with minimum total time
   - Ensures balanced load distribution
5. **Output**: Prints assigned tests to stdout, statistics to stderr

### Algorithm

The tool uses a **greedy bin-packing algorithm**:

```text
For each test (sorted by time, descending):
    Assign test to worker with minimum total time
```

This approach ensures near-optimal distribution across workers, minimizing total execution time.

## Output Format

### stdout

Test identifiers assigned to the selected worker (one per line):

```text
./pkg/api/handler_test.go
./pkg/service/user_test.go
./internal/auth/token_test.go
```

### stderr (console-formatted structured logs)

Statistics and distribution information:

```text
7:10PM INF Starting test split index=0 total=4
7:10PM INF Read tests from input count=120
7:10PM INF Split tests across workers tests=120 workers=4
7:10PM INF === Distribution Summary ===
7:10PM INF Total time: 120.456s, Avg per bucket: 30.114s
7:10PM INF Worker 0: 30.234s (15 test files, min 0.123s, max 5.678s)
7:10PM INF P50  = 1.234s percentile=50 value=1.234
7:10PM INF P75  = 2.345s percentile=75 value=2.345
7:10PM INF P95  = 4.567s percentile=95 value=4.567
7:10PM INF Rendering test files
7:10PM INF Split completed successfully
```

## Generating JUnit XML (example for Go)

`tests-helper` only reads timing data; it doesn’t generate JUnit reports itself. One common approach for Go is [`gotestsum`](https://github.com/gotestyourself/gotestsum):

```bash
# 1) Run tests and write a JUnit report (one time or every run)
gotestsum --junitfile junit.xml -- ./...

# 2) Split the next run using timings from the previous report
go list ./... | tests-helper split --stats "junit.xml" --index 0 --total 4
```

## CircleCI Integration

### Example Configuration

```yaml
version: 2.1

jobs:
  test:
    docker:
      - image: cimg/go:1.25
    parallelism: 4
    steps:
      - checkout

      # Install tests-helper
      - run:
          name: Install tests-helper
          command: go install github.com/prgtw/tests-helper@latest

      # (Optional) Install gotestsum to produce JUnit XML
      - run:
          name: Install gotestsum
          command: go install gotest.tools/gotestsum@latest

      # Run tests with automatic splitting
      - run:
          name: Run tests
          command: |
            mkdir -p test-results

            # If you have previous JUnit XML available (from a workspace/artifact/cache),
            # point --stats at those files. Otherwise omit --stats to use default times.
            go list ./... | \
            tests-helper split --stats "stats/*.xml" | \
            xargs gotestsum --junitfile "test-results/junit.xml" -- -v

      # Store test results for next run
      - store_test_results:
          path: test-results

workflows:
  test:
    jobs:
      - test
```

### Environment Variables

- `CIRCLE_NODE_TOTAL`: Total number of parallel containers (automatically set)
- `CIRCLE_NODE_INDEX`: Current container index, 0-based (automatically set)

## CI/CD

### Pull Request Checks

Every pull request automatically runs:

- **Tests**: Full test suite with coverage reporting
- **Linting**: golangci-lint with 50+ linters to ensure code quality

The Tests workflow (`.github/workflows/tests.yml`) runs on all PRs to ensure code quality before merging.

### Releases

Releases are automatically created on every push to the `master` branch using [GoReleaser](https://goreleaser.com/). The workflow:

1. **Push to master**: Triggers the GitHub Actions workflow
2. **Run tests**: Executes the reusable Tests workflow (tests + linting)
3. **Build binaries**: Compiles for multiple platforms (Linux, macOS, Windows) and architectures (amd64, arm64)
4. **Create release**: Automatically creates a GitHub release with:
   - Pre-built binaries for all platforms
   - SHA256 checksums
   - Changelog generated from commit messages

### Creating a Tagged Release

For versioned releases, push a tag:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

This will create a formal release instead of a snapshot build.

## Development

### Prerequisites

- Go 1.25.x
- golangci-lint (for linting)
- goreleaser (optional, for local release testing)

### Building

```bash
go build -o tests-helper
```

### Testing GoReleaser Locally

```bash
# Test the release process without publishing
goreleaser release --snapshot --clean

# Check what would be released
goreleaser check
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Linting

```bash
# Run linter
golangci-lint run

# Run with auto-fix
golangci-lint run --fix
```

### Project Structure

```text
tests-helper/
├── main.go                    # Application entry point
├── cmd/                       # CLI commands
│   ├── root.go               # Root command
│   └── split.go              # Split subcommand
├── internal/                 # Private application code
│   ├── config/               # Configuration management
│   ├── junit/                # JUnit XML parsing
│   ├── splitter/             # Test splitting logic
│   └── worker/               # Worker allocation
└── testdata/                 # Test fixtures
```

## Testing Strategy

Current test coverage:

- `internal/config`: **100.0%**
- `internal/worker`: **96.6%**
- `internal/junit`: **91.2%**
- `internal/splitter`: **88.0%**

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and linter locally (`go test ./...` and `golangci-lint run --fix`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

**Note**: Pull requests automatically run the full test suite and linting checks via GitHub Actions. Ensure your code passes locally before pushing to save CI time.

## Credits

Built with:

- [cobra](https://github.com/spf13/cobra) - CLI framework
- [zerolog](https://github.com/rs/zerolog) - Structured logging
- [env](https://github.com/caarlos0/env) - Environment variable parsing

## Support

For issues and questions:

- Open an issue on GitHub
- Check existing issues for solutions
