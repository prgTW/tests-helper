package splitter_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/prgtw/tests-helper/internal/junit"
	"github.com/prgtw/tests-helper/internal/splitter"
)

//nolint:gocognit // Don't care
func TestSplitter_ReadTests(t *testing.T) {
	logger := zerolog.New(os.Stderr).Level(zerolog.Disabled)
	s := splitter.NewSplitter(logger)

	t.Run("simple list", func(t *testing.T) {
		input := "test1.go\ntest2.go\ntest3.go\n"
		times := map[string]float64{
			"test1.go": 5.0,
			"test2.go": 3.0,
		}

		tests, err := s.ReadTests(strings.NewReader(input), times)
		if err != nil {
			t.Fatalf("ReadTests failed: %v", err)
		}

		if len(tests) != 3 {
			t.Fatalf("Expected 3 tests, got %d", len(tests))
		}

		// Check that times are assigned correctly
		if tests[0].Time != 5.0 {
			t.Errorf("test1.go: got time=%.1f, want 5.0", tests[0].Time)
		}
		if tests[1].Time != 3.0 {
			t.Errorf("test2.go: got time=%.1f, want 3.0", tests[1].Time)
		}
		// test3.go has no historical data, should get default
		if tests[2].Time != 1.0 {
			t.Errorf("test3.go: got time=%.1f, want 1.0 (default)", tests[2].Time)
		}
	})

	t.Run("empty lines ignored", func(t *testing.T) {
		input := "test1.go\n\n\ntest2.go\n\n"
		tests, err := s.ReadTests(strings.NewReader(input), map[string]float64{})
		if err != nil {
			t.Fatalf("ReadTests failed: %v", err)
		}

		if len(tests) != 2 {
			t.Fatalf("Expected 2 tests, got %d", len(tests))
		}
	})

	t.Run("whitespace trimmed", func(t *testing.T) {
		input := "  test1.go  \n\ttest2.go\t\n"
		tests, err := s.ReadTests(strings.NewReader(input), map[string]float64{})
		if err != nil {
			t.Fatalf("ReadTests failed: %v", err)
		}

		if tests[0].Name != "test1.go" {
			t.Errorf("Expected trimmed name 'test1.go', got %q", tests[0].Name)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		input := ""
		_, err := s.ReadTests(strings.NewReader(input), map[string]float64{})
		if err == nil {
			t.Error("Expected error for empty input, got nil")
		}
	})

	t.Run("only whitespace", func(t *testing.T) {
		input := "\n\n  \n\t\n"
		_, err := s.ReadTests(strings.NewReader(input), map[string]float64{})
		if err == nil {
			t.Error("Expected error for whitespace-only input, got nil")
		}
	})

	t.Run("from fixture file", func(t *testing.T) {
		file, err := os.Open("../../testdata/testlists/simple.txt")
		if err != nil {
			t.Fatalf("Failed to open fixture: %v", err)
		}
		defer func(file *os.File) { _ = file.Close() }(file)

		tests, err := s.ReadTests(file, map[string]float64{})
		if err != nil {
			t.Fatalf("ReadTests failed: %v", err)
		}

		if len(tests) != 4 {
			t.Fatalf("Expected 4 tests from fixture, got %d", len(tests))
		}
	})

	t.Run("fixture with empty lines", func(t *testing.T) {
		file, err := os.Open("../../testdata/testlists/empty-lines.txt")
		if err != nil {
			t.Fatalf("Failed to open fixture: %v", err)
		}
		defer func(file *os.File) { _ = file.Close() }(file)

		tests, err := s.ReadTests(file, map[string]float64{})
		if err != nil {
			t.Fatalf("ReadTests failed: %v", err)
		}

		if len(tests) != 3 {
			t.Fatalf("Expected 3 tests (empty lines ignored), got %d", len(tests))
		}
	})
}

func TestSplitter_SortTests(t *testing.T) {
	logger := zerolog.New(os.Stderr).Level(zerolog.Disabled)
	s := splitter.NewSplitter(logger)

	tests := []struct {
		name     string
		input    string
		expected []string // expected order of names
	}{
		{
			name:     "descending order",
			input:    "a.go\nb.go\nc.go\n",
			expected: []string{"a.go", "b.go", "c.go"}, // all default time 1.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testList, err := s.ReadTests(strings.NewReader(tt.input), map[string]float64{})
			if err != nil {
				t.Fatalf("ReadTests failed: %v", err)
			}

			s.SortTests(testList)

			// Check descending order
			for i := 1; i < len(testList); i++ {
				if testList[i].Time > testList[i-1].Time {
					t.Errorf("Tests not sorted in descending order: %v", testList)
					break
				}
			}
		})
	}

	t.Run("sort by time descending", func(t *testing.T) {
		input := "small.go\nlarge.go\nmedium.go\n"
		times := map[string]float64{
			"small.go":  1.0,
			"large.go":  10.0,
			"medium.go": 5.0,
		}

		testList, err := s.ReadTests(strings.NewReader(input), times)
		if err != nil {
			t.Fatalf("ReadTests failed: %v", err)
		}

		s.SortTests(testList)

		expected := []string{"large.go", "medium.go", "small.go"}
		for i, test := range testList {
			if test.Name != expected[i] {
				t.Errorf("Position %d: got %q, want %q", i, test.Name, expected[i])
			}
		}
	})
}

func TestSplitter_Split(t *testing.T) {
	logger := zerolog.New(os.Stderr).Level(zerolog.Disabled)
	s := splitter.NewSplitter(logger)

	input := "test1.go\ntest2.go\ntest3.go\ntest4.go\n"
	times := map[string]float64{
		"test1.go": 10.0,
		"test2.go": 8.0,
		"test3.go": 6.0,
		"test4.go": 4.0,
	}

	tests, err := s.ReadTests(strings.NewReader(input), times)
	if err != nil {
		t.Fatalf("ReadTests failed: %v", err)
	}

	allocator := s.Split(tests, 2)

	t.Run("returns allocator", func(t *testing.T) {
		if allocator == nil {
			t.Fatal("Split returned nil allocator")
		}
	})

	t.Run("distributes all tests", func(t *testing.T) {
		stats := allocator.GetStats()
		totalTests := 0
		for _, ws := range stats.Workers {
			totalTests += ws.TestCount
		}
		if totalTests != 4 {
			t.Errorf("Total tests distributed: got %d, want 4", totalTests)
		}
	})

	t.Run("balanced distribution", func(t *testing.T) {
		worker0 := allocator.GetWorker(0)
		worker1 := allocator.GetWorker(1)

		// With times [10, 8, 6, 4], optimal split is:
		// Worker 0: 10 + 4 = 14
		// Worker 1: 8 + 6 = 14
		if worker0.Total != 14.0 || worker1.Total != 14.0 {
			t.Errorf("Distribution not balanced: worker0=%.1f, worker1=%.1f",
				worker0.Total, worker1.Total)
		}
	})
}

func TestSplitter_Integration(t *testing.T) {
	logger := zerolog.New(os.Stderr).Level(zerolog.Disabled)
	s := splitter.NewSplitter(logger)

	// Read from fixture
	inputFile, err := os.Open("../../testdata/testlists/simple.txt")
	if err != nil {
		t.Fatalf("Failed to open fixture: %v", err)
	}
	defer func(file *os.File) { _ = file.Close() }(inputFile)

	// Use times from JUnit fixture
	times := map[string]float64{
		"pkg/service/auth_test.go":  5.234,
		"pkg/service/user_test.go":  3.456,
		"pkg/api/handler_test.go":   8.901,
		"pkg/db/connection_test.go": 12.567,
	}

	tests, err := s.ReadTests(inputFile, times)
	if err != nil {
		t.Fatalf("ReadTests failed: %v", err)
	}

	allocator := s.Split(tests, 2)
	stats := allocator.GetStats()

	t.Run("all tests accounted for", func(t *testing.T) {
		expectedTotal := 5.234 + 3.456 + 8.901 + 12.567
		if !floatEqual(stats.TotalTime, expectedTotal, 0.001) {
			t.Errorf("TotalTime: got %.3f, want %.3f", stats.TotalTime, expectedTotal)
		}
	})

	t.Run("workers have tests", func(t *testing.T) {
		for i, ws := range stats.Workers {
			if ws.TestCount == 0 {
				t.Errorf("Worker %d has no tests", i)
			}
		}
	})
}

func TestSplitter_ReadTests_MatchesJUnitSuiteAttributes(t *testing.T) {
	logger := zerolog.New(os.Stderr).Level(zerolog.Disabled)
	s := splitter.NewSplitter(logger)

	cases := []readTestsMatchCase{
		{
			name:             "only file attribute",
			xml:              `<testsuites><testsuite file="pkg/only-file.go" time="2.5"/></testsuites>`,
			input:            "pkg/only-file.go\n",
			expectedTime:     2.5,
			expectedTimesLen: 1,
		},
		{
			name:             "only name attribute",
			xml:              `<testsuites><testsuite name="pkg/only-name.go" time="3.5"/></testsuites>`,
			input:            "pkg/only-name.go\n",
			expectedTime:     3.5,
			expectedTimesLen: 1,
		},
		{
			name:                "both attributes where only name matches",
			xml:                 `<testsuites><testsuite name="pkg/name-match.go" file="pkg/file-other.go" time="4.25"/></testsuites>`,
			input:               "pkg/name-match.go\n",
			expectedTime:        4.25,
			expectedKeysPresent: []string{"pkg/file-other.go"},
		},
		{
			name:                "both attributes where only file matches",
			xml:                 `<testsuites><testsuite name="pkg/name-other.go" file="pkg/file-match.go" time="5.25"/></testsuites>`,
			input:               "pkg/file-match.go\n",
			expectedTime:        5.25,
			expectedKeysPresent: []string{"pkg/name-other.go"},
		},
		{
			name:             "both attributes where both attributes match",
			xml:              `<testsuites><testsuite name="pkg/both-match.go" file="pkg/both-match.go" time="6.5"/></testsuites>`,
			input:            "pkg/both-match.go\n",
			expectedTime:     6.5,
			expectedTimesLen: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runReadTestsMatchCase(t, s, tc)
		})
	}
}

// floatEqual checks if two floats are equal within tolerance.
func floatEqual(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}

type readTestsMatchCase struct {
	name                string
	xml                 string
	input               string
	expectedTime        float64
	expectedKeysPresent []string
	expectedTimesLen    int
}

func runReadTestsMatchCase(t *testing.T, s *splitter.Splitter, tc readTestsMatchCase) {
	t.Helper()

	times := loadTimesFromXML(t, tc.xml)
	tests, err := s.ReadTests(strings.NewReader(tc.input), times)
	if err != nil {
		t.Fatalf("ReadTests failed: %v", err)
	}

	if len(tests) != 1 {
		t.Fatalf("Expected 1 test, got %d", len(tests))
	}
	if tests[0].Time != tc.expectedTime {
		t.Errorf("Expected time %.2f, got %.2f", tc.expectedTime, tests[0].Time)
	}

	for _, key := range tc.expectedKeysPresent {
		if _, exists := times[key]; !exists {
			t.Errorf("Expected time entry for %q", key)
		}
	}

	if tc.expectedTimesLen != 0 && len(times) != tc.expectedTimesLen {
		t.Errorf("Expected %d time entries, got %d", tc.expectedTimesLen, len(times))
	}
}

func loadTimesFromXML(t *testing.T, xmlContent string) map[string]float64 {
	t.Helper()

	logger := zerolog.New(os.Stderr).Level(zerolog.Disabled)
	parser := junit.NewParser(logger)

	tmpDir := t.TempDir()
	xmlPath := filepath.Join(tmpDir, "report.xml")
	if err := os.WriteFile(xmlPath, []byte(xmlContent), 0o600); err != nil {
		t.Fatalf("Failed to write XML fixture: %v", err)
	}

	times, err := parser.LoadFiles([]string{xmlPath})
	if err != nil {
		t.Fatalf("LoadFiles failed: %v", err)
	}

	return times
}
