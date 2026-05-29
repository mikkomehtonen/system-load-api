package collectors

import (
	"context"
	"testing"
)

func TestGbConstant(t *testing.T) {
	if gb != 1024*1024*1024 {
		t.Errorf("gb = %d, want %d", gb, 1024*1024*1024)
	}
}

func TestCollectMemory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	stats, err := CollectMemory(context.Background())
	if err != nil {
		t.Fatalf("CollectMemory() error: %v", err)
	}
	if stats == nil {
		t.Fatal("CollectMemory() returned nil stats")
	}

	if stats.TotalGB <= 0 {
		t.Errorf("TotalGB = %v, want > 0", stats.TotalGB)
	}
	if stats.UsedGB < 0 {
		t.Errorf("UsedGB = %v, want >= 0", stats.UsedGB)
	}
	if stats.AvailableGB < 0 {
		t.Errorf("AvailableGB = %v, want >= 0", stats.AvailableGB)
	}
	if stats.UsagePercent < 0 || stats.UsagePercent > 100 {
		t.Errorf("UsagePercent = %v, want [0, 100]", stats.UsagePercent)
	}
	if stats.UsedGB+stats.AvailableGB > stats.TotalGB+0.01 {
		t.Errorf("UsedGB(%v) + AvailableGB(%v) > TotalGB(%v)", stats.UsedGB, stats.AvailableGB, stats.TotalGB)
	}
	if stats.SwapUsagePercent < 0 || stats.SwapUsagePercent > 100 {
		t.Errorf("SwapUsagePercent = %v, want [0, 100]", stats.SwapUsagePercent)
	}
	if stats.SwapUsedGB > stats.SwapTotalGB+0.01 && stats.SwapTotalGB > 0 {
		t.Errorf("SwapUsedGB(%v) > SwapTotalGB(%v)", stats.SwapUsedGB, stats.SwapTotalGB)
	}
}
