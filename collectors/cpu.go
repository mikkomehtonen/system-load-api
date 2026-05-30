package collectors

import (
	"context"
	"strings"
	"sysload/models"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
)

// CollectCPU gathers CPU load averages, overall usage, and per-core usage.
// The context is accepted for API consistency; gopsutil does not support
// cancellation of individual system calls.
func CollectCPU(ctx context.Context) (*models.CPUStats, error) {
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

	temperatureC := collectCPUTemperature()

	return &models.CPUStats{
		LoadAvg1min:    avg.Load1,
		LoadAvg5min:    avg.Load5,
		LoadAvg15min:   avg.Load15,
		UsagePercent:   roundTo1(usagePercent),
		CoreCount:      coreCount,
		PerCorePercent: roundSlice(perCorePercents),
		TemperatureC:   temperatureC,
	}, nil
}

func collectCPUTemperature() *float64 {
	temps, err := host.SensorsTemperatures()
	if err != nil || len(temps) == 0 {
		return nil
	}
	return selectCPUTemperature(temps)
}

func selectCPUTemperature(temps []host.TemperatureStat) *float64 {
	var packageTemp, dieTemp, highestCoreTemp float64
	hasPackage, hasDie, hasCore := false, false, false

	for _, t := range temps {
		key := strings.ToLower(t.SensorKey)
		if t.Temperature <= 0 {
			continue
		}
		switch {
		case key == "package" || key == "package_id_0":
			if t.Temperature > packageTemp {
				packageTemp = t.Temperature
				hasPackage = true
			}
		case key == "die" || key == "cpu_die":
			if t.Temperature > dieTemp {
				dieTemp = t.Temperature
				hasDie = true
			}
		case strings.HasPrefix(key, "core") || strings.HasPrefix(key, "cpu_core"):
			if t.Temperature > highestCoreTemp {
				highestCoreTemp = t.Temperature
				hasCore = true
			}
		}
	}

	if hasPackage {
		return tempPtr(roundTo1(packageTemp))
	}
	if hasDie {
		return tempPtr(roundTo1(dieTemp))
	}
	if hasCore {
		return tempPtr(roundTo1(highestCoreTemp))
	}
	return nil
}

func tempPtr(v float64) *float64 {
	return &v
}
