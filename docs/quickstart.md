# Quick Start Guide

Get your self-hosted 24/7 internet radio station broadcasting in less than 2 minutes.

---

## Step 1: Initialize Setup

Run the `radio` command from any terminal prompt:

```bash
radio
```

On first run, an interactive wizard guides you through basic station setup:
- Station Name (e.g. `My Indie Radio`)
- Station ID (e.g. `indie-radio`)
- Audio Folder (Path to your local music directory containing `.mp3`, `.flac`, `.wav`, etc.)
- Playback Mode (`Shuffle` or `Sequential`)

---

## Step 2: Add Audio Files

Add your audio tracks into the configured music folder. Scan the library anytime using:

```bash
radio playlist scan
```

---

## Step 3: Start Radio Station

Start your station in continuous background streaming mode:

```bash
radio start
```

Your station audio stream is immediately live:
- Local Stream URL: `http://127.0.0.1:8000/indie-radio`

---

## Step 4: Check Station Status

Inspect active playback and server status anytime:

```bash
radio status
```

---

## Step 5: Stop Radio Station

To stop your station cleanly:

```bash
radio stop
```
