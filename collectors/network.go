package collectors

import (
	"context"
	"strings"
	"sysload/models"
	"time"

	"github.com/shirou/gopsutil/v3/net"
)

// isVirtualInterface returns true for interfaces that should be filtered out.
func isVirtualInterface(name string) bool {
	if name == "lo" || name == "docker0" {
		return true
	}
	if strings.HasPrefix(name, "br-") || strings.HasPrefix(name, "veth") {
		return true
	}
	return false
}

// CollectNetwork gathers per-interface byte rates by sampling over a 1-second interval.
func CollectNetwork(ctx context.Context) (*models.NetworkStats, error) {
	t0, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return &models.NetworkStats{}, nil
	case <-time.After(1 * time.Second):
	}

	t1, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}

	ifaceMap := make(map[string]net.IOCountersStat)
	for _, c := range t0 {
		ifaceMap[c.Name] = c
	}

	var interfaces []models.NetworkInterface
	for _, c1 := range t1 {
		if isVirtualInterface(c1.Name) {
			continue
		}
		c0, ok := ifaceMap[c1.Name]
		if !ok {
			continue
		}
		var sent, recv uint64
		if c1.BytesSent >= c0.BytesSent {
			sent = c1.BytesSent - c0.BytesSent
		}
		if c1.BytesRecv >= c0.BytesRecv {
			recv = c1.BytesRecv - c0.BytesRecv
		}
		interfaces = append(interfaces, models.NetworkInterface{
			Name:         c1.Name,
			BytesSentSec: sent,
			BytesRecvSec: recv,
		})
	}

	return &models.NetworkStats{Interfaces: interfaces}, nil
}
