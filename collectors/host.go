package collectors

import (
	"context"
	"sysload/models"

	"github.com/shirou/gopsutil/v3/host"
)

// CollectHost gathers host-level system info (uptime, etc.).
// The context is accepted for API consistency; gopsutil does not support
// cancellation of individual system calls.
func CollectHost(ctx context.Context) (*models.HostStats, error) {
	uptime, err := host.Uptime()
	if err != nil {
		return nil, err
	}

	return &models.HostStats{
		UptimeSeconds: uptime,
	}, nil
}
