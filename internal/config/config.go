// Package config handles parsing and loading of station configurations.
package config

import (
	"encoding/json"
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
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}
	return cfg.Stations, nil
}
