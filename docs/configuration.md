# Configuration Reference

Aircoda configuration is stored in TOML format at platform-standard locations:
- **Linux / macOS**: `~/.config/Aircoda/config.toml` (or `/etc/aircoda/config.toml`)
- **Windows**: `%APPDATA%\Aircoda\config.toml`

---

## Example `config.toml`

```toml
version = 1

[station]
name = "986 FM"
id = "986fm"

[audio]
source = "local"
directory = "/home/user/Music/radio"

[playlist]
mode = "shuffle"
repeat = true
recursive = true

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

## Configuration Backup & Restore

```bash
# Backup configuration
radio config backup

# Restore configuration
radio config restore /path/to/config.backup.toml
```

## Reset Configuration

```bash
radio reset
```
Resets station configuration while strictly preserving audio files.
