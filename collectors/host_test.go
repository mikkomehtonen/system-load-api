package collectors

import (
	"context"
	"testing"
)

func TestCollectHost(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	stats, err := CollectHost(context.Background())
	if err != nil {
		t.Fatalf("CollectHost() error: %v", err)
	}
	if stats == nil {
		t.Fatal("CollectHost() returned nil stats")
	}

	if stats.UptimeSeconds == 0 {
		t.Errorf("UptimeSeconds = %d, want > 0", stats.UptimeSeconds)
	}
}
