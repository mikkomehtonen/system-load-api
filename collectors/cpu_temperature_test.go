package collectors

import (
	"context"
	"math"
	"testing"

	"github.com/shirou/gopsutil/v3/host"
)

func TestSelectCPUTemperature_PackagePreferred(t *testing.T) {
	temps := []host.TemperatureStat{
		{SensorKey: "core_0", Temperature: 55.0},
		{SensorKey: "core_1", Temperature: 58.0},
		{SensorKey: "package_id_0", Temperature: 48.0},
	}
	got := selectCPUTemperature(temps)
	if got == nil {
		t.Fatal("expected non-nil temperature")
	}
	if math.Abs(*got-48.0) > 0.01 {
		t.Errorf("temperature = %v, want 48.0", *got)
	}
}

func TestSelectCPUTemperature_DieFallback(t *testing.T) {
	temps := []host.TemperatureStat{
		{SensorKey: "core_0", Temperature: 62.0},
		{SensorKey: "die", Temperature: 50.5},
	}
	got := selectCPUTemperature(temps)
	if got == nil {
		t.Fatal("expected non-nil temperature")
	}
	if math.Abs(*got-50.5) > 0.01 {
		t.Errorf("temperature = %v, want 50.5", *got)
	}
}

func TestSelectCPUTemperature_CoreFallback(t *testing.T) {
	temps := []host.TemperatureStat{
		{SensorKey: "cpu_core_0", Temperature: 35.0},
		{SensorKey: "cpu_core_3", Temperature: 72.8},
	}
	got := selectCPUTemperature(temps)
	if got == nil {
		t.Fatal("expected non-nil temperature")
	}
	if math.Abs(*got-72.8) > 0.01 {
		t.Errorf("temperature = %v, want 72.8 (highest core)", *got)
	}
}

func TestSelectCPUTemperature_FiltersZeroOrNegative(t *testing.T) {
	temps := []host.TemperatureStat{
		{SensorKey: "core_0", Temperature: 0.0},
		{SensorKey: "core_1", Temperature: -5.0},
		{SensorKey: "package_id_0", Temperature: 45.0},
		{SensorKey: "core_2", Temperature: 60.0},
	}
	got := selectCPUTemperature(temps)
	if got == nil {
		t.Fatal("expected non-nil temperature")
	}
	if math.Abs(*got-45.0) > 0.01 {
		t.Errorf("temperature = %v, want 45.0 (package, ignoring 0/negative cores)", *got)
	}
}

func TestSelectCPUTemperature_EmptyReturnsNil(t *testing.T) {
	if got := selectCPUTemperature(nil); got != nil {
		t.Errorf("nil input: expected nil, got %v", *got)
	}
	if got := selectCPUTemperature([]host.TemperatureStat{}); got != nil {
		t.Errorf("empty slice: expected nil, got %v", *got)
	}
}

func TestSelectCPUTemperature_AllZeroReturnsNil(t *testing.T) {
	temps := []host.TemperatureStat{
		{SensorKey: "package_id_0", Temperature: 0.0},
		{SensorKey: "core_0", Temperature: 0.0},
	}
	if got := selectCPUTemperature(temps); got != nil {
		t.Errorf("expected nil, got %v", *got)
	}
}

func TestSelectCPUTemperature_Rounding(t *testing.T) {
	temps := []host.TemperatureStat{
		{SensorKey: "package_id_0", Temperature: 47.65},
	}
	got := selectCPUTemperature(temps)
	if got == nil {
		t.Fatal("expected non-nil temperature")
	}
	if math.Abs(*got-47.7) > 0.01 {
		t.Errorf("temperature = %v, want 47.7 (rounded to 1 decimal)", *got)
	}
}

func TestSelectCPUTemperature_CaseInsensitive(t *testing.T) {
	temps := []host.TemperatureStat{
		{SensorKey: "Package", Temperature: 52.0},
		{SensorKey: "CORE_0", Temperature: 70.0},
	}
	got := selectCPUTemperature(temps)
	if got == nil {
		t.Fatal("expected non-nil temperature")
	}
	if math.Abs(*got-52.0) > 0.01 {
		t.Errorf("temperature = %v, want 52.0 (package key case-insensitive)", *got)
	}
}

func TestCollectCPU_IncludesTemperature(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	stats, err := CollectCPU(context.Background())
	if err != nil {
		t.Fatalf("CollectCPU() error: %v", err)
	}
	if stats == nil {
		t.Fatal("CollectCPU() returned nil stats")
	}
	if stats.TemperatureC == nil {
		t.Log("TemperatureC is nil (no sensors available on this system)")
	}
	if stats.TemperatureC != nil {
		if *stats.TemperatureC < 0 {
			t.Errorf("TemperatureC = %v, want >= 0", *stats.TemperatureC)
		}
	}
}
