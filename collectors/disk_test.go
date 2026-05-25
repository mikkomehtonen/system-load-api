package collectors

import (
	"strings"
	"testing"
)

func TestCollectDisk(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode (1s delta sampling)")
	}

	stats, err := CollectDisk()
	if err != nil {
		t.Fatalf("CollectDisk() error: %v", err)
	}
	if stats == nil {
		t.Fatal("CollectDisk() returned nil stats")
	}

	if len(stats.Partitions) == 0 {
		t.Error("Partitions is empty, expected at least one partition")
	}

	for i, p := range stats.Partitions {
		if p.Device == "" {
			t.Errorf("Partitions[%d].Device is empty", i)
		}
		if strings.HasPrefix(p.Device, "/dev/loop") {
			t.Errorf("Partitions[%d].Device = %q, want no /dev/loop* devices", i, p.Device)
		}
		if p.TotalGB < 0 {
			t.Errorf("Partitions[%d].TotalGB = %v, want >= 0", i, p.TotalGB)
		}
		if p.UsagePercent < 0 || p.UsagePercent > 100 {
			t.Errorf("Partitions[%d].UsagePercent = %v, want [0, 100]", i, p.UsagePercent)
		}
	}

	if stats.IO != nil {
		if stats.IO.ReadBytesSec < 0 {
			t.Errorf("IO.ReadBytesSec = %d, want >= 0", stats.IO.ReadBytesSec)
		}
		if stats.IO.WriteBytesSec < 0 {
			t.Errorf("IO.WriteBytesSec = %d, want >= 0", stats.IO.WriteBytesSec)
		}
	}
}
