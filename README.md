# System Load API

A Go HTTP API that returns real-time system load information as JSON, including CPU (with temperature), memory, disk, GPU, and network metrics. Includes a built-in dark terminal dashboard UI.

## Quick Start

```bash
go build -o sysload .
./sysload
```

The server starts on `:8080` by default. Set the `PORT` environment variable to change it:

```bash
PORT=9090 ./sysload
```

## Dashboard UI

Open `http://localhost:8080` in a browser to see the real-time dashboard. It auto-refreshes every 5 seconds and displays:

- **CPU** — overall usage, load averages, per-core bars, package temperature
- **Memory** — RAM and swap usage with progress bars
- **Disk** — partition table with usage bars, I/O read/write rates
- **GPU** — utilization, VRAM, temperature, fan speed per device (graceful fallback if unavailable)
- **Network** — per-interface RX/TX byte rates (virtual interfaces filtered: lo, docker0, br-*, veth*)
- **Host** — system uptime (human-readable header display)

The UI is a single vanilla HTML file embedded in the binary — no external assets or build step needed.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/api/v1/stats` | All metrics in one response |
| GET | `/api/v1/cpu` | CPU metrics only |
| GET | `/api/v1/memory` | Memory metrics only |
| GET | `/api/v1/disk` | Disk metrics only |
| GET | `/api/v1/gpu` | GPU metrics only |
| GET | `/api/v1/network` | Network metrics only |
| GET | `/api/v1/host` | Host info (uptime) only |

## Response Examples

### GET /api/v1/stats

```json
{
  "timestamp": "2026-05-25T10:30:00Z",
  "cpu": {
    "load_avg_1min": 1.23,
    "load_avg_5min": 0.98,
    "load_avg_15min": 0.76,
    "usage_percent": 34.5,
    "core_count": 8,
    "per_core_percent": [45.2, 30.1, 28.7, 52.0, 20.3, 15.8, 40.1, 44.9],
    "temperature_c": 48.0
  },
  "memory": {
    "total_gb": 15.88,
    "used_gb": 9.42,
    "available_gb": 6.46,
    "usage_percent": 59.3,
    "swap_total_gb": 8.0,
    "swap_used_gb": 0.5,
    "swap_usage_percent": 6.25
  },
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
  },
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
  },
  "network": {
    "interfaces": [
      {
        "name": "eth0",
        "bytes_sent_sec": 524288,
        "bytes_recv_sec": 2097152
      }
    ]
  },
  "host": {
    "uptime_seconds": 3600
  }
}
```

### Individual Endpoints

Each individual endpoint returns only its section with a `timestamp` field:

```json
{
  "timestamp": "2026-05-25T10:30:00Z",
  "cpu": { ... }
}
```

### GPU Unavailable

When `nvidia-smi` is not present or fails, the GPU section gracefully degrades:

```json
{
  "timestamp": "...",
  "gpu": {
    "available": false,
    "devices": null,
    "error": "nvidia-smi not found"
  }
}
```

### Errors

On failure, the API returns HTTP 500:

```json
{"error": "message"}
```

## Latency Notes

- **Network and disk I/O rates** require 1-second delta sampling (snapshot at T0 and T1, compute delta). Endpoints returning these stats (`/network`, `/disk`, `/stats`) have ~1s latency.
- **GPU** metrics require `nvidia-smi` on the host. If unavailable, the endpoint still returns immediately with `available: false`.
- **CPU temperature** is gathered from system sensors via `gopsutil`. Prefers package-level readings, falls back to die then highest core temp. Returns `null` when sensors are unavailable.

## Network Interface Filtering

The network collector excludes virtual and loopback interfaces to keep metrics focused on physical networks. Filtered interfaces:

- `lo` — loopback
- `docker0` — Docker bridge
- `br-*` — bridge interfaces
- `veth*` — virtual ethernet pairs (containers)

## Development

```bash
go build ./...          # build all packages
go test ./...           # run tests (full suite, includes ~1s delta sampling)
go test -short ./...    # run fast unit tests only (skips integration/slow tests)
go test ./collectors/   # test a single package
go vet ./...            # static analysis
go fmt ./...            # format code
```

## Testing

The test suite covers unit tests for pure functions and parsing logic, plus integration tests that call the real system collectors and HTTP handlers.

### Test Layout

| Package | Focus |
|---------|-------|
| `collectors` | `roundTo1`/`roundTo2`/`roundSlice` edge cases, `parseNvidiaSMI` with mock CSV input, `CollectCPU`/`CollectMemory`/`CollectDisk`/`CollectNetwork`/`CollectGPU`/`CollectHost` integration |
| `handlers` | All 8 endpoints via `httptest`, `collectErrors`, `writeJSON`/`writeError` helpers |
| `models` | `Now()` RFC 3339/UTC validation, struct field construction |

### Short Mode

Integration tests that involve 1-second delta sampling (network, disk I/O) or real system calls are skipped with `-short`:

```bash
go test -short -v ./...   # fast unit-only (~instant)
go test -v ./...           # full suite (~3s due to delta sampling)
```

## Dependencies

- [gopsutil/v3](https://github.com/shirou/gopsutil) — CPU, memory, disk, and network metrics
- `nvidia-smi` CLI — GPU metrics (optional; graceful fallback if missing)
- Go stdlib `net/http` — HTTP routing (no framework)

## Graceful Shutdown

The server shuts down cleanly on `SIGINT` or `SIGTERM` with a 10-second drain timeout.
