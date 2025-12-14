# Dideban Agent 👁️‍🗨️

**Lightweight Resource Monitoring Agent for Dideban**

Dideban Agent is a minimal, fast, and secure system monitoring agent written in Go.
It is designed to collect **host-level resource metrics** and send them to the Dideban Core engine.

The agent follows the same philosophy as Dideban:

> **Low resource usage · Predictable behavior · Private by default**

---

## ✨ Features (v0.1 – MVP)

* 🖥️ CPU usage & load average
* 🧠 Memory usage
* 💽 Disk usage
* ⏱️ Metric collection duration (latency)
* 🔁 Periodic metric collection
* 📤 Push-based delivery to Dideban Core (HTTP)
* 📦 Single static Go binary
* 🔐 Token-based authentication

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

* Linux (amd64 / arm64)
* No root required (except for advanced metrics in future versions)

---

### Run Agent

```bash
./dideban-agent --config /etc/dideban-agent/config.yaml
```

The agent will:

1. Load configuration
2. Collect system metrics
3. Measure collection duration
4. Push metrics to Dideban Core
5. Sleep until next interval

---

## ⚙️ Configuration

The agent is configured using a **YAML file**.

Example:

```yaml
agent:
  id: "vpc-node-01"
  interval: 10s

core:
  endpoint: "https://dideban.internal/api/metrics"
  token: "AGENT_SECRET_TOKEN"
```

### Configuration Notes

* `agent.id` must be unique per host
* `interval` defines metric collection frequency
* `token` is used for authentication with Dideban Core

---

## 📊 Metrics Payload (Example)

```json
{
  "agent_id": "vpc-node-01",
  "timestamp": 1734000000,
  "collect_duration_ms": 14,
  "cpu": {
    "usage_percent": 37.2,
    "load_1": 0.64
  },
  "memory": {
    "used_mb": 2048,
    "total_mb": 8192,
    "usage_percent": 25
  },
  "disk": {
    "used_gb": 120,
    "total_gb": 250,
    "usage_percent": 48
  }
}
```

---

## 🔐 Security Model

* Push-only communication
* Static token authentication (MVP)
* No inbound ports required
* TLS recommended

Future plans:

* Token rotation
* mTLS

---

## 📦 Deployment

### Binary (Recommended)

Download the pre-built binary and run it directly on the host.

### Docker (Optional)

Docker support may be added later for containerized environments.

---

## 🛣️ Roadmap

### v0.1

* [ ] CPU / Memory / Disk metrics
* [ ] Metric latency tracking
* [ ] HTTP push

### v0.2

* [ ] Host metadata (OS, kernel, uptime)
* [ ] Process-level metrics
* [ ] Secure token rotation

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
