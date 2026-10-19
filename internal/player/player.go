// Package player manages the Bubble Tea UI model and the underlying mpv audio playback lifecycle.
package player

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"github.com/arunkhattri/radio-cli/internal/config"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Player struct {
	Stations []config.Station
	Index    int
	IsPaused bool
	ShowList bool
	Cmd      *exec.Cmd
	Err      error
}

func NewPlayer(stations []config.Station) *Player {
	return &Player{Stations: stations}
}

func (m *Player) startPlayback() {
	if m.Cmd != nil && m.Cmd.Process != nil {
		_ = m.Cmd.Process.Kill()
		_ = m.Cmd.Wait()
	}

	station := m.Stations[m.Index]
	m.Cmd = exec.Command("mpv", "--no-video", "--input-terminal=no", station.URL)
	m.Cmd.Stdout = nil
	m.Cmd.Stderr = nil

	if err := m.Cmd.Start(); err != nil {
		m.Err = err
	}
	m.IsPaused = false
}

func (m *Player) togglePause() {
	if m.Cmd != nil && m.Cmd.Process != nil {
		if m.IsPaused {
			_ = m.Cmd.Process.Signal(syscall.SIGCONT)
			m.IsPaused = false
		} else {
			_ = m.Cmd.Process.Signal(syscall.SIGSTOP)
			m.IsPaused = true
		}
	}
}

func (m *Player) Init() tea.Cmd {
	m.startPlayback()
	return nil
}

func (m *Player) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.Cmd != nil && m.Cmd.Process != nil {
				_ = m.Cmd.Process.Kill()
			}
			return m, tea.Quit

		case "n":
			m.Index = (m.Index + 1) % len(m.Stations)
			m.startPlayback()

		case "p":
			m.togglePause()

		case "l", "esc":
			m.ShowList = !m.ShowList

		case "up", "k":
			if m.ShowList && m.Index > 0 {
				m.Index--
				m.startPlayback()
			}

		case "down", "j":
			if m.ShowList && m.Index < len(m.Stations)-1 {
				m.Index++
				m.startPlayback()
			}

		case "enter":
			if m.ShowList {
				m.startPlayback()
				m.ShowList = false
			}
		}
	}
	return m, nil
}

func (m *Player) View() string {
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
	if m.IsPaused {
		statusStr = statusPaused
	}

	if m.Err != nil {
		return fmt.Sprintf("Error: %v\n", m.Err)
	}

	if m.ShowList {
		var sb strings.Builder
		sb.WriteString(titleStyle.Render("📻 STATION LIST"))
		sb.WriteString("\n\n")
		for i, st := range m.Stations {
			cursor := "  "
			line := fmt.Sprintf("%s[%2d] %s", cursor, i+1, st.Name)
			if i == m.Index {
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

	currentStation := m.Stations[m.Index].Name
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
