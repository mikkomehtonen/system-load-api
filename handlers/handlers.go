package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"sysload/collectors"
	"sysload/models"
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

// Stats returns all metrics concurrently. Partial results are returned if
// some collectors fail; only a total failure yields a 500 error.
func Stats(w http.ResponseWriter, r *http.Request) {
	var (
		cpuStats  *models.CPUStats
		memStats  *models.MemoryStats
		diskStats *models.DiskStats
		gpuStats  *models.GPUStats
		netStats  *models.NetworkStats
		cpuErr    error
		memErr    error
		diskErr   error
		netErr    error
	)

	var wg sync.WaitGroup
	wg.Add(5)

	go func() {
		defer wg.Done()
		cpuStats, cpuErr = collectors.CollectCPU()
	}()
	go func() {
		defer wg.Done()
		memStats, memErr = collectors.CollectMemory()
	}()
	go func() {
		defer wg.Done()
		diskStats, diskErr = collectors.CollectDisk()
	}()
	go func() {
		defer wg.Done()
		gpuStats = collectors.CollectGPU()
	}()
	go func() {
		defer wg.Done()
		netStats, netErr = collectors.CollectNetwork()
	}()

	wg.Wait()

	// If every collector failed, return error.
	if cpuStats == nil && memStats == nil && diskStats == nil && gpuStats == nil && netStats == nil {
		errs := collectErrors(cpuErr, memErr, diskErr, netErr)
		writeError(w, "all collectors failed: "+errs)
		return
	}

	writeJSON(w, http.StatusOK, models.StatsResponse{
		Timestamp: models.Now(),
		CPU:       cpuStats,
		Memory:    memStats,
		Disk:      diskStats,
		GPU:       gpuStats,
		Network:   netStats,
	})
}

func collectErrors(errs ...error) string {
	var parts []string
	for _, e := range errs {
		if e != nil {
			parts = append(parts, e.Error())
		}
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += "; "
		}
		result += p
	}
	return result
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: msg})
}
