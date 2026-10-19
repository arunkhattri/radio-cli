// Package main provides the entry point for the CLI radio application.
package main

import (
	"fmt"
	"os"

	"github.com/arunkhattri/radio-cli/internal/config"
	"github.com/arunkhattri/radio-cli/internal/player"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	stations, err := config.LoadConfig("stations.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading stations.json: %v\n", err)
		os.Exit(1)
	}

	initialPlayer := player.NewPlayer(stations)
	p := tea.NewProgram(initialPlayer)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
