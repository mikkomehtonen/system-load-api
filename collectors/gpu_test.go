package collectors

import (
	"testing"
)

func TestParseNvidiaSMI_SingleGPU(t *testing.T) {
	input := []byte("0, NVIDIA GeForce RTX 4090, 85, 24576, 12288, 72, 60\n")
	devices, err := parseNvidiaSMI(input)
	if err != nil {
		t.Fatalf("parseNvidiaSMI() error: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("len(devices) = %d, want 1", len(devices))
	}

	d := devices[0]
	if d.Index != 0 {
		t.Errorf("Index = %d, want 0", d.Index)
	}
	if d.Name != "NVIDIA GeForce RTX 4090" {
		t.Errorf("Name = %q, want %q", d.Name, "NVIDIA GeForce RTX 4090")
	}
	if d.UtilizationPercent != 85.0 {
		t.Errorf("UtilizationPercent = %v, want 85.0", d.UtilizationPercent)
	}
	if d.MemoryTotalMB != 24576.0 {
		t.Errorf("MemoryTotalMB = %v, want 24576.0", d.MemoryTotalMB)
	}
	if d.MemoryUsedMB != 12288.0 {
		t.Errorf("MemoryUsedMB = %v, want 12288.0", d.MemoryUsedMB)
	}
	if d.MemoryUsagePercent != 50.0 {
		t.Errorf("MemoryUsagePercent = %v, want 50.0", d.MemoryUsagePercent)
	}
	if d.TemperatureC != 72.0 {
		t.Errorf("TemperatureC = %v, want 72.0", d.TemperatureC)
	}
	if d.FanSpeedPercent != 60.0 {
		t.Errorf("FanSpeedPercent = %v, want 60.0", d.FanSpeedPercent)
	}
}

func TestParseNvidiaSMI_MultipleGPUs(t *testing.T) {
	input := []byte("0, NVIDIA A100, 42, 81920, 32768, 65, 45\n1, NVIDIA A100, 78, 81920, 65536, 81, 70\n")
	devices, err := parseNvidiaSMI(input)
	if err != nil {
		t.Fatalf("parseNvidiaSMI() error: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("len(devices) = %d, want 2", len(devices))
	}

	if devices[0].Index != 0 {
		t.Errorf("devices[0].Index = %d, want 0", devices[0].Index)
	}
	if devices[1].Index != 1 {
		t.Errorf("devices[1].Index = %d, want 1", devices[1].Index)
	}
	if devices[1].UtilizationPercent != 78.0 {
		t.Errorf("devices[1].UtilizationPercent = %v, want 78.0", devices[1].UtilizationPercent)
	}
}

func TestParseNvidiaSMI_EmptyInput(t *testing.T) {
	_, err := parseNvidiaSMI([]byte(""))
	if err == nil {
		t.Fatal("parseNvidiaSMI('') should return error for no devices")
	}
}

func TestParseNvidiaSMI_OnlyWhitespace(t *testing.T) {
	_, err := parseNvidiaSMI([]byte("\n\n"))
	if err == nil {
		t.Fatal("parseNvidiaSMI should return error when no devices parsed")
	}
}

func TestParseNvidiaSMI_TruncatedRow(t *testing.T) {
	input := []byte("0, NVIDIA RTX 4090, 85\n")
	devices, err := parseNvidiaSMI(input)
	if err == nil {
		t.Fatalf("expected error for truncated row, got devices: %v", devices)
	}
}

func TestParseNvidiaSMI_MalformedNumber(t *testing.T) {
	input := []byte("0, NVIDIA RTX 4090, N/A, 24576, 12288, 72, 55\n")
	devices, err := parseNvidiaSMI(input)
	if err != nil {
		t.Fatalf("parseNvidiaSMI() with N/A utilization error: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("len(devices) = %d, want 1", len(devices))
	}
	if devices[0].UtilizationPercent != 0 {
		t.Errorf("UtilizationPercent = %v, want 0 for unparseable value", devices[0].UtilizationPercent)
	}
}

func TestParseNvidiaSMI_FractionalValues(t *testing.T) {
	input := []byte("0, NVIDIA RTX 3080, 33.3, 10240, 5120.5, 67.8, 42.5\n")
	devices, err := parseNvidiaSMI(input)
	if err != nil {
		t.Fatalf("parseNvidiaSMI() error: %v", err)
	}
	if devices[0].UtilizationPercent != 33.3 {
		t.Errorf("UtilizationPercent = %v, want 33.3", devices[0].UtilizationPercent)
	}
	if devices[0].MemoryUsedMB != 5120.5 {
		t.Errorf("MemoryUsedMB = %v, want 5120.5", devices[0].MemoryUsedMB)
	}
	if devices[0].TemperatureC != 67.8 {
		t.Errorf("TemperatureC = %v, want 67.8", devices[0].TemperatureC)
	}
	if devices[0].FanSpeedPercent != 42.5 {
		t.Errorf("FanSpeedPercent = %v, want 42.5", devices[0].FanSpeedPercent)
	}
}

func TestParseNvidiaSMI_NoFanSpeedColumn(t *testing.T) {
	input := []byte("0, NVIDIA RTX 4090, 85, 24576, 12288, 72\n")
	devices, err := parseNvidiaSMI(input)
	if err != nil {
		t.Fatalf("parseNvidiaSMI() error: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("len(devices) = %d, want 1", len(devices))
	}
	if devices[0].FanSpeedPercent != 0 {
		t.Errorf("FanSpeedPercent = %v, want 0 when fan column missing", devices[0].FanSpeedPercent)
	}
}

func TestParseNvidiaSMI_FanSpeedNA(t *testing.T) {
	input := []byte("0, NVIDIA A100, 42, 81920, 32768, 65, N/A\n")
	devices, err := parseNvidiaSMI(input)
	if err != nil {
		t.Fatalf("parseNvidiaSMI() error: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("len(devices) = %d, want 1", len(devices))
	}
	if devices[0].FanSpeedPercent != 0 {
		t.Errorf("FanSpeedPercent = %v, want 0 for N/A fan speed", devices[0].FanSpeedPercent)
	}
}

func TestCollectGPU_NvidiaSMINotFound(t *testing.T) {
	stats := CollectGPU()
	if stats == nil {
		t.Fatal("CollectGPU() returned nil")
	}
	if !stats.Available {
		if stats.Error == "" {
			t.Error("GPU not available but Error is empty")
		}
		if stats.Devices != nil {
			t.Error("Devices should be nil when GPU not available")
		}
	}
}

func TestCollectGPU_AvailableHasDevices(t *testing.T) {
	stats := CollectGPU()
	if stats == nil {
		t.Fatal("CollectGPU() returned nil")
	}
	if stats.Available {
		if len(stats.Devices) == 0 {
			t.Error("GPU available but no devices returned")
		}
		for i, d := range stats.Devices {
			if d.Name == "" {
				t.Errorf("devices[%d].Name is empty", i)
			}
			if d.UtilizationPercent < 0 || d.UtilizationPercent > 100 {
				t.Errorf("devices[%d].UtilizationPercent = %v, want [0, 100]", i, d.UtilizationPercent)
			}
		}
	}
}
