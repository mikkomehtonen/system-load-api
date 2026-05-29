package collectors

import (
	"context"
	"encoding/csv"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sysload/models"
)

// CollectGPU gathers GPU metrics via nvidia-smi. If nvidia-smi is unavailable,
// it returns a GPUStats with Available=false and never returns an error.
func CollectGPU(ctx context.Context) *models.GPUStats {
	path, err := exec.LookPath("nvidia-smi")
	if err != nil {
		return &models.GPUStats{
			Available: false,
			Devices:   nil,
			Error:     "nvidia-smi not found",
		}
	}

	queryFields := "index,name,utilization.gpu,memory.total,memory.used,temperature.gpu,fan.speed"
	cmd := exec.CommandContext(ctx, path,
		"--query-gpu="+queryFields,
		"--format=csv,noheader,nounits",
	)
	out, err := cmd.Output()
	if err != nil {
		return &models.GPUStats{
			Available: false,
			Devices:   nil,
			Error:     fmt.Sprintf("nvidia-smi failed: %v", err),
		}
	}

	devices, warnings, parseErr := parseNvidiaSMI(out)
	if parseErr != nil {
		return &models.GPUStats{
			Available: false,
			Devices:   nil,
			Error:     fmt.Sprintf("nvidia-smi parse error: %v", parseErr),
		}
	}

	stats := &models.GPUStats{
		Available: true,
		Devices:   devices,
	}
	if len(warnings) > 0 {
		stats.Error = "nvidia-smi field parse warnings: " + strings.Join(warnings, "; ")
	}
	return stats
}

func parseNvidiaSMI(out []byte) ([]models.GPUDevice, []string, error) {
	reader := csv.NewReader(strings.NewReader(string(out)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	var devices []models.GPUDevice
	var warnings []string
	for _, r := range records {
		if len(r) < 6 {
			continue
		}
		index, err := strconv.Atoi(strings.TrimSpace(r[0]))
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("index %q: %v", r[0], err))
			continue
		}
		name := strings.TrimSpace(r[1])
		utilPct, err := strconv.ParseFloat(strings.TrimSpace(r[2]), 64)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("device %d utilization %q: %v", index, r[2], err))
		}
		memTotal, err := strconv.ParseFloat(strings.TrimSpace(r[3]), 64)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("device %d memory.total %q: %v", index, r[3], err))
		}
		memUsed, err := strconv.ParseFloat(strings.TrimSpace(r[4]), 64)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("device %d memory.used %q: %v", index, r[4], err))
		}
		temp, err := strconv.ParseFloat(strings.TrimSpace(r[5]), 64)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("device %d temperature %q: %v", index, r[5], err))
		}
		var fanPct float64
		if len(r) >= 7 {
			fanPct, err = strconv.ParseFloat(strings.TrimSpace(r[6]), 64)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("device %d fan.speed %q: %v", index, r[6], err))
			}
		}
		var memPct float64
		if memTotal > 0 {
			memPct = (memUsed / memTotal) * 100
		}

		devices = append(devices, models.GPUDevice{
			Index:              index,
			Name:               name,
			UtilizationPercent: roundTo1(utilPct),
			MemoryTotalMB:      roundTo1(memTotal),
			MemoryUsedMB:       roundTo1(memUsed),
			MemoryUsagePercent: roundTo1(memPct),
			TemperatureC:       roundTo1(temp),
			FanSpeedPercent:    roundTo1(fanPct),
		})
	}

	if len(devices) == 0 {
		return nil, warnings, fmt.Errorf("no GPU devices parsed")
	}
	return devices, warnings, nil
}
