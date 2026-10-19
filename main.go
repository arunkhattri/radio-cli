package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Station struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Config struct {
	Stations []Station `json:"stations"`
}

type model struct {
	stations []Station
	index    int
	isPaused bool
	showList bool
	cmd      *exec.Cmd
	err      error
}

func loadConfig(filename string) ([]Station, error) {
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

func (m *model) startPlayback() {
	if m.cmd != nil && m.cmd.Process != nil {
		_ = m.cmd.Process.Kill()
		_ = m.cmd.Wait()
	}

	station := m.stations[m.index]
	m.cmd = exec.Command("mpv", "--no-video", "--input-terminal=no", station.URL)
	m.cmd.Stdout = nil
	m.cmd.Stderr = nil

	if err := m.cmd.Start(); err != nil {
		m.err = err
	}
	m.isPaused = false
}

func (m *model) togglePause() {
	if m.cmd != nil && m.cmd.Process != nil {
		if m.isPaused {
			_ = m.cmd.Process.Signal(syscall.SIGCONT)
			m.isPaused = false
		} else {
			_ = m.cmd.Process.Signal(syscall.SIGSTOP)
			m.isPaused = true
		}
	}
}

func (m *model) Init() tea.Cmd {
	m.startPlayback()
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.cmd != nil && m.cmd.Process != nil {
				_ = m.cmd.Process.Kill()
			}
			return m, tea.Quit

		case "n":
			m.index = (m.index + 1) % len(m.stations)
			m.startPlayback()

		case "p":
			m.togglePause()

		case "l", "esc":
			m.showList = !m.showList

		case "up", "k":
			if m.showList && m.index > 0 {
				m.index--
				m.startPlayback()
			}

		case "down", "j":
			if m.showList && m.index < len(m.stations)-1 {
				m.index++
				m.startPlayback()
			}

		case "enter":
			if m.showList {
				m.startPlayback()
				m.showList = false
			}
		}
	}
	return m, nil
}

func (m *model) View() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(54)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Align(lipgloss.Center)

	statusPlaying := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).Render("▶ PLAYING")
	statusPaused := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("⏸ PAUSED")

	statusStr := statusPlaying
	if m.isPaused {
		statusStr = statusPaused
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	if m.showList {
		var sb strings.Builder
		sb.WriteString(titleStyle.Render("📻 STATION LIST"))
		sb.WriteString("\n\n")
		for i, st := range m.stations {
			cursor := "  "
			line := fmt.Sprintf("%s[%2d] %s", cursor, i+1, st.Name)
			if i == m.index {
				cursor = "▶ "
				line = fmt.Sprintf("%s[%2d] %s", cursor, i+1, st.Name)
				sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true).Render(line))
				sb.WriteString("\n")
			} else {
				sb.WriteString(line)
				sb.WriteString("\n")
			}
		}
		sb.WriteString("\n[↑/↓] Select  [Enter] Play  [l/esc] Back")
		return boxStyle.Render(sb.String())
	}

	currentStation := m.stations[m.index].Name
	if len(currentStation) > 32 {
		currentStation = currentStation[:29] + "..."
	}

	body := fmt.Sprintf(
		"%s\n\n  Status  : %s\n  Station : %s\n\n──────────────────────────────────────────────\n  [n] Next   [p] Pause   [l] List   [q] Quit",
		titleStyle.Render("🎵 GO CLI RADIO"),
		statusStr,
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Render(currentStation),
	)

	return boxStyle.Render(body)
}

func main() {
	stations, err := loadConfig("stations.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading stations.json: %v\n", err)
		os.Exit(1)
	}

	initialModel := &model{stations: stations}
	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
