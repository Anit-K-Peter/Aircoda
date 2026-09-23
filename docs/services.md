# OS Background Service Integration

Aircoda supports native operating system background services across Linux, Windows, and macOS.

---

## Native Backend Mapping

- **Linux**: `systemd` (`/etc/systemd/system/aircoda.service`)
- **Windows**: `Windows Service`
- **macOS**: `launchd` (`~/Library/LaunchAgents/com.aircoda.radio.plist`)

---

## Service Management Commands

```bash
# Install system background service
radio service install

# Uninstall system background service
radio service uninstall

# Start background service
radio service start

# Stop background service
radio service stop

# Restart background service
radio service restart

# Check background service status
radio service status

# Enable start-on-boot
radio service enable

# Disable start-on-boot
radio service disable
```
