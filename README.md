# System Load API

A Go HTTP API that returns real-time system load information as JSON, including CPU, memory, disk, GPU, and network metrics.

## Quick Start

```bash
go build -o sysload .
./sysload
```

The server starts on `:8080` by default. Set the `PORT` environment variable to change it:

```bash
PORT=9090 ./sysload
```

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
    "per_core_percent": [45.2, 30.1, 28.7, 52.0, 20.3, 15.8, 40.1, 44.9]
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
        "temperature_c": 72
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

## Development

```bash
go build ./...          # build all packages
go test ./...           # run tests
go test ./collectors/   # test a single package
go vet ./...            # static analysis
go fmt ./...            # format code
```

## Dependencies

- [gopsutil/v3](https://github.com/shirou/gopsutil) — CPU, memory, disk, and network metrics
- `nvidia-smi` CLI — GPU metrics (optional; graceful fallback if missing)
- Go stdlib `net/http` — HTTP routing (no framework)

## Graceful Shutdown

The server shuts down cleanly on `SIGINT` or `SIGTERM` with a 10-second drain timeout.
