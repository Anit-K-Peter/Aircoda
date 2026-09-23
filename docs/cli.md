# Aircoda CLI Documentation

`radio` is the global command line interface for managing self-hosted Aircoda internet radio servers.

## Installation

```bash
# Build standalone binary
go build -o radio ./cmd/radio

# Install globally to system PATH
sudo install ./radio /usr/local/bin/radio
```

## CLI Usage

### 1. Main Command

```bash
radio
```

- **First-run behavior**: If no configuration exists, automatically starts the interactive terminal setup wizard.
- **Normal behavior**: Opens the interactive CLI main menu displaying station metadata and option navigation.

---

### 2. Subcommands

#### `radio setup`
Runs the interactive setup wizard to configure or update station parameters:
- Station Name
- Station ID (slug)
- Audio Source (Local audio, YouTube playlist, Configure later)
- Local Audio Folder path
- Start on boot preference

Configuration is written in TOML format to `/etc/aircoda/config.toml` (or `~/.config/aircoda/config.toml` if non-root).

#### `radio status`
Displays basic station status, version, audio source configuration, and service state.

Example output:
```text
Aircoda v0.1.0

Station: 986 FM
ID:      986fm
Source:  Local audio
Folder:  /home/user/Music/radio

Streaming: Not implemented
Service:   Not configured
```

#### `radio config`
Prints the current configuration file location and raw TOML configuration file contents.

#### `radio version`
Displays version information, operating system, architecture, and Go runtime compiler version.

Example output:
```text
Aircoda v0.1.0
OS: linux
Architecture: amd64
Go Version: go1.26.0
```

#### `radio help`
Prints command usage summary and available flags.
