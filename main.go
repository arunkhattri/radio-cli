package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	"unicode/utf8"

	"golang.org/x/term"
)

type Station struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Config struct {
	Stations []Station `json:"stations"`
}

type Player struct {
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	stations []Station
	index    int
	isPaused bool
}

func loadConfig(filename string) ([]Station, error) {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", absPath, err)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse JSON in %s: %w", absPath, err)
	}

	if len(cfg.Stations) == 0 {
		return nil, fmt.Errorf("no stations found in %s", absPath)
	}

	return cfg.Stations, nil
}

func (p *Player) startStation() {
	p.stop()

	station := p.stations[p.index]
	p.cmd = exec.Command("mpv", "--no-video", "--input-terminal=no", station.URL)

	var err error
	p.stdin, err = p.cmd.StdinPipe()
	if err != nil {
		return
	}

	p.cmd.Stdout = nil
	p.cmd.Stderr = nil

	if err := p.cmd.Start(); err != nil {
		os.Exit(1)
	}

	p.isPaused = false
}

func (p *Player) togglePause() {
	if p.stdin != nil {
		p.stdin.Write([]byte("cycle pause\n"))
		p.isPaused = !p.isPaused
	}
}

func (p *Player) nextStation() {
	p.index = (p.index + 1) % len(p.stations)
	p.startStation()
}

func (p *Player) stop() {
	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
		_ = p.cmd.Wait()
	}
}

func padRight(s string, totalWidth int) string {
	rLen := utf8.RuneCountInString(s)
	if rLen >= totalWidth {
		return s
	}
	return s + fmt.Sprintf("%*s", totalWidth-rLen, "")
}

func truncate(s string, maxLen int) string {
	if utf8.RuneCountInString(s) > maxLen {
		runes := []rune(s)
		return string(runes[:maxLen-3]) + "..."
	}
	return s
}

func (p *Player) printStatus() {
	fmt.Print("\033[H\033[2J\033[3J")

	status := "PLAYING"
	if p.isPaused {
		status = "PAUSED"
	}

	current := p.stations[p.index]
	stationName := truncate(current.Name, 29)

	fmt.Println("╭─────────────────────────────────────────────╮")
	fmt.Println("│               GO CLI RADIO                  │")
	fmt.Println("├─────────────────────────────────────────────┤")
	fmt.Printf("│  Status  : [%-31s] │\n", status)
	fmt.Printf("│  Station : %-32s │\n", stationName)
	fmt.Println("├─────────────────────────────────────────────┤")
	fmt.Println("│  [n] Next   [p] Pause   [l] List   [q] Quit │")
	fmt.Println("╰─────────────────────────────────────────────╯")
}

func (p *Player) printList(oldState *term.State) {
	_ = term.Restore(int(os.Stdin.Fd()), oldState)

	fmt.Print("\033[H\033[2J\033[3J")
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│             AVAILABLE STATIONS              │")
	fmt.Println("├─────────────────────────────────────────────┤")
	for i, st := range p.stations {
		marker := " "
		if i == p.index {
			marker = ">"
		}
		line := fmt.Sprintf(" %s [%2d] %s", marker, i+1, st.Name)
		paddedLine := padRight(truncate(line, 43), 43)
		fmt.Printf("│ %s │\n", paddedLine)
	}
	fmt.Println("├─────────────────────────────────────────────┤")
	fmt.Println("│ Press Enter to return to player...          │")
	fmt.Println("└─────────────────────────────────────────────┘")

	var b [1]byte
	_, _ = os.Stdin.Read(b[:])

	_, _ = term.MakeRaw(int(os.Stdin.Fd()))
}

func main() {
	stations, err := loadConfig("stations.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	player := &Player{stations: stations}

	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		os.Exit(1)
	}

	cleanup := func() {
		player.stop()
		_ = term.Restore(fd, oldState)
		fmt.Print("\033[H\033[2J\033[3J")
		fmt.Println("Radio stopped. Goodbye!")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cleanup()
		os.Exit(0)
	}()

	player.startStation()
	player.printStatus()

	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		char := string(buf[0])
		switch char {
		case "q", "\x03":
			cleanup()
			return
		case "n":
			player.nextStation()
			player.printStatus()
		case "p":
			player.togglePause()
			player.printStatus()
		case "l":
			player.printList(oldState)
			player.printStatus()
		}
	}
}
