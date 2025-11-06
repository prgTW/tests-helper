package config_test

import (
	"testing"

	"github.com/prgtw/tests-helper/internal/config"
)

func TestLoad(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if cfg.CircleNodeIndex != -1 {
			t.Errorf("CircleNodeIndex: got %d, want -1 (default)", cfg.CircleNodeIndex)
		}
		if cfg.CircleNodeTotal != -1 {
			t.Errorf("CircleNodeTotal: got %d, want -1 (default)", cfg.CircleNodeTotal)
		}
	})

	t.Run("from environment", func(t *testing.T) {
		t.Setenv("CIRCLE_NODE_INDEX", "2")
		t.Setenv("CIRCLE_NODE_TOTAL", "5")

		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if cfg.CircleNodeIndex != 2 {
			t.Errorf("CircleNodeIndex: got %d, want 2", cfg.CircleNodeIndex)
		}
		if cfg.CircleNodeTotal != 5 {
			t.Errorf("CircleNodeTotal: got %d, want 5", cfg.CircleNodeTotal)
		}
	})

	t.Run("invalid environment values", func(t *testing.T) {
		t.Setenv("CIRCLE_NODE_INDEX", "invalid")

		_, err := config.Load()
		if err == nil {
			t.Error("Expected error for invalid CIRCLE_NODE_INDEX, got nil")
		}
	})
}

func TestConfig_GetNodeIndex(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		flagValue    int
		defaultValue int
		want         int
	}{
		{
			name:         "flag takes precedence",
			envValue:     "5",
			flagValue:    3,
			defaultValue: 0,
			want:         3,
		},
		{
			name:         "env when no flag",
			envValue:     "7",
			flagValue:    -1,
			defaultValue: 0,
			want:         7,
		},
		{
			name:         "default when neither",
			envValue:     "",
			flagValue:    -1,
			defaultValue: 2,
			want:         2,
		},
		{
			name:         "zero flag is valid",
			envValue:     "5",
			flagValue:    0,
			defaultValue: 10,
			want:         0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv("CIRCLE_NODE_INDEX", tt.envValue)
			}

			cfg, err := config.Load()
			if err != nil {
				t.Fatalf("Load failed: %v", err)
			}

			got := cfg.GetNodeIndex(tt.flagValue, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetNodeIndex: got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestConfig_GetNodeTotal(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		flagValue    int
		defaultValue int
		want         int
	}{
		{
			name:         "flag takes precedence",
			envValue:     "10",
			flagValue:    4,
			defaultValue: 1,
			want:         4,
		},
		{
			name:         "env when no flag",
			envValue:     "8",
			flagValue:    -1,
			defaultValue: 1,
			want:         8,
		},
		{
			name:         "default when neither",
			envValue:     "",
			flagValue:    -1,
			defaultValue: 3,
			want:         3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv("CIRCLE_NODE_TOTAL", tt.envValue)
			}

			cfg, err := config.Load()
			if err != nil {
				t.Fatalf("Load failed: %v", err)
			}

			got := cfg.GetNodeTotal(tt.flagValue, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetNodeTotal: got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestConfig_Integration(t *testing.T) {
	// Simulate CircleCI environment
	t.Setenv("CIRCLE_NODE_INDEX", "1")
	t.Setenv("CIRCLE_NODE_TOTAL", "4")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Simulate no CLI flags (using -1)
	index := cfg.GetNodeIndex(-1, 0)
	total := cfg.GetNodeTotal(-1, 1)

	if index != 1 {
		t.Errorf("Index from CircleCI env: got %d, want 1", index)
	}
	if total != 4 {
		t.Errorf("Total from CircleCI env: got %d, want 4", total)
	}

	// Simulate CLI override
	indexOverride := cfg.GetNodeIndex(2, 0)
	totalOverride := cfg.GetNodeTotal(8, 1)

	if indexOverride != 2 {
		t.Errorf("Index from CLI override: got %d, want 2", indexOverride)
	}
	if totalOverride != 8 {
		t.Errorf("Total from CLI override: got %d, want 8", totalOverride)
	}
}
