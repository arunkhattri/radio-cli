# Radio CLI 📻

A modern, lightweight, and modular terminal radio player built in Go. It combines the robust audio-streaming engine of `mpv` with a responsive terminal user interface (TUI) powered by Charmbracelet's Bubble Tea and Lip Gloss.

![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)
![License](https://img.shields.io/badge/License-GPLv3%2B-blue.svg)
![Status](https://img.shields.io/badge/status-v1.0.0-success)

---

## Features

* **Terminal UI (TUI):** Clean, interactive interface with a built-in station list view, custom status indicators, and sleek styling via Lip Gloss.
* **Reliable Audio Backend:** Powered by `mpv` with precise state management using system signals (`SIGSTOP`/`SIGCONT`) for instant pause/resume without tearing down streams.
* **Standard Go Architecture:** Clean separation of concerns following standard Go project layouts (`cmd/` for entry points, `internal/` for core logic and packages).
* **Global Configuration:** Automatically manages configuration files in the user's system configuration directory (`~/.config/radio-cli/stations.json`) with a local fallback for development.

---

## Project Structure

```text
.
├── cmd
│   └── radio
│       └── main.go           # Application entry point
└── internal
    ├── config
    │   └── config.go         # Station configuration parsing and path handling
    └── player
        └── player.go         # Bubble Tea model, view rendering, and mpv lifecycle
```

---

## Installation

### Prerequisites
Make sure you have the following installed on your system:
* [Go](https://golang.org/) (v1.21 or higher)
* [`mpv`](https://mpv.io/) (used for audio stream playback)

### Installing from Source
1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/radio-cli.git
   cd radio-cli
   ```
2. Build the binary:
   ```bash
   go build -o radio cmd/radio/main.go
   ```
3. Move the binary to a directory in your system's `PATH` (e.g., `/usr/local/bin`):
   ```bash
   sudo mv radio /usr/local/bin/
   ```

---

## Configuration

Radio CLI looks for your station configuration file inside your user configuration directory. 

1. Create the configuration directory:
   ```bash
   mkdir -p ~/.config/radio-cli
   ```
2. Create or copy your `stations.json` file into that directory:
   ```json
   {
     "stations": [
       {
         "name": "Lofi Girl Radio",
         "url": "https://stream.zeno.fm/f3wvbbqmdg8uv"
       },
       {
         "name": "Classical Public Radio",
         "url": "https://stream.publicradio.org/classical.m3u"
       }
     ]
   }
   ```

---

## Usage

Run the application globally from any terminal window:
```bash
radio
```

### Key Bindings
* **`n`**: Next station
* **`p`**: Pause / Resume playback
* **`l` or `esc`**: Open/close the station list view
* **`↑` / `k` & `↓` / `j`**: Navigate stations in the list view
* **`Enter`**: Play selected station from the list
* **`q` or `ctrl+c`**: Quit application

---

## Roadmap (Upcoming Features)

* [ ] **Station Search & Filtering:** Live text filtering to quickly find stations by name in the TUI.
* [ ] **CRUD Station Management:** Add, edit, or delete radio stations directly from the interface, automatically saving updates to `stations.json`.
* [ ] **Stream Recording:** Capture and record live audio streams directly to disk.

---

## License

This project is licensed under the terms of the **GNU General Public License v3.0 or later** (GPLv3+). 

Knowledge is meant to be shared freely. Anyone is welcome to run, study, share, and modify this software under the same copyleft terms. See the [LICENSE](LICENSE) file for full details.