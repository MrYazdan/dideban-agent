package collector

import (
	"context"
	"fmt"
	"math"

	"github.com/shirou/gopsutil/disk"
)

// DiskCollector is responsible for collecting disk-related metrics
// such as total, used disk space and usage percentage.
type DiskCollector struct{}

// Name returns the unique name of this collector.
// It is used for logging and debugging purposes.
func (d *DiskCollector) Name() string {
	return "disk"
}

// DiskStats represents disk-related metrics for a specific filesystem.
type DiskStats struct {
	UsedGB       uint64  `json:"used_gb"`
	TotalGB      uint64  `json:"total_gb"`
	UsagePercent float64 `json:"usage_percent"`
}

// Collect gathers disk usage metrics for the root filesystem.
// The operation respects the provided context for cancellation.
func (d *DiskCollector) Collect(ctx context.Context, metrics *Metrics) error {
	// Check if the context has already been cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Retrieve disk usage statistics for the root filesystem
	usage, err := disk.Usage("/")
	if err != nil {
		return fmt.Errorf("failed to get disk usage: %w", err)
	}

	metrics.Disk.UsedGB = usage.Used / 1024 / 1024 / 1024
	metrics.Disk.TotalGB = usage.Total / 1024 / 1024 / 1024
	metrics.Disk.UsagePercent = math.Round(usage.UsedPercent)

	return nil
}
