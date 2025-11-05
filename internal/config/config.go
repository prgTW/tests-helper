package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Config holds the application configuration.
type Config struct {
	// CircleCI environment variables
	CircleNodeIndex int `env:"CIRCLE_NODE_INDEX" envDefault:"-1"`
	CircleNodeTotal int `env:"CIRCLE_NODE_TOTAL" envDefault:"-1"`
}

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}
	return cfg, nil
}

// GetNodeIndex returns the node index, preferring flag value over env var.
func (c *Config) GetNodeIndex(flagValue int, defaultValue int) int {
	if flagValue >= 0 {
		return flagValue
	}
	if c.CircleNodeIndex >= 0 {
		return c.CircleNodeIndex
	}
	return defaultValue
}

// GetNodeTotal returns the total number of nodes, preferring flag value over env var.
func (c *Config) GetNodeTotal(flagValue int, defaultValue int) int {
	if flagValue >= 0 {
		return flagValue
	}
	if c.CircleNodeTotal >= 0 {
		return c.CircleNodeTotal
	}
	return defaultValue
}
