# Public Access & Listener URLs

Aircoda allows station operators to expose their live radio streams to listeners over the internet while keeping management interfaces private.

---

## Access Modes

### 1. Local Mode (`local`)
- Stream accessible only on local machine (`http://127.0.0.1:8000/aircoda`).
- Private by default.

### 2. Direct Mode (`direct`)
- Stream accessible directly via public WAN IP or custom domain (`http://<public_ip>:8000/aircoda`).
- Requires opening/forwarding TCP port 8000 on your router/firewall.

### 3. Tunnel Mode (`tunnel`)
- Stream accessible securely over HTTPS via Cloudflare Tunnel (`https://radio.example.com/aircoda`).
- No open incoming router ports required.

---

## Public Access Commands

```bash
# Check public access status
radio public status

# Run interactive setup wizard
radio public setup

# Quick toggle mode
radio public enable
radio public disable
```
