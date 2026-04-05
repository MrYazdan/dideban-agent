# Dideban Agent — Improvement Roadmap

## Phase 1: Cross-Platform Compatibility (Windows + Linux)

### 1.1 CPU Load Average — Windows Fallback
- **Problem:** `load.Avg()` from `gopsutil` is not supported on Windows and returns an error.
- **Impact:** CPU collector fails on Windows, resulting in missing `load_1`, `load_5`, `load_15` fields.
- **Solution:** Detect OS at runtime; on Windows, set load fields to `0` (zero-value) and omit them from JSON output using `omitempty`. The receiver side already handles missing fields gracefully.
- **Files:** `internal/collector/cpu.go`
- **Status:** ✅ Done

---

## Phase 2: Configuration Cleanup

### 2.1 Remove `agent.name` Field
- **Problem:** `agent.name` is no longer needed in the agent config. The receiver identifies agents by token/endpoint, not by a name field.
- **Impact:** Removing it simplifies config, removes unnecessary validation, and cleans up startup logs.
- **Files:**
  - `internal/config/config.go` — remove `Name` field from `Agent` struct
  - `internal/config/defaults.go` — remove `agent.name` default and `getDefaultAgentName()`
  - `internal/config/validate.go` — remove `agent.name` validation
  - `config.example.yaml` — remove `agent.name` entry
  - `config.yaml` — remove `agent.name` entry
  - `cmd/dideban-agent/main.go` — remove `agent_name` from startup log
- **Status:** ✅ Done

---

## Phase 3: Logging Improvements

### 3.1 Allow `log.pretty` in Production
- **Problem:** `validate.go` explicitly rejects `log.pretty: true` in production mode, which is too restrictive.
- **Solution:** Remove the restriction. Pretty logging is a valid choice even in production (e.g., systemd journal, human-monitored deployments).
- **Files:** `internal/config/validate.go`
- **Status:** ✅ Done

---

## Phase 4: Future Improvements (Backlog)

### 4.1 Network Metrics Collector
- Collect bytes sent/received, packet counts per interface.
- Cross-platform via `gopsutil/net`.

### 4.2 Process Metrics Collector
- Top N processes by CPU/memory.
- Useful for anomaly detection on the receiver side.

### 4.3 Uptime Metric
- System uptime in seconds.
- Available cross-platform via `gopsutil/host`.

### 4.4 Configurable Disk Paths
- Allow multiple mount points to be monitored (not just `/` or `C:`).
- Config: `collector.disk.paths: ["/", "/data"]`

### 4.5 Agent Version in Payload
- Include agent version string in the metrics payload for receiver-side diagnostics.

### 4.6 Health Check Endpoint
- Expose a lightweight HTTP `/healthz` endpoint for orchestration systems (k8s, systemd watchdog).
