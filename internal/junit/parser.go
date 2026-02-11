package junit

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
)

// Parser handles parsing of JUnit XML files.
type Parser struct {
	logger zerolog.Logger
}

// NewParser creates a new JUnit parser.
func NewParser(logger zerolog.Logger) *Parser {
	return &Parser{logger: logger}
}

// LoadFiles loads and parses multiple JUnit XML files, returning a map of test names to execution times.
func (p *Parser) LoadFiles(patterns []string) (map[string]float64, error) {
	times := make(map[string]float64)

	// Expand glob patterns
	var files []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			p.logger.Warn().
				Err(err).
				Str("pattern", pattern).
				Msg("Invalid glob pattern")
			continue
		}
		files = append(files, matches...)
	}

	if len(files) == 0 {
		return times, errors.New("no files matched the provided patterns")
	}

	// Load each file
	var loadErrs []error
	for _, file := range files {
		if err := p.loadFile(file, times); err != nil {
			p.logger.Warn().
				Err(err).
				Str("file", file).
				Msg("Failed to load file")
			loadErrs = append(loadErrs, err)
		}
	}

	if len(loadErrs) > 0 {
		return times, fmt.Errorf(
			"failed to load %d/%d JUnit XML files: %w",
			len(loadErrs),
			len(files),
			errors.Join(loadErrs...),
		)
	}

	return times, nil
}

// loadFile loads a single JUnit XML file and accumulates test times.
func (p *Parser) loadFile(path string, times map[string]float64) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	var root TestSuites
	if parseErr := xml.Unmarshal(data, &root); parseErr != nil {
		return fmt.Errorf("cannot parse XML: %w", parseErr)
	}

	count := 0
	p.accumulateTimes(root.TestSuites, times, &count)

	p.logger.Info().
		Int("count", count).
		Str("file", filepath.Base(path)).
		Msg("Loaded test times")

	return nil
}

// accumulateTimes recursively accumulates test times from test suites.
func (p *Parser) accumulateTimes(suites []TestSuite, times map[string]float64, count *int) {
	for _, suite := range suites {
		if val, ok := parseTime(suite.Time); ok {
			p.addTime(times, suite.File, val, "file", count)
			if suite.Name != suite.File {
				p.addTime(times, suite.Name, val, "name", count)
			}
		}
		// Recursively process nested test suites
		p.accumulateTimes(suite.TestSuites, times, count)
	}
}

func parseTime(timeStr string) (float64, bool) {
	if timeStr == "" {
		return 0, false
	}

	// Normalize time string (replace comma with dot for some locales)
	normalized := strings.ReplaceAll(timeStr, ",", ".")
	val, err := strconv.ParseFloat(normalized, 64)
	if err != nil {
		return 0, false
	}

	return val, true
}

func (p *Parser) addTime(times map[string]float64, key string, val float64, field string, count *int) {
	if key == "" {
		return
	}

	times[key] += val
	p.logger.Debug().
		Str(field, key).
		Float64("time", val).
		Msg("Accumulated test time")
	*count++
}
