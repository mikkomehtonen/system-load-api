package collectors

import (
	"sysload/models"

	"github.com/shirou/gopsutil/v3/mem"
)

const gb = 1024 * 1024 * 1024

// CollectMemory gathers RAM and swap usage.
func CollectMemory() (*models.MemoryStats, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	s, err := mem.SwapMemory()
	if err != nil {
		return nil, err
	}

	return &models.MemoryStats{
		TotalGB:          roundTo2(float64(v.Total) / gb),
		UsedGB:           roundTo2(float64(v.Used) / gb),
		AvailableGB:      roundTo2(float64(v.Available) / gb),
		UsagePercent:     roundTo1(v.UsedPercent),
		SwapTotalGB:      roundTo2(float64(s.Total) / gb),
		SwapUsedGB:       roundTo2(float64(s.Used) / gb),
		SwapUsagePercent: roundTo1(s.UsedPercent),
	}, nil
}

func roundTo2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
