# Contributing to Aircoda

Thank you for your interest in contributing to Aircoda! We welcome contributions, bug reports, feature proposals, and documentation enhancements.

---

## Code Guidelines & Standards

1. **Go Code Style**: Follow standard Go formatting conventions (`gofmt`, `go vet`).
2. **Cross-Platform Compatibility**: Ensure code works natively across Linux, Windows, and macOS. Avoid shell-specific commands (`sh`, `bash`, `grep`, `sed`).
3. **No Emojis**: Never use emojis in CLI output, logs, UI, or documentation.
4. **Data Safety**: Never write code that deletes, modifies, or renames user audio files.

---

## Submitting Pull Requests

1. Fork the repository and create a feature branch (`git checkout -b feature/my-feature`).
2. Run tests to ensure everything passes (`go test -v ./...`).
3. Commit your changes cleanly.
4. Push to your fork and submit a Pull Request.
