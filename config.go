package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the configuration for quick-ci
type Config struct {
	Repository             string   `json:"repository"`
	Commands               []string `json:"commands"`
	PollingIntervalMinutes int      `json:"polling_interval_minutes"`
	ResultsDirectory       string   `json:"results_directory"`
}

// LoadConfig loads and parses the configuration file
func LoadConfig(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Apply defaults
	if config.PollingIntervalMinutes == 0 {
		config.PollingIntervalMinutes = 15
	}
	if config.ResultsDirectory == "" {
		config.ResultsDirectory = "./ci-results"
	}

	// Validate
	if config.Repository == "" {
		return nil, fmt.Errorf("repository is required")
	}
	if len(config.Commands) == 0 {
		return nil, fmt.Errorf("at least one command is required")
	}

	return &config, nil
}
