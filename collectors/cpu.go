package collectors

import (
	"sysload/models"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
)

// CollectCPU gathers CPU load averages, overall usage, and per-core usage.
func CollectCPU() (*models.CPUStats, error) {
	avg, err := load.Avg()
	if err != nil {
		return nil, err
	}

	coreCount, err := cpu.Counts(true)
	if err != nil {
		return nil, err
	}

	// Overall CPU usage (average over a short interval).
	usagePercents, err := cpu.Percent(0, false)
	if err != nil {
		return nil, err
	}
	var usagePercent float64
	if len(usagePercents) > 0 {
		usagePercent = usagePercents[0]
	}

	// Per-core CPU usage.
	perCorePercents, err := cpu.Percent(0, true)
	if err != nil {
		return nil, err
	}

	return &models.CPUStats{
		LoadAvg1min:    avg.Load1,
		LoadAvg5min:    avg.Load5,
		LoadAvg15min:   avg.Load15,
		UsagePercent:   roundTo1(usagePercent),
		CoreCount:      coreCount,
		PerCorePercent: roundSlice(perCorePercents),
	}, nil
}

func roundTo1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}

func roundSlice(s []float64) []float64 {
	out := make([]float64, len(s))
	for i, v := range s {
		out[i] = roundTo1(v)
	}
	return out
}
