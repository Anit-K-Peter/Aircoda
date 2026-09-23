# Audio Streaming Engine

The Aircoda audio engine manages FFmpeg audio encoding and Icecast broadcasting.

---

## Streaming Flow

```text
Audio Files -> Playlist Engine -> FFmpeg MP3 Encoder -> Icecast Server -> HTTP Audio Stream
```

---

## Fallback Embedded Streaming
If system `icecast` is not installed, Aircoda automatically falls back to its built-in pure Go HTTP streaming hub (`StreamHub`) without failing.

---

## Audio Engine Controls

```bash
# Start live stream
radio start

# Foreground debug mode
radio start --foreground

# Stop live stream
radio stop

# Restart live stream
radio restart
```
