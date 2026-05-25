package collectors

import (
	"encoding/csv"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sysload/models"
)

// CollectGPU gathers GPU metrics via nvidia-smi. If nvidia-smi is unavailable,
// it returns a GPUStats with Available=false and never returns an error.
func CollectGPU() *models.GPUStats {
	path, err := exec.LookPath("nvidia-smi")
	if err != nil {
		return &models.GPUStats{
			Available: false,
			Devices:   nil,
			Error:     "nvidia-smi not found",
		}
	}

	queryFields := "index,name,utilization.gpu,memory.total,memory.used,temperature.gpu"
	cmd := exec.Command(path,
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

	devices, parseErr := parseNvidiaSMI(out)
	if parseErr != nil {
		return &models.GPUStats{
			Available: false,
			Devices:   nil,
			Error:     fmt.Sprintf("nvidia-smi parse error: %v", parseErr),
		}
	}

	return &models.GPUStats{
		Available: true,
		Devices:   devices,
	}
}

func parseNvidiaSMI(out []byte) ([]models.GPUDevice, error) {
	reader := csv.NewReader(strings.NewReader(string(out)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var devices []models.GPUDevice
	for _, r := range records {
		if len(r) < 6 {
			continue
		}
		index, _ := strconv.Atoi(strings.TrimSpace(r[0]))
		name := strings.TrimSpace(r[1])
		utilPct, _ := strconv.ParseFloat(strings.TrimSpace(r[2]), 64)
		memTotal, _ := strconv.ParseFloat(strings.TrimSpace(r[3]), 64)
		memUsed, _ := strconv.ParseFloat(strings.TrimSpace(r[4]), 64)
		temp, _ := strconv.ParseFloat(strings.TrimSpace(r[5]), 64)
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
		})
	}

	if len(devices) == 0 {
		return nil, fmt.Errorf("no GPU devices parsed")
	}
	return devices, nil
}
