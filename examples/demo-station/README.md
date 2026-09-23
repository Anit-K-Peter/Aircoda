# Official Aircoda Demo Radio Station

This directory contains a complete, deployable example configuration for hosting an official public Aircoda demo radio station.

An official demo station uses the exact same unmodified Aircoda open-source engine as any personal installation.

---

## Deploying the Demo Station

### 1. Copy Configuration
Copy `config.example.toml` to your system configuration location:

**Linux / macOS**:
```bash
mkdir -p ~/.config/Aircoda
cp config.example.toml ~/.config/Aircoda/config.toml
```

**Windows**:
```powershell
New-Item -ItemType Directory -Path "$env:APPDATA\Aircoda" -Force
Copy-Item config.example.toml "$env:APPDATA\Aircoda\config.toml"
```

---

### 2. Populate Audio Files
Place your MP3, OGG, WAV, or FLAC audio files into the `music/` directory (or update `[audio].directory` in `config.toml` to point to your media folder).

---

### 3. Start Radio Station

```bash
radio start
```

---

### 4. Enable Public Access

**Option A: Direct Public Access (Port 8000)**
```bash
radio public setup
```

**Option B: Cloudflare Tunnel (HTTPS)**
```bash
radio tunnel setup
radio tunnel start
```

---

## Verification

Check station status and stream URL anytime with:

```bash
radio status
radio doctor
```
