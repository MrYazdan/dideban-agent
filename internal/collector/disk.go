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

// DiskStats represents disk-related metrics for the primary filesystem.
type DiskStats struct {
	// UsedGB is the amount of disk space currently in use, in gigabytes.
	UsedGB uint64 `json:"used_gb"`

	// TotalGB is the total disk capacity, in gigabytes.
	TotalGB uint64 `json:"total_gb"`

	// UsagePercent is the disk utilization as a percentage (0–100).
	UsagePercent float64 `json:"usage_percent"`
}

// Collect gathers disk usage metrics for the primary filesystem.
//
// The monitored path is platform-specific and resolved via rootDiskPath():
//   - Linux / macOS: "/"  (root filesystem)
//   - Windows:       "C:\" (system drive root, with trailing backslash required by gopsutil)
//
// The operation respects the provided context for cancellation.
func (d *DiskCollector) Collect(ctx context.Context, metrics *Metrics) error {
	// Check if the context has already been cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Resolve the platform-specific root path (see disk_path_unix.go / disk_path_windows.go)
	path := rootDiskPath()

	usage, err := disk.Usage(path)
	if err != nil {
		return fmt.Errorf("failed to get disk usage for %q: %w", path, err)
	}

	metrics.Disk.UsedGB = usage.Used / 1024 / 1024 / 1024
	metrics.Disk.TotalGB = usage.Total / 1024 / 1024 / 1024
	metrics.Disk.UsagePercent = math.Round(usage.UsedPercent)

	return nil
}
