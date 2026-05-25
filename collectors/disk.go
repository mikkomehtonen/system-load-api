package collectors

import (
	"sysload/models"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

// CollectDisk gathers partition usage and aggregate I/O rates (delta-sampled over 1s).
func CollectDisk() (*models.DiskStats, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	var parts []models.DiskPartition
	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue // skip inaccessible partitions
		}
		parts = append(parts, models.DiskPartition{
			Device:       p.Device,
			Mountpoint:   p.Mountpoint,
			Fstype:       p.Fstype,
			TotalGB:      roundTo2(float64(usage.Total) / gb),
			UsedGB:       roundTo2(float64(usage.Used) / gb),
			UsagePercent: roundTo1(usage.UsedPercent),
		})
	}

	// Delta-sample I/O counters over 1 second for per-second rates.
	t0, err := disk.IOCounters()
	if err != nil {
		return &models.DiskStats{Partitions: parts}, nil
	}

	time.Sleep(1 * time.Second)

	t1, err := disk.IOCounters()
	if err != nil {
		return &models.DiskStats{Partitions: parts}, nil
	}

	// Sum deltas across all disks.
	var readBytesSec, writeBytesSec uint64
	for name, c1 := range t1 {
		c0, ok := t0[name]
		if !ok {
			continue
		}
		readBytesSec += c1.ReadBytes - c0.ReadBytes
		writeBytesSec += c1.WriteBytes - c0.WriteBytes
	}

	return &models.DiskStats{
		Partitions: parts,
		IO: &models.DiskIO{
			ReadBytesSec:  readBytesSec,
			WriteBytesSec: writeBytesSec,
		},
	}, nil
}
