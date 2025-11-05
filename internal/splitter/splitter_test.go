package splitter_test

import (
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/tomaszwojcik/tests-helper/internal/splitter"
)

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
		defer file.Close()

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
		defer file.Close()

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
	defer inputFile.Close()

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

// floatEqual checks if two floats are equal within tolerance.
func floatEqual(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}
