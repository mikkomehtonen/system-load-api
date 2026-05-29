package collectors

import (
	"context"
	"sysload/models"

	"github.com/shirou/gopsutil/v3/mem"
)

const gb = 1024 * 1024 * 1024

// CollectMemory gathers RAM and swap usage.
// The context is accepted for API consistency; gopsutil does not support
// cancellation of individual system calls.
func CollectMemory(ctx context.Context) (*models.MemoryStats, error) {
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
