# Troubleshooting & Diagnostics

Aircoda includes built-in diagnostic tools to help troubleshoot setup and runtime issues.

---

## 1. Diagnostics Tool (`radio doctor`)

Run `radio doctor` to execute complete system health checks:

```bash
radio doctor
```

It inspects:
- Core installation & TOML configuration
- Audio directory existence & track scan counts
- Dependencies (`ffmpeg`, `ffprobe`, `icecast`, `cloudflared`)
- Stream port availability (Port 8000)
- Native OS background service support
- Public Access & Tunnel connection state

---

## 2. Viewing Station Logs

```bash
# View last 30 log lines
radio logs

# Follow log output continuously
radio logs --follow
```

---

## 3. Common Issues & Solutions

### Missing FFmpeg
- **Error**: `FFmpeg is required but was not found in PATH.`
- **Solution**: Install FFmpeg for your OS package manager (`sudo apt install ffmpeg` on Debian/Ubuntu, `brew install ffmpeg` on macOS, or `choco install ffmpeg` on Windows).

### Stream Port Conflict
- **Error**: `Stream Port CONFLICT (Port 8000 in use)`
- **Solution**: Stop conflicting services or change `[streaming].port` in `config.toml`.

### Empty Audio Directory
- **Error**: `No playable audio files were found.`
- **Solution**: Place supported `.mp3`, `.wav`, or `.flac` files into your audio directory and run `radio playlist scan`.
