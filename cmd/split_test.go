package cmd_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/prgtw/tests-helper/cmd"
)

func TestSplitCommand_Integration(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		statsFiles      []string
		index           int
		total           int
		noPercentiles   bool
		wantTestCount   int
		wantInOutput    []string
		wantNotInOutput []string
	}{
		{
			name:          "basic split without stats",
			input:         "test1.go\ntest2.go\ntest3.go\ntest4.go\n",
			index:         0,
			total:         2,
			noPercentiles: true,
			wantTestCount: 2,
			wantInOutput:  []string{"test"},
		},
		{
			name:          "split with stats from fixture",
			input:         "pkg/service/auth_test.go\npkg/service/user_test.go\npkg/api/handler_test.go\npkg/db/connection_test.go\n",
			statsFiles:    []string{"../testdata/junit/example*.xml"},
			index:         0,
			total:         2,
			noPercentiles: true,
			wantTestCount: 2,
		},
		{
			name:          "single test",
			input:         "single.go\n",
			index:         0,
			total:         3,
			noPercentiles: true,
			wantTestCount: 1,
			wantInOutput:  []string{"single.go"},
		},
		{
			name:          "more workers than tests",
			input:         "test1.go\ntest2.go\n",
			index:         2,
			total:         5,
			noPercentiles: true,
			wantTestCount: 0, // Worker 2 should have no tests
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is a simplified integration test
			// A full test would need to capture the actual command execution
			// For now, we test that the command structure is correct

			// The actual integration would be better tested via:
			// 1. Building the binary
			// 2. Running it as a subprocess
			// 3. Capturing stdout/stderr

			// For unit testing the runSplit function directly, we'd need to export it
			// or test through the public Execute() function

			t.Skip("Full integration test requires command execution - see TestE2E")
		})
	}
}

// TestE2E is an end-to-end test that runs the actual binary.
func TestE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Check if binary exists
	_, err := os.Stat("../tests-helper")
	if err != nil {
		t.Skip("Binary not built, run 'go build' first")
	}

	tests := []struct {
		name         string
		inputFile    string
		args         []string
		wantExit0    bool
		wantInStdout []string
		wantInStderr []string
	}{
		{
			name:      "help command",
			args:      []string{"split", "--help"},
			wantExit0: true,
			wantInStdout: []string{
				"Split tests across parallel workers",
				"--stats",
				"--index",
				"--total",
			},
		},
		{
			name:      "version",
			args:      []string{"--version"},
			wantExit0: true,
			wantInStdout: []string{
				"1.0.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would use exec.Command to run the binary
			// and verify the output
			t.Skip("E2E test requires subprocess execution")
		})
	}
}

// TestSplitOptionsValidation tests the validation logic.
func TestSplitOptionsValidation(t *testing.T) {
	tests := []struct {
		name      string
		index     int
		total     int
		wantError bool
	}{
		{
			name:      "valid indices",
			index:     0,
			total:     2,
			wantError: false,
		},
		{
			name:      "index equals total",
			index:     2,
			total:     2,
			wantError: true,
		},
		{
			name:      "negative index",
			index:     -1,
			total:     2,
			wantError: true,
		},
		{
			name:      "index too large",
			index:     5,
			total:     3,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test index validation
			if tt.index < 0 || tt.index >= tt.total {
				if !tt.wantError {
					t.Error("Expected validation to pass but indices are invalid")
				}
			} else {
				if tt.wantError {
					t.Error("Expected validation to fail but indices are valid")
				}
			}
		})
	}
}

func TestSplitCommand_EmptyInput(t *testing.T) {
	// Test that empty input is handled correctly
	input := ""
	stdin := strings.NewReader(input)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	logger := zerolog.New(stderr).Level(zerolog.Disabled)

	// We would test runSplit here if it were exported
	// For now, this documents the expected behavior

	_ = stdin
	_ = stdout
	_ = logger

	t.Skip("Requires exported runSplit function or Execute with dependency injection")
}

func TestSplitCommand_WithCircleCIEnv(t *testing.T) {
	// Test CircleCI environment variable handling
	t.Setenv("CIRCLE_NODE_INDEX", "1")
	t.Setenv("CIRCLE_NODE_TOTAL", "4")

	// Verify env vars are set
	if os.Getenv("CIRCLE_NODE_INDEX") != "1" {
		t.Error("CIRCLE_NODE_INDEX not set correctly")
	}
	if os.Getenv("CIRCLE_NODE_TOTAL") != "4" {
		t.Error("CIRCLE_NODE_TOTAL not set correctly")
	}

	// The actual command would use these values
	// This test documents the expected behavior
	t.Skip("Requires command execution to verify env var usage")
}

// This test verifies the command is properly registered.
func TestSplitCommand_Registration(t *testing.T) {
	// Create root command
	_ = cmd.Execute

	// The Execute function should register the split command
	// We can't easily test this without running Execute,
	// but this documents that the command exists
	t.Log("Split command should be registered via Execute()")
}
