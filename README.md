# 📻 Go CLI Radio Player

A lightweight, terminal-based internet radio player and station manager built in Go using Bubble Tea, Lip Gloss, and `mpv`. Discover, preview, and manage global radio streams seamlessly from your command line.

---

## ✨ Features

- **Live Station Search:** Search thousands of global radio streams instantly using the Radio Browser API (`goradios`).
- **Interactive TUI:** Clean, keyboard-driven interface powered by Charmbracelet components (`Bubble Tea` / `Lip Gloss`).
- **Instant Preview & Save:** Play any search result immediately to test the stream, and press `a` to save your favorites directly to your local configuration.
- **Persistent Management:** Automatically stores custom station lists in `~/.config/radio-cli/stations.json`.
- **Cross-Platform:** Native support for Linux, macOS, and Windows with robust process handling.

---

## 🚀 Installation & Usage

### Prerequisites
Make sure you have **`mpv`** installed on your system, as it powers the underlying audio playback:
- **Arch Linux:** `sudo pacman -S mpv`
- **Ubuntu/Debian:** `sudo apt install mpv`
- **macOS:** `brew install mpv`
- **Windows:** Download from [mpv.io](https://mpv.io/) or via `winget install mpv.mpv`

### Downloading Binaries
Head over to the [Releases](../../releases) page and download the pre-compiled binary for your operating system (`linux`, `darwin`, or `windows`).

### Building from Source
If you prefer to compile it yourself:

```bash
git clone [https://gitlab.com/AgentNemo/radio-cli.git](https://gitlab.com/AgentNemo/radio-cli.git)  # or your GitHub repository URL
cd radio-cli
go build -ldflags="-s -w" -o radio-cli .
