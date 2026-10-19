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

	// Track currently active stream (from playlist or search)
	CurrentStation config.Station

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
	ti.Width = 48

	p := &Player{
		Stations:  stations,
		TextInput: ti,
		ViewState: ViewPlayer,
	}

	if len(stations) > 0 {
		p.CurrentStation = stations[0]
	}

	return p
}

func (m *Player) startPlayback() {
	if m.Cmd != nil && m.Cmd.Process != nil {
		_ = m.Cmd.Process.Kill()
		_ = m.Cmd.Wait()
	}

	if m.CurrentStation.URL == "" {
		return
	}

	m.Cmd = exec.Command("mpv", "--no-video", "--input-terminal=no", m.CurrentStation.URL)
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

// saveCurrentStation saves whichever station is currently playing into stations.json
func (m *Player) saveCurrentStation() {
	if m.CurrentStation.URL == "" {
		m.StatusMessage = "No active station to save"
		return
	}

	// Prevent duplicates
	for _, st := range m.Stations {
		if st.URL == m.CurrentStation.URL {
			m.StatusMessage = "Station already saved in list!"
			return
		}
	}

	m.Stations = append(m.Stations, m.CurrentStation)

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
		m.StatusMessage = "Error opening station file"
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		m.StatusMessage = "Error encoding config JSON"
		return
	}

	m.StatusMessage = fmt.Sprintf("Saved: %s", m.CurrentStation.Name)
}

func (m *Player) Init() tea.Cmd {
	if m.CurrentStation.URL != "" {
		m.startPlayback()
	}
	return textinput.Blink
}

func (m *Player) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
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
					m.CurrentStation = m.Stations[m.Index]
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

			case "a":
				m.saveCurrentStation()
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
					m.CurrentStation = m.Stations[m.Index]
					m.startPlayback()
				}
			case "down", "j":
				if m.Index < len(m.Stations)-1 {
					m.Index++
					m.CurrentStation = m.Stations[m.Index]
					m.startPlayback()
				}
			case "enter":
				m.ViewState = ViewPlayer
				m.CurrentStation = m.Stations[m.Index]
				m.startPlayback()
			}

		case ViewSearch:
			switch msg.String() {
			case "esc":
				m.ViewState = ViewPlayer
				return m, nil
			case "enter":
				// If search results are shown, play highlighted result directly
				if len(m.SearchResults) > 0 {
					selected := m.SearchResults[m.SearchIndex]
					m.CurrentStation = config.Station{Name: selected.Name, URL: selected.URL}
					m.ViewState = ViewPlayer
					m.startPlayback()
					return m, nil
				} else {
					// Otherwise, submit search query
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
			case "a":
				// Save highlighted result directly from search screen
				if len(m.SearchResults) > 0 {
					selected := m.SearchResults[m.SearchIndex]
					m.CurrentStation = config.Station{Name: selected.Name, URL: selected.URL}
					m.saveCurrentStation()
				}
			}

			// Pass keystrokes to search input
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
		Width(76)

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
		sb.WriteString("\n[↑/↓] Select  [Enter] Play  [l/esc] Back")
		return boxStyle.Render(sb.String())
	}

	if m.ViewState == ViewSearch {
		var sb strings.Builder
		sb.WriteString(titleStyle.Render("🔍 SEARCH RADIO BROWSER"))
		sb.WriteString("\n\n  Search: ")
		sb.WriteString(m.TextInput.View())
		sb.WriteString("\n\n──────────────────────────────────────────────────────────────────\n")

		if len(m.SearchResults) == 0 {
			sb.WriteString("\n  Type query and press [Enter] to search...\n  [esc] Back")
		} else {
			for i, st := range m.SearchResults {
				if i > 8 {
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
			sb.WriteString("\n[↑/↓] Navigate  [Enter] Preview & Play  [a] Add  [esc] Back")
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

	currentStation := "No station active"
	if m.CurrentStation.Name != "" {
		currentStation = m.CurrentStation.Name
		if len(currentStation) > 48 {
			currentStation = currentStation[:45] + "..."
		}
	}

	body := fmt.Sprintf(
		"%s\n\n  Status  : %s\n  Station : %s\n\n──────────────────────────────────────────────────────────────────\n  [n] Next  [p] Pause  [l] List  [s] Search  [a] Add  [q] Quit",
		titleStyle.Render("🎵 GO CLI RADIO"),
		statusStr,
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Render(currentStation),
	)

	if m.StatusMessage != "" {
		body += fmt.Sprintf("\n  Msg: %s", lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(m.StatusMessage))
	}

	return boxStyle.Render(body)
}
