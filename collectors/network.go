package collectors

import (
	"sysload/models"
	"time"

	"github.com/shirou/gopsutil/v3/net"
)

// CollectNetwork gathers per-interface byte rates by sampling over a 1-second interval.
func CollectNetwork() (*models.NetworkStats, error) {
	t0, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}

	time.Sleep(1 * time.Second)

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
		c0, ok := ifaceMap[c1.Name]
		if !ok {
			continue
		}
		interfaces = append(interfaces, models.NetworkInterface{
			Name:         c1.Name,
			BytesSentSec: c1.BytesSent - c0.BytesSent,
			BytesRecvSec: c1.BytesRecv - c0.BytesRecv,
		})
	}

	return &models.NetworkStats{Interfaces: interfaces}, nil
}
