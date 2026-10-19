// Package player manages the Bubble Tea UI model, views, and audio playback lifecycle.
package player

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/arunkhattri/radio-cli/internal/config"
	"github.com/arunkhattri/radio-cli/internal/search"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ViewState int

const (
	ViewPlayer ViewState = iota
	ViewList
	ViewSearch
)

type Player struct {
	Stations  []config.Station
	Index     int
	IsPaused  bool
	ViewState ViewState
	Cmd       *exec.Cmd
	Err       error

	// Search state fields
	TextInput     textinput.Model
	SearchResults []search.StationResult
	SearchIndex   int
	StatusMessage string
}

func NewPlayer(stations []config.Station) *Player {
	ti := textinput.New()
	ti.Placeholder = "Type genre, country, or station name..."
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 40

	return &Player{
		Stations:  stations,
		TextInput: ti,
		ViewState: ViewPlayer,
	}
}

func (m *Player) startPlayback() {
	if m.Cmd != nil && m.Cmd.Process != nil {
		_ = m.Cmd.Process.Kill()
		_ = m.Cmd.Wait()
	}

	if len(m.Stations) == 0 {
		return
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

// saveStationToFile appends a newly discovered station to stations.json
func (m *Player) saveStationToFile(st search.StationResult) {
	newStation := config.Station{Name: st.Name, URL: st.URL}
	m.Stations = append(m.Stations, newStation)

	// Resolve global config path
	home, err := os.UserConfigDir()
	var targetPath string
	if err == nil {
		targetPath = filepath.Join(home, "radio-cli", "stations.json")
	} else {
		targetPath = "stations.json"
	}

	cfg := config.Config{Stations: m.Stations}
	file, err := os.Create(targetPath)
	if err != nil {
		m.StatusMessage = "Error saving station file"
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		m.StatusMessage = "Error encoding config JSON"
		return
	}

	m.StatusMessage = fmt.Sprintf("Saved: %s", st.Name)
}

func (m *Player) Init() tea.Cmd {
	if len(m.Stations) > 0 {
		m.startPlayback()
	}
	return textinput.Blink
}

func (m *Player) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear transient status message on keypress
		m.StatusMessage = ""

		switch m.ViewState {
		case ViewPlayer:
			switch msg.String() {
			case "q", "ctrl+c":
				if m.Cmd != nil && m.Cmd.Process != nil {
					_ = m.Cmd.Process.Kill()
				}
				return m, tea.Quit

			case "n":
				if len(m.Stations) > 0 {
					m.Index = (m.Index + 1) % len(m.Stations)
					m.startPlayback()
				}

			case "p":
				m.togglePause()

			case "l":
				m.ViewState = ViewList

			case "s":
				m.ViewState = ViewSearch
				m.TextInput.Focus()
				return m, textinput.Blink
			}

		case ViewList:
			switch msg.String() {
			case "q", "ctrl+c":
				if m.Cmd != nil && m.Cmd.Process != nil {
					_ = m.Cmd.Process.Kill()
				}
				return m, tea.Quit
			case "esc", "l":
				m.ViewState = ViewPlayer
			case "up", "k":
				if m.Index > 0 {
					m.Index--
					m.startPlayback()
				}
			case "down", "j":
				if m.Index < len(m.Stations)-1 {
					m.Index++
					m.startPlayback()
				}
			case "enter":
				m.ViewState = ViewPlayer
				m.startPlayback()
			}

		case ViewSearch:
			switch msg.String() {
			case "esc":
				m.ViewState = ViewPlayer
				return m, nil
			case "enter":
				// If results are visible, enter plays the highlighted search result
				if len(m.SearchResults) > 0 {
					selected := m.SearchResults[m.SearchIndex]
					// Temporarily prepend or play directly
					tempStation := config.Station{Name: selected.Name, URL: selected.URL}
					m.Stations = append([]config.Station{tempStation}, m.Stations...)
					m.Index = 0
					m.ViewState = ViewPlayer
					m.startPlayback()
					return m, nil
				} else {
					// Perform the search query
					query := m.TextInput.Value()
					results, err := search.SearchStations(query)
					if err != nil {
						m.StatusMessage = "No stations found"
						m.SearchResults = nil
					} else {
						m.SearchResults = results
						m.SearchIndex = 0
					}
				}
				return m, nil

			case "up", "k":
				if len(m.SearchResults) > 0 && m.SearchIndex > 0 {
					m.SearchIndex--
				}
			case "down", "j":
				if len(m.SearchResults) > 0 && m.SearchIndex < len(m.SearchResults)-1 {
					m.SearchIndex++
				}
			case "s":
				// Press 's' while highlighting a search result to save it to stations.json permanently
				if len(m.SearchResults) > 0 {
					m.saveStationToFile(m.SearchResults[m.SearchIndex])
				}
			}

			// Pass keys to text input when typing query
			if len(m.SearchResults) == 0 || msg.String() == "backspace" || len(msg.String()) == 1 {
				m.TextInput, cmd = m.TextInput.Update(msg)
				return m, cmd
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
		Width(62)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Align(lipgloss.Center)

	if m.ViewState == ViewList {
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
		sb.WriteString("\n[↑/↓] Select  [Enter] Play  [esc] Back")
		return boxStyle.Render(sb.String())
	}

	if m.ViewState == ViewSearch {
		var sb strings.Builder
		sb.WriteString(titleStyle.Render("🔍 SEARCH RADIO BROWSER"))
		sb.WriteString("\n\n  Search: ")
		sb.WriteString(m.TextInput.View())
		sb.WriteString("\n\n──────────────────────────────────────────────────────────\n")

		if len(m.SearchResults) == 0 {
			sb.WriteString("\n  Type query and press [Enter] to search...\n  [esc] Back")
		} else {
			for i, st := range m.SearchResults {
				if i > 8 { // Limit displayed list items for TUI fit
					break
				}
				cursor := "  "
				line := fmt.Sprintf("%s %s", cursor, st.Name)
				if i == m.SearchIndex {
					cursor = "▶"
					line = fmt.Sprintf("%s %s", cursor, st.Name)
					sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true).Render(line))
					sb.WriteString("\n")
				} else {
					sb.WriteString(line)
					sb.WriteString("\n")
				}
			}
			sb.WriteString("\n[↑/↓] Navigate  [Enter] Play  [s] Save to Config  [esc] Back")
		}

		if m.StatusMessage != "" {
			sb.WriteString(fmt.Sprintf("\n\n  Status: %s", lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(m.StatusMessage)))
		}

		return boxStyle.Render(sb.String())
	}

	// Default Player View
	statusPlaying := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).Render("▶ PLAYING")
	statusPaused := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("⏸ PAUSED")

	statusStr := statusPlaying
	if m.IsPaused {
		statusStr = statusPaused
	}

	if m.Err != nil {
		return fmt.Sprintf("Error: %v\n", m.Err)
	}

	currentStation := "No stations available"
	if len(m.Stations) > 0 {
		currentStation = m.Stations[m.Index].Name
		if len(currentStation) > 34 {
			currentStation = currentStation[:31] + "..."
		}
	}

	body := fmt.Sprintf(
		"%s\n\n  Status  : %s\n  Station : %s\n\n──────────────────────────────────────────────────────────\n  [n] Next  [p] Pause  [l] List  [s] Search  [q] Quit",
		titleStyle.Render("🎵 GO CLI RADIO"),
		statusStr,
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Render(currentStation),
	)

	if m.StatusMessage != "" {
		body += fmt.Sprintf("\n  Msg: %s", m.StatusMessage)
	}

	return boxStyle.Render(body)
}
