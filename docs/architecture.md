# Architecture & System Design

Aircoda is designed around modular, decoupled components to ensure stability, cross-platform portability, and crash resilience.

---

## High-Level Component Diagram

```text
                    AIRCODA
                       │
                 ┌─────▼─────┐
                 │ radio CLI │
                 └─────┬─────┘
                       │
                 ┌─────▼─────┐
                 │  Aircoda  │
                 │   Core    │
                 └─────┬─────┘
                       │
          ┌────────────┼────────────┐
          ▼            ▼            ▼
      Playlist      FFmpeg       Icecast
          │            │            │
          └────────────┼────────────┘
                       ▼
                 LIVE STREAM
                       │
          ┌────────────┴────────────┐
          ▼                         ▼
      LOCAL STREAM            PUBLIC ACCESS
 (http://127.0.0.1:8000)  (Direct IP / Cloudflare)
```

---

## Core Layers

1. **CLI Layer (`internal/cli`)**: Operates via Cobra CLI framework. Provides setup, commands, menu, doctor, status, and control actions.
2. **Configuration & Path Layer (`internal/config`)**: Manages TOML parsing, validation, atomic writing, backup/restore, and platform-aware directory resolution.
3. **Audio Library & Queue Engine (`internal/library`, `internal/playlist`)**: Recursively discovers audio files and maintains persistent queue state (`queue.json`).
4. **Streaming Engine (`internal/audio`)**: Manages FFmpeg background worker loop, Icecast XML generator & process runner, and embedded stream hub fallback.
5. **Platform Service Layer (`internal/service`)**: Wraps native OS background services (`systemd` on Linux, `Windows Service` on Windows, `launchd` on macOS).
6. **Public Access & Tunnel Layer (`internal/public`, `internal/tunnel`)**: Resolves listener URLs and manages Cloudflare Tunnel lifecycle.
