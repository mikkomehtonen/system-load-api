package models

import "time"

// StatsResponse is the full response for /api/v1/stats.
type StatsResponse struct {
	Timestamp string        `json:"timestamp"`
	CPU       *CPUStats     `json:"cpu,omitempty"`
	Memory    *MemoryStats  `json:"memory,omitempty"`
	Disk      *DiskStats    `json:"disk,omitempty"`
	GPU       *GPUStats     `json:"gpu,omitempty"`
	Network   *NetworkStats `json:"network,omitempty"`
	Host      *HostStats    `json:"host,omitempty"`
}

// CPUResponse is the response for /api/v1/cpu.
type CPUResponse struct {
	Timestamp string    `json:"timestamp"`
	CPU       *CPUStats `json:"cpu"`
}

// MemoryResponse is the response for /api/v1/memory.
type MemoryResponse struct {
	Timestamp string       `json:"timestamp"`
	Memory    *MemoryStats `json:"memory"`
}

// DiskResponse is the response for /api/v1/disk.
type DiskResponse struct {
	Timestamp string     `json:"timestamp"`
	Disk      *DiskStats `json:"disk"`
}

// GPUResponse is the response for /api/v1/gpu.
type GPUResponse struct {
	Timestamp string    `json:"timestamp"`
	GPU       *GPUStats `json:"gpu"`
}

// NetworkResponse is the response for /api/v1/network.
type NetworkResponse struct {
	Timestamp string        `json:"timestamp"`
	Network   *NetworkStats `json:"network"`
}

// HostResponse is the response for /api/v1/host.
type HostResponse struct {
	Timestamp string     `json:"timestamp"`
	Host      *HostStats `json:"host"`
}

// ErrorResponse is returned on failure.
type ErrorResponse struct {
	Error string `json:"error"`
}

// CPUStats holds CPU metrics.
type CPUStats struct {
	LoadAvg1min    float64   `json:"load_avg_1min"`
	LoadAvg5min    float64   `json:"load_avg_5min"`
	LoadAvg15min   float64   `json:"load_avg_15min"`
	UsagePercent   float64   `json:"usage_percent"`
	CoreCount      int       `json:"core_count"`
	PerCorePercent []float64 `json:"per_core_percent"`
}

// MemoryStats holds memory metrics.
type MemoryStats struct {
	TotalGB          float64 `json:"total_gb"`
	UsedGB           float64 `json:"used_gb"`
	AvailableGB      float64 `json:"available_gb"`
	UsagePercent     float64 `json:"usage_percent"`
	SwapTotalGB      float64 `json:"swap_total_gb"`
	SwapUsedGB       float64 `json:"swap_used_gb"`
	SwapUsagePercent float64 `json:"swap_usage_percent"`
}

// DiskStats holds disk metrics.
type DiskStats struct {
	Partitions []DiskPartition `json:"partitions"`
	IO         *DiskIO         `json:"io"`
}

// DiskPartition holds per-partition info.
type DiskPartition struct {
	Device       string  `json:"device"`
	Mountpoint   string  `json:"mountpoint"`
	Fstype       string  `json:"fstype"`
	TotalGB      float64 `json:"total_gb"`
	UsedGB       float64 `json:"used_gb"`
	UsagePercent float64 `json:"usage_percent"`
}

// DiskIO holds aggregate I/O rates.
type DiskIO struct {
	ReadBytesSec  uint64 `json:"read_bytes_sec"`
	WriteBytesSec uint64 `json:"write_bytes_sec"`
}

// GPUStats holds GPU metrics.
type GPUStats struct {
	Available bool        `json:"available"`
	Devices   []GPUDevice `json:"devices"`
	Error     string      `json:"error,omitempty"`
}

// GPUDevice holds per-GPU info.
type GPUDevice struct {
	Index              int     `json:"index"`
	Name               string  `json:"name"`
	UtilizationPercent float64 `json:"utilization_percent"`
	MemoryTotalMB      float64 `json:"memory_total_mb"`
	MemoryUsedMB       float64 `json:"memory_used_mb"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`
	TemperatureC       float64 `json:"temperature_c"`
	FanSpeedPercent    float64 `json:"fan_speed_percent"`
}

// NetworkStats holds network metrics.
type NetworkStats struct {
	Interfaces []NetworkInterface `json:"interfaces"`
}

// NetworkInterface holds per-interface rates.
type NetworkInterface struct {
	Name         string `json:"name"`
	BytesSentSec uint64 `json:"bytes_sent_sec"`
	BytesRecvSec uint64 `json:"bytes_recv_sec"`
}

// HostStats holds host-level system info.
type HostStats struct {
	UptimeSeconds uint64 `json:"uptime_seconds"`
}

// Now returns the current timestamp in RFC 3339 format.
func Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}
