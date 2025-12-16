package collector

import (
	"context"
	"fmt"
	"math"

	"github.com/shirou/gopsutil/mem"
)

// MemoryCollector is responsible for collecting memory-related metrics
// such as total, used, available memory and usage percentage.
type MemoryCollector struct{}

// Name returns the unique name of this collector.
// It is used for logging and debugging purposes.
func (m *MemoryCollector) Name() string {
	return "memory"
}

// MemStats represents memory-related metrics at a given point in time.
type MemStats struct {
	UsedMB       uint64  `json:"used_mb"`
	TotalMB      uint64  `json:"total_mb"`
	UsagePercent float64 `json:"usage_percent"`
	AvailableMB  uint64  `json:"available_mb,omitempty"`
}

// Collect gathers memory usage metrics and populates the Metrics struct.
// The operation respects the provided context for cancellation.
func (m *MemoryCollector) Collect(ctx context.Context, metrics *Metrics) error {
	// Check if the context has already been cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Retrieve virtual memory statistics
	virtualMemory, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("failed to get memory info: %w", err)
	}

	metrics.Memory.UsedMB = virtualMemory.Used / 1024 / 1024
	metrics.Memory.TotalMB = virtualMemory.Total / 1024 / 1024
	metrics.Memory.UsagePercent = math.Round(virtualMemory.UsedPercent)
	metrics.Memory.AvailableMB = virtualMemory.Available / 1024 / 1024

	return nil
}
