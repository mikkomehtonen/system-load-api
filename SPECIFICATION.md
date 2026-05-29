# System Load API — Specification

## 1. Overview

The System Load API is a lightweight Go HTTP service that exposes real-time system resource metrics as JSON. It collects CPU, memory, disk, GPU, network, and host statistics and serves them through RESTful endpoints. The service uses only the Go standard library for HTTP routing and relies on `gopsutil/v3` and `nvidia-smi` for system metrics. An embedded dark terminal dashboard UI is served at `/` for browser-based monitoring.

### 1.1 Goals

- Provide a single binary with no external service dependencies
- Return structured JSON responses with consistent timestamping
- Support both aggregate and per-metric endpoint queries
- Gracefully degrade when optional hardware (GPU) is unavailable
- Operate with a 10-second request timeout to accommodate delta-sampled metrics
- Include an embedded dashboard UI for browser-based monitoring

### 1.2 Non-Goals

- Authentication or authorization
- Historical metric storage or time-series data
- Configuration files or complex setup
- Framework-based routing

---

## 2. System Requirements

### 2.1 Runtime

- Go 1.25+
- Linux, macOS, or Windows (gopsutil cross-platform support)
- `nvidia-smi` CLI (optional, for GPU metrics only)

### 2.2 Dependencies

| Dependency | Version | Purpose |
|---|---|---|
| `github.com/shirou/gopsutil/v3` | v3.24.5 | CPU, memory, disk, network metrics |
| `golang.org/x/sync` | v0.20.0 | `errgroup` for concurrent collection |

### 2.3 Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `PORT` | No | `8080` | HTTP listen port |

---

## 3. Architecture

### 3.1 Package Layout

```
main.go              # Entry point, server setup, signal handling, routing
static/              # Embedded dashboard UI
  index.html         # Single-page dashboard (dark terminal theme, 5s auto-refresh)
collectors/          # Metric collection functions (one per resource type)
  cpu.go             # CPU load averages, usage %, per-core usage
  memory.go          # RAM and swap usage
  disk.go            # Partition usage + I/O rates (1s delta)
  gpu.go             # nvidia-smi parsing (utilization, VRAM, temperature, fan speed)
  network.go         # Per-interface byte rates (1s delta)
  host.go            # Host-level system info (uptime)
  rounding.go        # Shared rounding helpers
handlers/            # HTTP handlers and middleware
  handlers.go        # All endpoint handlers + TimeoutMiddleware
models/              # Shared response data structures
  models.go          # All response structs + timestamp helper
```

### 3.2 Data Flow

```
Client Request
    │
    ▼
TimeoutMiddleware (10s deadline)
    │
    ▼
Handler (individual or Stats)
    │
    ├──► CollectCPU()    ──► gopsutil/load, gopsutil/cpu
    ├──► CollectMemory() ──► gopsutil/mem
    ├──► CollectDisk()   ──► gopsutil/disk (1s delta for I/O)
    ├──► CollectGPU()    ──► exec.Command("nvidia-smi")
    ├──► CollectNetwork()──► gopsutil/net (1s delta)
    └──► CollectHost()   ──► gopsutil/host
    │
    ▼
JSON Serialization → Response
```

### 3.3 Concurrency Model

The `/api/v1/stats` endpoint uses `errgroup.WithContext()` to run all six collectors concurrently. Partial failures are tolerated — if at least one collector succeeds, a 200 response is returned with available metrics and `null` for failed sections. Only when **all** collectors fail does the endpoint return HTTP 500.

The GPU collector never propagates errors to the handler; it always returns a `GPUStats` struct with `Available: false` and an error message on failure.

---

## 4. API Specification

### 4.1 General Conventions

- All endpoints use `GET` method
- All responses have `Content-Type: application/json`
- All responses include a top-level `"timestamp"` field in RFC 3339 format (UTC)
- Errors return HTTP 500 with `{"error": "message"}`
- Timeout returns HTTP 504 with `{"error": "request timeout"}`

### 4.2 Endpoints

#### 4.2.1 Health Check

```
GET /health
```

**Response (200):**
```json
{"status": "ok"}
```

#### 4.2.2 All Metrics

```
GET /api/v1/stats
```

**Latency:** ~1s (due to network and disk delta sampling)

**Response (200):**
```json
{
  "timestamp": "2026-05-25T10:30:00Z",
  "cpu": { ... },
  "memory": { ... },
  "disk": { ... },
  "gpu": { ... },
  "network": { ... },
  "host": { ... }
}
```

Sections are omitted (`null`/omitted via `omitempty`) if their collector fails.

**Response (500) — all collectors failed:**
```json
{"error": "all collectors failed: <error1>; <error2>; ..."}
```

#### 4.2.3 CPU Metrics

```
GET /api/v1/cpu
```

**Response (200):**
```json
{
  "timestamp": "2026-05-25T10:30:00Z",
  "cpu": {
    "load_avg_1min": 1.23,
    "load_avg_5min": 0.98,
    "load_avg_15min": 0.76,
    "usage_percent": 34.5,
    "core_count": 8,
    "per_core_percent": [45.2, 30.1, 28.7, 52.0, 20.3, 15.8, 40.1, 44.9]
  }
}
```

**Fields:**

| Field | Type | Description |
|---|---|---|
| `load_avg_1min` | float64 | 1-minute load average |
| `load_avg_5min` | float64 | 5-minute load average |
| `load_avg_15min` | float64 | 15-minute load average |
| `usage_percent` | float64 | Overall CPU usage percentage (rounded to 1 decimal) |
| `core_count` | int | Number of logical CPU cores |
| `per_core_percent` | []float64 | Per-core usage percentages (rounded to 1 decimal) |

#### 4.2.4 Memory Metrics

```
GET /api/v1/memory
```

**Response (200):**
```json
{
  "timestamp": "2026-05-25T10:30:00Z",
  "memory": {
    "total_gb": 15.88,
    "used_gb": 9.42,
    "available_gb": 6.46,
    "usage_percent": 59.3,
    "swap_total_gb": 8.0,
    "swap_used_gb": 0.5,
    "swap_usage_percent": 6.25
  }
}
```

**Fields:**

| Field | Type | Description |
|---|---|---|
| `total_gb` | float64 | Total physical RAM in GB (rounded to 2 decimals) |
| `used_gb` | float64 | Used RAM in GB (rounded to 2 decimals) |
| `available_gb` | float64 | Available RAM in GB (rounded to 2 decimals) |
| `usage_percent` | float64 | RAM usage percentage (rounded to 1 decimal) |
| `swap_total_gb` | float64 | Total swap in GB (rounded to 2 decimals) |
| `swap_used_gb` | float64 | Used swap in GB (rounded to 2 decimals) |
| `swap_usage_percent` | float64 | Swap usage percentage (rounded to 1 decimal) |

#### 4.2.5 Disk Metrics

```
GET /api/v1/disk
```

**Latency:** ~1s (due to I/O delta sampling)

**Response (200):**
```json
{
  "timestamp": "2026-05-25T10:30:00Z",
  "disk": {
    "partitions": [
      {
        "device": "/dev/sda1",
        "mountpoint": "/",
        "fstype": "ext4",
        "total_gb": 500.0,
        "used_gb": 230.5,
        "usage_percent": 46.1
      }
    ],
    "io": {
      "read_bytes_sec": 1048576,
      "write_bytes_sec": 524288
    }
  }
}
```

**Fields — partitions[]:**

| Field | Type | Description |
|---|---|---|
| `device` | string | Device path (e.g., `/dev/sda1`) |
| `mountpoint` | string | Mount point (e.g., `/`) |
| `fstype` | string | Filesystem type (e.g., `ext4`) |
| `total_gb` | float64 | Partition total size in GB (rounded to 2 decimals) |
| `used_gb` | float64 | Partition used space in GB (rounded to 2 decimals) |
| `usage_percent` | float64 | Partition usage percentage (rounded to 1 decimal) |

**Fields — io:**

| Field | Type | Description |
|---|---|---|
| `read_bytes_sec` | uint64 | Aggregate read throughput in bytes/sec (1s delta) |
| `write_bytes_sec` | uint64 | Aggregate write throughput in bytes/sec (1s delta) |

**Notes:**
- Inaccessible partitions are silently skipped
- `/dev/loop*` devices (snap squashfs mounts) are filtered out
- If I/O counters are unavailable, `io` is `null` but `partitions` is still returned

#### 4.2.6 GPU Metrics

```
GET /api/v1/gpu
```

**Response (200) — GPU available:**
```json
{
  "timestamp": "2026-05-25T10:30:00Z",
  "gpu": {
    "available": true,
    "devices": [
      {
        "index": 0,
        "name": "NVIDIA GeForce RTX 4090",
        "utilization_percent": 67.0,
        "memory_total_mb": 24576,
        "memory_used_mb": 8192,
        "memory_usage_percent": 33.3,
        "temperature_c": 72,
        "fan_speed_percent": 60
      }
    ]
  }
}
```

**Response (200) — GPU unavailable:**
```json
{
  "timestamp": "2026-05-25T10:30:00Z",
  "gpu": {
    "available": false,
    "devices": null,
    "error": "nvidia-smi not found"
  }
}
```

**Fields — devices[]:**

| Field | Type | Description |
|---|---|---|
| `index` | int | GPU device index |
| `name` | string | GPU model name |
| `utilization_percent` | float64 | GPU core utilization % (rounded to 1 decimal) |
| `memory_total_mb` | float64 | Total VRAM in MB (rounded to 1 decimal) |
| `memory_used_mb` | float64 | Used VRAM in MB (rounded to 1 decimal) |
| `memory_usage_percent` | float64 | VRAM usage % (computed, rounded to 1 decimal) |
| `temperature_c` | float64 | GPU temperature in Celsius (rounded to 1 decimal) |
| `fan_speed_percent` | float64 | Fan speed as a percentage 0–100 (rounded to 1 decimal) |

**Notes:**
- This endpoint **never** returns HTTP 500
- GPU metrics are sourced from `nvidia-smi` CLI with CSV output parsing
- `memory_usage_percent` is computed as `(memory_used_mb / memory_total_mb) * 100`
- `fan_speed_percent` is queried via `nvidia-smi --query-gpu=fan.speed`; if the value is unavailable or `N/A`, it defaults to `0`

#### 4.2.7 Network Metrics

```
GET /api/v1/network
```

**Latency:** ~1s (due to delta sampling)

**Response (200):**
```json
{
  "timestamp": "2026-05-25T10:30:00Z",
  "network": {
    "interfaces": [
      {
        "name": "eth0",
        "bytes_sent_sec": 524288,
        "bytes_recv_sec": 2097152
      }
    ]
  }
}
```

**Fields — interfaces[]:**

| Field | Type | Description |
|---|---|---|
| `name` | string | Network interface name |
| `bytes_sent_sec` | uint64 | Outbound bytes per second (1s delta) |
| `bytes_recv_sec` | uint64 | Inbound bytes per second (1s delta) |

#### 4.2.8 Host Metrics

```
GET /api/v1/host
```

**Latency:** Instant

**Response (200):**
```json
{
  "timestamp": "2026-05-25T10:30:00Z",
  "host": {
    "uptime_seconds": 3600
  }
}
```

**Fields:**

| Field | Type | Description |
|---|---|---|
| `uptime_seconds` | uint64 | System uptime in seconds |

---

## 5. Dashboard UI

### 5.1 Overview

The API binary embeds a single-page dashboard UI served at `/`. The UI is a vanilla HTML/CSS/JS file with no external dependencies, embedded via Go's `embed.FS`.

### 5.2 Access

- **URL:** `http://<host>:<port>/` (e.g., `http://localhost:8080/`)
- **Refresh:** Auto-polls `/api/v1/stats` every 5 seconds
- **Theme:** Dark terminal style, monospace font, oklch color palette

### 5.3 Sections

| Section | Content |
|---|---|
| CPU | Overall usage %, load averages (1m/5m/15m), per-core progress bars |
| Memory | Total/used/available RAM, swap usage, progress bars |
| Disk | Partition table (device, mount, type, size, usage bar), I/O read/write rates |
| GPU | Per-device utilization, VRAM usage, temperature, fan speed bars; graceful "unavailable" state |
| Network | Per-interface RX/TX byte rates |
| Host | Uptime displayed in the header |

### 5.4 Color Thresholds

| Level | Threshold | Color |
|---|---|---|
| Normal | < 75% | Green (accent) |
| Warning | 75–89% | Orange |
| Error | ≥ 90% | Red |

### 5.5 Technical Details

- Single `static/index.html` file, embedded at build time
- `fs.Sub(staticFiles, "static")` strips the `static/` prefix for serving at `/`
- API routes (`/api/v1/*`, `/health`) take precedence over the file server (Go `ServeMux` most-specific-first matching)
- Responsive layout collapses to single column below 640px

---

## 6. Implementation Details

### 6.1 Rounding

All floating-point values are rounded before serialization:

| Function | Precision | Used For |
|---|---|---|
| `roundTo1(v)` | 1 decimal place | Percentages, temperatures, GPU metrics |
| `roundTo2(v)` | 2 decimal places | GB values (memory, disk) |
| `roundSlice(s)` | 1 decimal place | Per-core CPU usage array |

Rounding uses the formula: `float64(int(v*10^n + 0.5)) / 10^n`

### 6.2 Delta Sampling

Network and disk I/O metrics require two snapshots separated by 1 second:

```
T0: snapshot counters
Sleep 1s
T1: snapshot counters
Delta = T1 - T1 (per interface or aggregate)
```

This means any endpoint returning network or disk I/O stats has a minimum latency of ~1 second.

### 6.3 Timeout Handling

A custom `TimeoutMiddleware` wraps all routes:

- Request context deadline: **10 seconds**
- On timeout: returns HTTP 504 with `{"error": "request timeout"}`
- Server `WriteTimeout`: 12 seconds (allows middleware to respond)
- Server `ReadHeaderTimeout`: 5 seconds
- Server `IdleTimeout`: 60 seconds
- Server `MaxHeaderBytes`: 1 MB

### 6.4 Graceful Shutdown

- Listens for `SIGINT` and `SIGTERM`
- Shutdown context timeout: 10 seconds
- In-flight requests are allowed to complete within the drain period

### 6.5 Error Handling Strategy

| Collector | Error Behavior |
|---|---|
| CPU | Propagates error to handler → HTTP 500 |
| Memory | Propagates error to handler → HTTP 500 |
| Disk | Propagates error to handler → HTTP 500 (partitions may be partial) |
| GPU | **Never propagates** — returns `Available: false` with error message |
| Network | Propagates error to handler → HTTP 500 |
| Host | Propagates error to handler → HTTP 500 |

In `/api/v1/stats`, partial failures yield 200 with `null` sections. Only total failure yields 500.

---

## 7. Testing

### 7.1 Test Categories

| Category | Scope | Flag |
|---|---|---|
| Unit tests | Pure functions (`roundTo1`, `roundTo2`, `roundSlice`, `parseNvidiaSMI`, `Now()`) | Always run |
| Integration tests | Real system collectors, HTTP endpoint responses | Skipped with `-short` |

### 7.2 Test Commands

```bash
go test ./...              # Full suite (~3s, includes delta sampling)
go test -short ./...       # Fast unit tests only (~instant)
go test ./collectors/      # Single package tests
go vet ./...               # Static analysis
go fmt ./...               # Code formatting
```

### 7.3 Test Coverage

| Package | Coverage Areas |
|---|---|
| `collectors` | Rounding edge cases, nvidia-smi CSV parsing, collector integration calls |
| `handlers` | All 8 endpoints via `httptest`, JSON helpers, error aggregation |
| `models` | RFC 3339/UTC timestamp validation, struct construction |

---

## 8. Endpoint Summary

| Method | Path | Latency | GPU Required | Description |
|---|---|---|---|---|
| GET | `/health` | Instant | No | Health check |
| GET | `/api/v1/stats` | ~1s | No | All metrics (concurrent) |
| GET | `/api/v1/cpu` | Instant | No | CPU load and usage |
| GET | `/api/v1/memory` | Instant | No | RAM and swap usage |
| GET | `/api/v1/disk` | ~1s | No | Partition usage + I/O rates |
| GET | `/api/v1/gpu` | Instant | Optional | GPU metrics (graceful fallback) |
| GET | `/api/v1/network` | ~1s | No | Per-interface byte rates |
| GET | `/api/v1/host` | Instant | No | Host uptime |

---

## 9. Response Schema Index

### 9.1 Top-Level Response Types

| Endpoint | Response Type |
|---|---|
| `/health` | `{"status": "ok"}` |
| `/api/v1/stats` | `StatsResponse` |
| `/api/v1/cpu` | `CPUResponse` |
| `/api/v1/memory` | `MemoryResponse` |
| `/api/v1/disk` | `DiskResponse` |
| `/api/v1/gpu` | `GPUResponse` |
| `/api/v1/network` | `NetworkResponse` |
| `/api/v1/host` | `HostResponse` |
| Any (error) | `ErrorResponse` |

### 9.2 Nested Types

| Type | Parent | Description |
|---|---|---|
| `CPUStats` | `CPUResponse`, `StatsResponse` | CPU metrics |
| `MemoryStats` | `MemoryResponse`, `StatsResponse` | Memory metrics |
| `DiskStats` | `DiskResponse`, `StatsResponse` | Disk metrics |
| `DiskPartition` | `DiskStats` | Per-partition info |
| `DiskIO` | `DiskStats` | Aggregate I/O rates |
| `GPUStats` | `GPUResponse`, `StatsResponse` | GPU metrics |
| `GPUDevice` | `GPUStats` | Per-GPU info |
| `NetworkStats` | `NetworkResponse`, `StatsResponse` | Network metrics |
| `NetworkInterface` | `NetworkStats` | Per-interface rates |
| `HostStats` | `HostResponse`, `StatsResponse` | Host metrics |
| `HostResponse` | — | Host endpoint response |
