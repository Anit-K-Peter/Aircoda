# Aircoda

Aircoda is an open-source, self-hosted 24/7 internet radio server and CLI controller written in Go.

It combines an interactive Bubbletea terminal user interface (TUI), a Cobra CLI (`radio`), a continuous Master Encoder streaming pipeline, automatic Cloudflare Quick Tunnels for public broadcasting, YouTube audio downloader support, Station ID jingle drops, and time-based playlist daypart scheduling.

---

## Architecture

```text
                    AIRCODA
                       │
                 ┌─────▼─────┐
                 │ radio CLI │ (TUI / Commands)
                 └─────┬─────┘
                       │
                 ┌─────▼─────┐
                 │  Aircoda  │
                 │   Core    │
                 └─────┬─────┘
                       │
          ┌────────────┼────────────┐
          ▼            ▼            ▼
      Playlist      Master        Icecast /
     Manager       Encoder       StreamHub
          │            │            │
          └────────────┼────────────┘
                       ▼
                 LIVE STREAM
                       │
          ┌────────────┴────────────┐
          ▼                         ▼
      LOCAL STREAM            PUBLIC ACCESS
 (http://127.0.0.1:8000)  (Cloudflare Quick Tunnel)
```

---

## Key Features

- **Interactive Terminal UI (TUI)**: Driven by Bubbletea and Lipgloss, providing real-time status dashboards, queue managers, and live progress indicators.
- **Continuous Master Encoder**: Employs a long-running Master FFmpeg Encoder to prevent stream drops and listener disconnections when songs switch.
- **Guided Setup Wizard**: Single-command interactive setup with cross-platform folder picker support (Linux Zenity/Kdialog, macOS osascript, Windows PowerShell, with text-based fallback).
- **Automatic Self-Repair**: Built-in dependency repair system (`radio repair`) that auto-downloads missing binaries (`yt-dlp`, `cloudflared`) to `~/.config/Aircoda/bin/` and configures automatic fallbacks.
- **Audio Dependencies**:
  - `ffmpeg` & `ffprobe` (Required for audio encoding/decoding)
  - `icecast` / `icecast2` (Optional; built-in embedded HTTP streamer is automatically used if Icecast is not installed)
  - `yt-dlp` (Optional; auto-downloaded on demand for YouTube links)

---

## Building & Installation

### 1. Clone Repository & Build

```bash
git clone https://github.com/yourusername/aircoda.git
cd aircoda
go build -o radio ./cmd/radio
```

### 2. Install Executable

```bash
# Linux / macOS
cp radio ~/.local/bin/

# System-wide (Linux / macOS)
sudo cp radio /usr/local/bin/

# Windows (PowerShell as Administrator)
Copy-Item radio.exe C:\Windows\System32\
```

---

## Quick Start

Launch the interactive interface by running `radio` in your terminal:

```bash
radio
```

On first run, Aircoda automatically launches the setup wizard to configure:
1. Station Name & ID (e.g., `986 FM`, `986fm`)
2. Local Audio Folder Path
3. Playback Mode (`Shuffle` or `Sequential`)
4. Startup preferences

Configuration is saved in TOML format to platform-standard configuration paths (`~/.config/Aircoda/config.toml` or `%APPDATA%\Aircoda\config.toml`).

---

## CLI Reference

| Command | Description |
| --- | --- |
| `radio` | Opens main interactive TUI dashboard or first-run setup wizard. |
| `radio start [--foreground]` | Starts station streaming engine in detached background mode (or foreground). |
| `radio stop` | Cleanly stops live station streams and background processes. |
| `radio restart` | Restarts live station streaming processes. |
| `radio status` | Displays station streaming status, Now Playing track, and stream URL. |
| `radio playlist [scan]` | Manages playlist queue or forces a library rescan. |
| `radio config [get/set]` | Views or modifies station configuration settings. |
| `radio doctor` | Runs system dependency and environment diagnostics (`radio doctor --fix` for auto-repair). |
| `radio repair` | Scans host system and auto-installs missing dependencies (`yt-dlp`, `cloudflared`). |
| `radio service [install/start/stop/status]` | Manages background user service (`systemd` / `launchd`). |
| `radio public` | Manages public stream access mode (local, direct IP, tunnel). |
| `radio tunnel [setup/start/stop/status]` | Configures and manages Cloudflare Quick Tunnels. |
| `radio logs [-f]` | Displays recent station log entries. |
| `radio version` | Prints Aircoda version and build metadata. |

---

## Configuration Reference (`config.toml`)

```toml
version = 1

[station]
name = "986 FM"
id = "986fm"

[audio]
source = "local"
directory = "/home/user/Music"

[playlist]
mode = "shuffle"
repeat = true
recursive = true
station_id_enabled = false
station_id_interval = 5
station_id_path = ""

[schedule]
enabled = false
morning_path = "/home/user/Music/Morning"
afternoon_path = "/home/user/Music/Afternoon"
evening_path = "/home/user/Music/Evening"
night_path = "/home/user/Music/Night"

[streaming]
host = "127.0.0.1"
port = 8000
mount = "/986fm"
bitrate = 128
format = "mp3"
source_password = "hackme"
admin_password = "hackme"

[public]
mode = "local"
hostname = ""

[tunnel]
enabled = false
provider = "cloudflare"
hostname = ""
target = "127.0.0.1:8000"

[system]
start_on_boot = true
log_level = "info"
```

---

## License

Aircoda is licensed under the [MIT License](LICENSE).
