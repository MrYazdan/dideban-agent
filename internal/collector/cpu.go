package collector

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
)

// CPUCollector is responsible for collecting CPU-related metrics
// such as usage percentage and system load averages.
type CPUCollector struct{}

// Name returns the unique name of this collector.
// It is used for logging and debugging purposes.
func (c *CPUCollector) Name() string {
	return "cpu"
}

// CPUStats represents CPU-related metrics at a given point in time.
type CPUStats struct {
	UsagePercent float64 `json:"usage_percent"`
	Load1        float64 `json:"load_1"`
	Load5        float64 `json:"load_5,omitempty"`
	Load15       float64 `json:"load_15,omitempty"`
}

// Collect gathers CPU usage and load average metrics.
// It respects the provided context for cancellation.
func (c *CPUCollector) Collect(ctx context.Context, metrics *Metrics) error {
	// Check if the context has already been cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Use a goroutine with context to make CPU collection cancellable
	type cpuResult struct {
		percentages []float64
		err         error
	}

	cpuChan := make(chan cpuResult, 1)
	go func() {
		// Get CPU usage with a short interval for responsiveness
		percentages, err := cpu.Percent(200*time.Millisecond, false)
		cpuChan <- cpuResult{percentages: percentages, err: err}
	}()

	// Wait for CPU result or context cancellation
	var cpuRes cpuResult
	select {
	case <-ctx.Done():
		return ctx.Err()
	case cpuRes = <-cpuChan:
	}

	if cpuRes.err != nil {
		return fmt.Errorf("failed to get CPU usage: %w", cpuRes.err)
	}

	if len(cpuRes.percentages) > 0 {
		metrics.CPU.UsagePercent = math.Round(cpuRes.percentages[0])
	}

	// Retrieve system load averages (1m, 5m, 15m)
	avg, err := load.Avg()
	if err != nil {
		return fmt.Errorf("failed to get load average: %w", err)
	}

	metrics.CPU.Load1 = avg.Load1
	metrics.CPU.Load5 = avg.Load5
	metrics.CPU.Load15 = avg.Load15

	return nil
}
