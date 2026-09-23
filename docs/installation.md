# Installation Guide

Aircoda is distributed as a single standalone executable binary for Linux, Windows, and macOS. Normal users do not need Go installed to run pre-compiled release binaries.

---

## 1. Downloading Pre-Built Binaries

Download the appropriate binary for your operating system and architecture from the official GitHub Releases page:

- **Linux**: `aircoda_v0.1.0_linux_amd64.tar.gz` (or `linux_arm64`)
- **Windows**: `aircoda_v0.1.0_windows_amd64.zip` (or `windows_arm64`)
- **macOS**: `aircoda_v0.1.0_macos_arm64.tar.gz` (Apple Silicon) or `macos_amd64.tar.gz` (Intel)

---

## 2. Linux Installation

```bash
tar -xzf aircoda_v0.1.0_linux_amd64.tar.gz
cd aircoda_v0.1.0_linux_amd64
sudo install ./radio /usr/local/bin/radio
```

Verify installation:
```bash
radio version
```

---

## 3. Windows Installation

1. Extract the downloaded `aircoda_v0.1.0_windows_amd64.zip` package.
2. Move `radio.exe` to a folder in your System PATH (e.g. `C:\Windows\System32\` or `C:\Program Files\Aircoda\`).

Verify installation in Command Prompt or PowerShell:
```powershell
radio version
```

---

## 4. macOS Installation

```bash
tar -xzf aircoda_v0.1.0_macos_arm64.tar.gz
cd aircoda_v0.1.0_macos_arm64
sudo install ./radio /usr/local/bin/radio
```

Verify installation:
```bash
radio version
```

---

## 5. Building from Source

If you prefer building from source, ensure you have Go 1.22+ installed:

```bash
git clone https://github.com/aircoda/aircoda.git
cd aircoda
go build -o radio ./cmd/radio
```
