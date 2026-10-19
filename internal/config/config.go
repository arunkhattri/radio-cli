// Package config handles parsing and loading of station configurations.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Station struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Config struct {
	Stations []Station `json:"stations"`
}

func LoadConfig(filename string) ([]Station, error) {
	var absPath string

	// Try ~/.config/radio-cli/ first
	home, err := os.UserHomeDir()
	if err == nil {
		globalPath := filepath.Join(home, ".config", "radio-cli", filename)
		if _, err := os.Stat(globalPath); err == nil {
			absPath = globalPath
		}
	}

	// Fallback to local working directory if not found globally
	if absPath == "" {
		absPath, err = filepath.Abs(filename)
		if err != nil {
			return nil, err
		}
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config at %s: %w", absPath, err)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if len(cfg.Stations) == 0 {
		return nil, fmt.Errorf("no stations found in %s", absPath)
	}
	return cfg.Stations, nil
}
