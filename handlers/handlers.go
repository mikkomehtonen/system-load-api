package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sysload/collectors"
	"sysload/models"
	"time"

	"golang.org/x/sync/errgroup"
)

// Health returns a simple health check.
func Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// CPU returns CPU metrics.
func CPU(w http.ResponseWriter, r *http.Request) {
	stats, err := collectors.CollectCPU()
	if err != nil {
		writeError(w, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models.CPUResponse{
		Timestamp: models.Now(),
		CPU:       stats,
	})
}

// Memory returns memory metrics.
func Memory(w http.ResponseWriter, r *http.Request) {
	stats, err := collectors.CollectMemory()
	if err != nil {
		writeError(w, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models.MemoryResponse{
		Timestamp: models.Now(),
		Memory:    stats,
	})
}

// Disk returns disk metrics.
func Disk(w http.ResponseWriter, r *http.Request) {
	stats, err := collectors.CollectDisk()
	if err != nil {
		writeError(w, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models.DiskResponse{
		Timestamp: models.Now(),
		Disk:      stats,
	})
}

// GPU returns GPU metrics.
func GPU(w http.ResponseWriter, r *http.Request) {
	stats := collectors.CollectGPU() // never returns error
	writeJSON(w, http.StatusOK, models.GPUResponse{
		Timestamp: models.Now(),
		GPU:       stats,
	})
}

// Network returns network metrics.
func Network(w http.ResponseWriter, r *http.Request) {
	stats, err := collectors.CollectNetwork()
	if err != nil {
		writeError(w, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models.NetworkResponse{
		Timestamp: models.Now(),
		Network:   stats,
	})
}

// Host returns host-level system info.
func Host(w http.ResponseWriter, r *http.Request) {
	stats, err := collectors.CollectHost()
	if err != nil {
		writeError(w, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models.HostResponse{
		Timestamp: models.Now(),
		Host:      stats,
	})
}

// Stats returns all metrics concurrently. Partial results are returned if
// some collectors fail; only a total failure yields a 500 error.
func Stats(w http.ResponseWriter, r *http.Request) {
	var (
		cpuStats  *models.CPUStats
		memStats  *models.MemoryStats
		diskStats *models.DiskStats
		gpuStats  *models.GPUStats
		netStats  *models.NetworkStats
		hostStats *models.HostStats
		cpuErr    error
		memErr    error
		diskErr   error
		netErr    error
		hostErr   error
	)

	g, _ := errgroup.WithContext(r.Context())

	g.Go(func() error {
		cpuStats, cpuErr = collectors.CollectCPU()
		return cpuErr
	})
	g.Go(func() error {
		memStats, memErr = collectors.CollectMemory()
		return memErr
	})
	g.Go(func() error {
		diskStats, diskErr = collectors.CollectDisk()
		return diskErr
	})
	g.Go(func() error {
		gpuStats = collectors.CollectGPU()
		return nil
	})
	g.Go(func() error {
		netStats, netErr = collectors.CollectNetwork()
		return netErr
	})
	g.Go(func() error {
		hostStats, hostErr = collectors.CollectHost()
		return hostErr
	})

	_ = g.Wait()

	if cpuStats == nil && memStats == nil && diskStats == nil && gpuStats == nil && netStats == nil && hostStats == nil {
		var parts []string
		for _, e := range []error{cpuErr, memErr, diskErr, netErr, hostErr} {
			if e != nil {
				parts = append(parts, e.Error())
			}
		}
		writeError(w, "all collectors failed: "+strings.Join(parts, "; "))
		return
	}

	writeJSON(w, http.StatusOK, models.StatsResponse{
		Timestamp: models.Now(),
		CPU:       cpuStats,
		Memory:    memStats,
		Disk:      diskStats,
		GPU:       gpuStats,
		Network:   netStats,
		Host:      hostStats,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: msg})
}

// TimeoutMiddleware enforces a 10s request deadline. Returns HTTP 504 on timeout.
func TimeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		r = r.WithContext(ctx)

		done := make(chan struct{})
		go func() {
			next.ServeHTTP(w, r)
			close(done)
		}()

		select {
		case <-done:
		case <-ctx.Done():
			writeJSON(w, http.StatusGatewayTimeout, models.ErrorResponse{Error: "request timeout"})
		}
	})
}
