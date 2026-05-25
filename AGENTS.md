# AGENTS.md

## Project

Go HTTP API returning real-time system load as JSON (CPU, memory, disk, GPU, network). Single binary, no framework.

## Key Commands

```bash
go build ./...          # build all
go test ./...           # test all
go test ./collectors/   # test single package
go fmt ./...            # format code
PORT=9090 go run main.go  # run with custom port (default :8080)
```

## Architecture

- **stdlib only** for HTTP routing (`net/http`). Do not add a router framework.
- **`gopsutil/v3`** (not v4) for CPU/memory/disk/network metrics.
- **`nvidia-smi` CLI** via `exec.Command` for GPU — parse CSV output with `--query-gpu=... --format=csv,noheader,nounits`.
- **`errgroup`** for concurrent collector calls in the `/api/v1/stats` handler.

## Critical Gotchas

- **Network rates require 1s delta sampling** (snapshot at T0 and T1, compute delta). This adds ~1s latency to any endpoint returning network stats. Do not try to make it instant.
- **GPU collector must never fail the whole request.** If `nvidia-smi` is missing or errors, return `{"gpu": {"available": false, "devices": null, "error": "..."}}` — no error propagated.
- **Request timeout is 10s** (covers the 1s network sampling). Do not reduce below 2s.
- **Port via `PORT` env var**, not a flag.

## Package Layout

```
main.go              # Entry point, server, routing, graceful shutdown
collectors/          # One file per metric type, each exposes a single Collect*() function
handlers/            # HTTP handlers, JSON serialization
models/              # Shared response structs
```

## Response Conventions

- All responses include a top-level `"timestamp"` field (RFC 3339).
- Errors: HTTP 500 with `{"error": "message"}`.
- Individual endpoints (`/cpu`, `/memory`, etc.) return only their section + timestamp.
- `/api/v1/stats` returns all sections concurrently collected.
