package models

import (
	"strings"
	"testing"
	"time"
)

func TestNow_Format(t *testing.T) {
	ts := Now()
	if ts == "" {
		t.Fatal("Now() returned empty string")
	}

	_, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		t.Errorf("Now() = %q is not valid RFC 3339: %v", ts, err)
	}
}

func TestNow_UTC(t *testing.T) {
	ts := Now()
	if !strings.HasSuffix(ts, "Z") && !strings.Contains(ts, "+00:00") {
		t.Errorf("Now() = %q, expected UTC timezone indicator", ts)
	}
}

func TestNow_NotZero(t *testing.T) {
	before := time.Now().UTC().Format(time.RFC3339)
	ts := Now()
	after := time.Now().UTC().Format(time.RFC3339)

	if ts < before || ts > after {
		t.Errorf("Now() = %q, expected between %q and %q", ts, before, after)
	}
}

func TestStatsResponse_JSONTags(t *testing.T) {
	s := StatsResponse{
		Timestamp: "2024-01-01T00:00:00Z",
	}
	if s.Timestamp == "" {
		t.Error("Timestamp should be settable")
	}
}

func TestGPUStats_AvailableFalse(t *testing.T) {
	s := GPUStats{
		Available: false,
		Error:     "nvidia-smi not found",
	}
	if s.Available {
		t.Error("Available should be false")
	}
	if s.Error != "nvidia-smi not found" {
		t.Errorf("Error = %q, want %q", s.Error, "nvidia-smi not found")
	}
}

func TestCPUStats_Fields(t *testing.T) {
	s := CPUStats{
		LoadAvg1min:    1.5,
		LoadAvg5min:    2.0,
		LoadAvg15min:   2.5,
		UsagePercent:   45.6,
		CoreCount:      8,
		PerCorePercent: []float64{10.0, 20.0, 30.0, 40.0, 50.0, 60.0, 70.0, 80.0},
	}
	if s.CoreCount != 8 {
		t.Errorf("CoreCount = %d, want 8", s.CoreCount)
	}
	if len(s.PerCorePercent) != 8 {
		t.Errorf("PerCorePercent length = %d, want 8", len(s.PerCorePercent))
	}
}

func TestMemoryStats_Fields(t *testing.T) {
	s := MemoryStats{
		TotalGB:          16.0,
		UsedGB:           8.0,
		AvailableGB:      8.0,
		UsagePercent:     50.0,
		SwapTotalGB:      8.0,
		SwapUsedGB:       2.0,
		SwapUsagePercent: 25.0,
	}
	if s.UsedGB+s.AvailableGB > s.TotalGB {
		t.Errorf("UsedGB(%v) + AvailableGB(%v) > TotalGB(%v)", s.UsedGB, s.AvailableGB, s.TotalGB)
	}
}

func TestDiskPartition_Fields(t *testing.T) {
	p := DiskPartition{
		Device:       "/dev/sda1",
		Mountpoint:   "/",
		Fstype:       "ext4",
		TotalGB:      500.0,
		UsedGB:       250.0,
		UsagePercent: 50.0,
	}
	if p.Device != "/dev/sda1" {
		t.Errorf("Device = %q, want %q", p.Device, "/dev/sda1")
	}
}

func TestGPUDevice_Fields(t *testing.T) {
	d := GPUDevice{
		Index:              0,
		Name:               "NVIDIA RTX 4090",
		UtilizationPercent: 75.0,
		MemoryTotalMB:      24576.0,
		MemoryUsedMB:       12288.0,
		MemoryUsagePercent: 50.0,
		TemperatureC:       72.0,
	}
	if d.Index != 0 {
		t.Errorf("Index = %d, want 0", d.Index)
	}
}

func TestNetworkInterface_Fields(t *testing.T) {
	n := NetworkInterface{
		Name:         "eth0",
		BytesSentSec: 1024,
		BytesRecvSec: 2048,
	}
	if n.Name != "eth0" {
		t.Errorf("Name = %q, want %q", n.Name, "eth0")
	}
}

func TestErrorResponse(t *testing.T) {
	e := ErrorResponse{Error: "something failed"}
	if e.Error != "something failed" {
		t.Errorf("Error = %q, want %q", e.Error, "something failed")
	}
}
