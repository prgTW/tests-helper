# tests-helper - AI Agent Context

**tests-helper** is a CLI tool that distributes test files across parallel workers based on historical execution times from JUnit XML reports. User-facing documentation lives in `README.md`.

## Stack

- Go (version in `go.mod`)
- spf13/cobra for the CLI, rs/zerolog for logging, caarlos0/env for environment variables

## Core Behaviour

- **Algorithm**: greedy bin packing. Tests are sorted by time (descending) and each goes to the worker with the lowest total time.
- **Historical times**: read from `<testsuite file="..." time="...">` in JUnit XML files matched by `--stats` globs. Times accumulate across files and nested suites; comma decimal separators are accepted.
- **Missing data**: a test without historical time gets 1.0 s and a warning is logged.
- **Worker selection**: index/total come from flags first, then `CIRCLE_NODE_INDEX` / `CIRCLE_NODE_TOTAL`, then defaults (index 0, total 1). Only the selected worker's tests are printed.
- **I/O contract**: test list on stdin (error if stdin is a terminal or input is empty); assigned tests on stdout, one per line; all logs and statistics on stderr. Never write anything else to stdout.

## Where to Change Things

- New command: file in `cmd/`, register with `rootCmd.AddCommand()` in `init()`, logic in `RunE`
- Statistics and percentiles: `internal/splitter/stats.go`
- Input formats: `internal/junit/`
- Distribution algorithm: `internal/worker/worker.go`
- Configuration: `internal/config/config.go`
- Version info: `cmd` package variables, injected by GoReleaser via ldflags

## Conventions

- Wrap errors with `fmt.Errorf("...: %w", err)` and return them up to the command; cobra prints them and exits 1.
- Non-fatal problems (missing stats files, unparseable XML) are logged as warnings, not returned as errors.
- Log levels: Debug for per-test detail, Info for key steps, Warn for recoverable problems. `--debug` enables Debug.
- Log values as structured zerolog fields, not only in the message text.
- Test fixtures go in `testdata/` (`junit/*.xml`, `testlists/*.txt`).

## Before Committing

```bash
go test ./...
golangci-lint run --fix
```

CI runs tests, golangci-lint (strict config in `.golangci.yaml`) and govulncheck on every PR; all must pass.
