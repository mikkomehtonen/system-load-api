package collectors

import (
	"sysload/models"

	"github.com/shirou/gopsutil/v3/host"
)

// CollectHost gathers host-level system info (uptime, etc.).
func CollectHost() (*models.HostStats, error) {
	uptime, err := host.Uptime()
	if err != nil {
		return nil, err
	}

	return &models.HostStats{
		UptimeSeconds: uptime,
	}, nil
}
