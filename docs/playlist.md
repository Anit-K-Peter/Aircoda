# Playlist & Audio Library Engine

Aircoda recursively scans local music directories for supported audio formats.

---

## Supported Audio Formats
- MP3 (`.mp3`)
- WAV (`.wav`)
- OGG (`.ogg`)
- FLAC (`.flac`)
- M4A (`.m4a`)
- AAC (`.aac`)

---

## Playback Modes

### Shuffle Mode
Randomizes track selection from the active queue. When `repeat = true`, re-shuffles the queue upon reaching the end.

### Sequential Mode
Plays tracks in alphabetical file order.

---

## Playlist CLI Commands

```bash
# List all tracks in queue
radio playlist list

# Show currently playing track
radio playlist current

# Skip to next track
radio playlist next

# Force library directory rescan
radio playlist scan
```
