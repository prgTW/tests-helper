package junit_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"

	"github.com/prgtw/tests-helper/internal/junit"
)

func TestParser_LoadFiles(t *testing.T) {
	logger := zerolog.New(os.Stderr).Level(zerolog.Disabled)
	parser := junit.NewParser(logger)

	t.Run("single file", func(t *testing.T) {
		pattern := "../../testdata/junit/example1.xml"
		times, err := parser.LoadFiles([]string{pattern})
		if err != nil {
			t.Fatalf("LoadFiles failed: %v", err)
		}

		expected := map[string]float64{
			"pkg/service/auth_test.go": 5.234,
			"pkg/service/user_test.go": 3.456,
			"pkg/api/handler_test.go":  8.901,
		}

		for file, expectedTime := range expected {
			gotTime, exists := times[file]
			if !exists {
				t.Errorf("File %q not found in times", file)
				continue
			}
			if !floatEqual(gotTime, expectedTime, 0.001) {
				t.Errorf("File %q: got time=%.3f, want %.3f", file, gotTime, expectedTime)
			}
		}
	})

	t.Run("multiple files", func(t *testing.T) {
		patterns := []string{
			"../../testdata/junit/example1.xml",
			"../../testdata/junit/example2.xml",
		}
		times, err := parser.LoadFiles(patterns)
		if err != nil {
			t.Fatalf("LoadFiles failed: %v", err)
		}

		// example1.xml has auth_test.go with 5.234s
		// example2.xml has auth_test.go with 2.100s
		// Should accumulate to 7.334s
		authTime := times["pkg/service/auth_test.go"]
		expected := 7.334
		if !floatEqual(authTime, expected, 0.001) {
			t.Errorf("auth_test.go: got %.3f, want %.3f (accumulated)", authTime, expected)
		}

		// example2.xml has connection_test.go
		connTime := times["pkg/db/connection_test.go"]
		if !floatEqual(connTime, 12.567, 0.001) {
			t.Errorf("connection_test.go: got %.3f, want 12.567", connTime)
		}
	})

	t.Run("glob pattern", func(t *testing.T) {
		pattern := "../../testdata/junit/example*.xml"
		times, err := parser.LoadFiles([]string{pattern})
		if err != nil {
			t.Fatalf("LoadFiles failed: %v", err)
		}

		if len(times) == 0 {
			t.Error("No times loaded from glob pattern")
		}

		// Should have files from both example1 and example2
		if _, exists := times["pkg/api/handler_test.go"]; !exists {
			t.Error("Missing file from example1.xml")
		}
		if _, exists := times["pkg/db/connection_test.go"]; !exists {
			t.Error("Missing file from example2.xml")
		}
	})

	t.Run("nested testsuites", func(t *testing.T) {
		pattern := "../../testdata/junit/nested.xml"
		times, err := parser.LoadFiles([]string{pattern})
		if err != nil {
			t.Fatalf("LoadFiles failed: %v", err)
		}

		expected := map[string]float64{
			"pkg/nested/test1.go": 3.5,
			"pkg/nested/test2.go": 6.5,
		}

		for file, expectedTime := range expected {
			gotTime, exists := times[file]
			if !exists {
				t.Errorf("File %q not found in times", file)
				continue
			}
			if !floatEqual(gotTime, expectedTime, 0.001) {
				t.Errorf("File %q: got time=%.3f, want %.3f", file, gotTime, expectedTime)
			}
		}
	})

	t.Run("comma decimal separator", func(t *testing.T) {
		pattern := "../../testdata/junit/comma-decimal.xml"
		times, err := parser.LoadFiles([]string{pattern})
		if err != nil {
			t.Fatalf("LoadFiles failed: %v", err)
		}

		time := times["pkg/locale/test.go"]
		expected := 1.234
		if !floatEqual(time, expected, 0.001) {
			t.Errorf("comma-decimal parsing: got %.3f, want %.3f", time, expected)
		}
	})

	t.Run("no matching files", func(t *testing.T) {
		pattern := "../../testdata/junit/nonexistent-*.xml"
		_, err := parser.LoadFiles([]string{pattern})
		if err == nil {
			t.Error("Expected error for non-matching pattern, got nil")
		}
	})

	t.Run("invalid XML", func(t *testing.T) {
		// Create a temporary invalid XML file
		tmpDir := t.TempDir()
		invalidFile := filepath.Join(tmpDir, "invalid.xml")
		err := os.WriteFile(invalidFile, []byte("<invalid>not closed"), 0o600)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		times, err := parser.LoadFiles([]string{invalidFile})
		// Should not return error, but log warning and continue
		if err != nil {
			t.Errorf("LoadFiles returned error for invalid XML: %v", err)
		}
		// Should return empty or partial results
		if len(times) > 0 {
			t.Logf("Got %d times from invalid file (warnings logged)", len(times))
		}
	})

	t.Run("missing file", func(t *testing.T) {
		pattern := "/nonexistent/path/file.xml"
		_, err := parser.LoadFiles([]string{pattern})
		// Should return error for no matching files
		if err == nil {
			t.Error("Expected error for missing file pattern, got nil")
		}
	})
}

func TestParser_EmptyInput(t *testing.T) {
	logger := zerolog.New(os.Stderr).Level(zerolog.Disabled)
	parser := junit.NewParser(logger)

	_, err := parser.LoadFiles([]string{})
	if err == nil {
		t.Error("Expected error for empty patterns list, got nil")
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
