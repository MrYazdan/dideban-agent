# Dideban Agent 👁️‍🗨️

**Lightweight Resource Monitoring Agent for Dideban**

Dideban Agent is a minimal, fast, and secure system monitoring agent written in Go.
It is designed to collect **host-level resource metrics** and send them to the Dideban Core engine.

The agent follows the same philosophy as Dideban:

> **Low resource usage · Predictable behavior · Private by default**

---

## ✨ Features (v0.1 – MVP)

* 🖥️ **CPU metrics** - Usage percentage & load averages (1m, 5m, 15m)
* 🧠 **Memory metrics** - Used, total, available memory with usage percentage
* 💽 **Disk metrics** - Used, total disk space with usage percentage
* ⏱️ **Performance tracking** - Metric collection duration measurement
* 🔁 **Periodic collection** - Configurable interval-based metric gathering
* 📤 **HTTP delivery** - Push-based transmission with retry logic
* 🔧 **Dual mode operation** - Development (mock) and production (HTTP) modes
* 📦 **Single binary** - No external dependencies or runtime requirements
* 🔐 **Secure authentication** - Bearer token-based API authentication
* 🛡️ **Resilient design** - Exponential backoff, timeout handling, graceful shutdown

---

## 🎯 Design Goals

Dideban Agent is intentionally **simple and boring**:

* No UI
* No embedded database
* No runtime configuration changes
* No external dependencies

This ensures:

* Minimal attack surface
* Predictable performance
* Easy auditing
* Safe execution on production hosts

---

## 🧠 Architecture Overview

```
+-------------------+
|   Dideban Agent   |
|                   |
|  +-------------+  |
|  |  Collectors |  |
|  |  CPU / RAM  |  |
|  |  Disk       |  |
|  +------+------+  |
|         |         |
|  +------+------+  |
|  |  Metrics    |  |
|  |  Aggregator |  |
|  +------+------+  |
|         |         |
|  +------+------+  |
|  |  HTTP       |  |
|  |  Sender     |  |
|  +-------------+  |
+-------------------+
          |
          v
+-------------------+
| Dideban Core API  |
+-------------------+
```

---

## 🚀 Getting Started

### Requirements

* **Linux** (amd64 / arm64) or **Windows** (amd64)
* No root/administrator privileges required
* Go 1.21+ (for building from source)

### Quick Start

#### Linux/macOS:
```bash
# Copy and edit config
cp config.example.yaml config.yaml
vim config.yaml

# Run agent
./dideban-agent --config config.yaml
```

#### Windows:
```cmd
# Copy and edit config
copy config.example.yaml config.yaml
notepad config.yaml

# Run agent
dideban-agent.exe --config config.yaml
```

### Build from Source

#### Linux/macOS:
```bash
git clone https://github.com/MrYazdan/dideban-agent
cd dideban-agent
go build -o dideban-agent ./cmd/dideban-agent
```

#### Windows:
```cmd
git clone https://github.com/MrYazdan/dideban-agent
cd dideban-agent
go build -o dideban-agent.exe ./cmd/dideban-agent
```

### Agent Lifecycle

The agent follows this execution pattern:

1. **Initialization** - Load config, setup logging, initialize collectors
2. **Collection Loop** - Gather metrics from CPU, memory, disk collectors
3. **Transmission** - Send metrics via HTTP (prod) or mock (dev) sender
4. **Sleep** - Wait for next collection interval
5. **Graceful Shutdown** - Handle SIGINT/SIGQUIT/SIGTERM signals

---

## ⚙️ Configuration

The agent uses a **YAML configuration file** with comprehensive validation and defaults.

### Complete Configuration Example

```yaml
# Agent identification and behavior
agent:
  id: "vpc-node-01"          # Unique identifier (required)
  interval: 30s              # Collection interval (default: 30s)

# Dideban Core backend
core:
  endpoint: "https://dideban.internal/api/metrics"  # API endpoint (required)
  token: "AGENT_SECRET_TOKEN"                       # Auth token (required)

# HTTP sender configuration (optional)
sender:
  max_retries: 3             # Retry attempts (default: 3)
  initial_retry_delay: 1s    # Initial backoff (default: 1s)
  max_retry_delay: 30s       # Max backoff (default: 30s)
  request_timeout: 10s       # Request timeout (default: 10s)
  client_timeout: 30s        # Client timeout (default: 30s)

# Logging configuration
log:
  level: "info"              # debug, info, warn, error (default: info)
  pretty: true               # Console formatting (default: true)

# Application mode
mode: "production"           # development, production (default: development)
```

### Environment Variables

All configuration can be overridden using environment variables with `DIDEBAN_` prefix:

#### Linux/macOS:
```bash
export DIDEBAN_AGENT_ID="my-server"
export DIDEBAN_CORE_ENDPOINT="https://api.dideban.com/metrics"
export DIDEBAN_CORE_TOKEN="your-secret-token"
export DIDEBAN_MODE="production"
```

#### Windows:
```cmd
set DIDEBAN_AGENT_ID=my-server
set DIDEBAN_CORE_ENDPOINT=https://api.dideban.com/metrics
set DIDEBAN_CORE_TOKEN=your-secret-token
set DIDEBAN_MODE=production
```

### Configuration Notes

* **agent.id** - Must be unique per host (used for metric identification)
* **interval** - Supports Go duration format: `30s`, `1m`, `5m30s`
* **mode** - `development` uses mock sender, `production` uses HTTP sender
* **Config locations**:
  - Linux/macOS: `~/.dideban/agent/config.yaml`
  - Windows: `%APPDATA%\dideban\agent\config.yaml`
* **Environment variables** - Override YAML values using dot notation with underscores

---

## 📊 Metrics Payload

The agent collects and transmits metrics in JSON format:

```json
{
  "agent_id": "vpc-node-01",
  "timestamp_ms": 1734000000000,
  "collect_duration_ms": 14,
  "cpu": {
    "usage_percent": 37.2,
    "load_1": 0.64,
    "load_5": 0.52,
    "load_15": 0.48
  },
  "memory": {
    "used_mb": 2048,
    "total_mb": 8192,
    "usage_percent": 25,
    "available_mb": 6144
  },
  "disk": {
    "used_gb": 120,
    "total_gb": 250,
    "usage_percent": 48
  }
}
```

### Metric Details

| Field | Description | Unit |
|-------|-------------|------|
| `agent_id` | Unique agent identifier | string |
| `timestamp_ms` | Collection timestamp | milliseconds (Unix) |
| `collect_duration_ms` | Time taken to collect metrics | milliseconds |
| `cpu.usage_percent` | Overall CPU utilization | percentage (0-100) |
| `cpu.load_*` | System load averages | float |
| `memory.*_mb` | Memory statistics | megabytes |
| `disk.*_gb` | Disk space statistics | gigabytes |

---

## 🔐 Security Model

### Current Implementation (v0.1)

* **Push-only architecture** - Agent initiates all connections
* **Bearer token authentication** - Static token-based API auth
* **No inbound ports** - Zero attack surface from network
* **TLS support** - HTTPS endpoints recommended
* **Minimal privileges** - No root access required
* **Connection pooling** - Reuses HTTP connections securely

### Security Best Practices

1. **Use HTTPS endpoints** for production deployments
2. **Rotate tokens regularly** (manual process in v0.1)
3. **Run as non-root user** with minimal system permissions
4. **Monitor agent logs** for authentication failures
5. **Network isolation** - Restrict outbound connections to Dideban Core only

### Future Security Enhancements

* **Automatic token rotation** with refresh mechanism
* **mTLS support** for certificate-based authentication
* **Encrypted local storage** for sensitive configuration
* **Agent attestation** and integrity verification

---

## 📦 Deployment

### Linux Deployment (Systemd)

1. **Install binary:**
   ```bash
   sudo cp dideban-agent /usr/local/bin/
   sudo chmod +x /usr/local/bin/dideban-agent
   ```

2. **Create configuration directory:**
   ```bash
   sudo mkdir -p /etc/dideban-agent
   sudo cp config.example.yaml /etc/dideban-agent/config.yaml
   ```

3. **Create systemd service:**
   ```ini
   # /etc/systemd/system/dideban-agent.service
   [Unit]
   Description=Dideban Monitoring Agent
   After=network.target
   
   [Service]
   Type=simple
   ExecStart=/usr/local/bin/dideban-agent --config /etc/dideban-agent/config.yaml
   Restart=always
   RestartSec=10
   
   [Install]
   WantedBy=multi-user.target
   ```

4. **Enable and start:**
   ```bash
   sudo systemctl enable dideban-agent
   sudo systemctl start dideban-agent
   ```

### Windows Deployment (Service)

1. **Install as Windows Service** (using [NSSM](https://nssm.cc/) or similar):
   ```cmd
   # Using NSSM
   nssm install "Dideban Agent" "C:\Program Files\dideban-agent\dideban-agent.exe"
   nssm set "Dideban Agent" Arguments "--config C:\Program Files\dideban-agent\config.yaml"
   nssm start "Dideban Agent"
   ```

### Manual Deployment

#### Linux/macOS:
```bash
# Development mode
./dideban-agent --config config.yaml

# Production mode
DIDEBAN_MODE=production ./dideban-agent --config /etc/dideban-agent/config.yaml
```

#### Windows:
```cmd
# Development mode
dideban-agent.exe --config config.yaml

# Production mode
set DIDEBAN_MODE=production
dideban-agent.exe --config config.yaml
```

### Container Support

Docker and container deployment will be available in future versions.

---

## 🛣️ Roadmap

### v0.1 (Current) ✅

* ✅ **Core metrics collection** - CPU, Memory, Disk with concurrent gathering
* ✅ **HTTP transmission** - Production-ready sender with retry logic
* ✅ **Development mode** - Mock sender for testing and development
* ✅ **Configuration system** - YAML + environment variable support
* ✅ **Structured logging** - JSON and pretty console output
* ✅ **Graceful shutdown** - Signal handling and resource cleanup
* ✅ **Performance tracking** - Collection duration measurement

### v0.2 (Planned)

* [ ] **Host metadata** - OS version, kernel, uptime, hardware info
* [ ] **Network metrics** - Interface statistics, bandwidth usage
* [ ] **Process monitoring** - Top processes by CPU/memory usage
* [ ] **Custom metrics** - Plugin system for application-specific metrics
* [ ] **Health checks** - Agent self-monitoring and diagnostics

### v0.3 (Future)

* [ ] **Security enhancements** - Token rotation, mTLS support
* [ ] **Container support** - Docker metrics, Kubernetes integration
* [ ] **Advanced filtering** - Metric sampling and aggregation
* [ ] **Local buffering** - Offline operation and metric queuing
* [ ] **Configuration hot-reload** - Runtime configuration updates

---

## 🤝 Contributing

Contributions are welcome, but simplicity is sacred.

Please:

* Avoid unnecessary abstractions
* Keep the agent boring and predictable

---

## 📄 License

MIT License

---

## ❤️ Name Origin

**Dideban (دیدبان)** means *Watcher / Guardian* —
The Agent is the eye that observes each machine silently.
