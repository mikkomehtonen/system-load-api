package collectors

import "testing"

func TestCollectNetwork(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode (1s delta sampling)")
	}

	stats, err := CollectNetwork()
	if err != nil {
		t.Fatalf("CollectNetwork() error: %v", err)
	}
	if stats == nil {
		t.Fatal("CollectNetwork() returned nil stats")
	}

	if len(stats.Interfaces) == 0 {
		t.Error("Interfaces is empty, expected at least one interface")
	}

	for i, iface := range stats.Interfaces {
		if iface.Name == "" {
			t.Errorf("Interfaces[%d].Name is empty", i)
		}
	}
}
