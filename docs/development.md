# Development Guide

Guide for developers contributing to or extending Aircoda.

---

## 1. Repository Layout

```text
Aircoda/
├── cmd/
│   ├── radio/               # CLI main entrypoint
│   └── icecast_mock/        # Mock Icecast server for testing
├── internal/
│   ├── audio/               # FFmpeg stream worker & Icecast process runner
│   ├── cli/                 # Cobra CLI subcommands & menu UX
│   ├── config/              # TOML config schema, paths, atomic saver
│   ├── deps/                # External binary detector
│   ├── library/             # Recursive audio library scanner
│   ├── logger/              # Log formatting & file appender
│   ├── playlist/            # Queue state machine
│   ├── public/              # Stream URL builder & IP detector
│   ├── service/             # Cross-platform kardianos/service manager
│   ├── system/              # Version, OS, and sysinfo metadata
│   └── tunnel/              # Cloudflare tunnel manager
├── scripts/
│   ├── build.sh             # Cross-compilation release script
│   └── release.sh           # Package archives & SHA256SUMS generator
└── tests/                   # Unit & integration test suites
```

---

## 2. Local Setup & Building

```bash
git clone https://github.com/aircoda/aircoda.git
cd aircoda
go build -o radio ./cmd/radio
```

---

## 3. Running Unit Tests

```bash
go test -v ./...
```

---

## 4. Testing Cross-Compilation

```bash
./scripts/build.sh
```
