package collector

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/rs/zerolog/log"
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
//
// Load averages (Load1, Load5, Load15) represent the average number of
// processes in the run queue over 1, 5, and 15 minute intervals respectively.
//
// On Windows, load averages are not natively supported by the OS.
// In that case, these fields are set to 0 and omitted from JSON output
// via `omitempty` to avoid sending misleading data to the receiver.
type CPUStats struct {
	// UsagePercent is the overall CPU utilization as a percentage (0–100).
	UsagePercent float64 `json:"usage_percent"`

	// Load1 is the 1-minute load average. Zero (and omitted) on Windows.
	Load1 float64 `json:"load_1,omitempty"`

	// Load5 is the 5-minute load average. Zero (and omitted) on Windows.
	Load5 float64 `json:"load_5,omitempty"`

	// Load15 is the 15-minute load average. Zero (and omitted) on Windows.
	Load15 float64 `json:"load_15,omitempty"`
}

// Collect gathers CPU usage and load average metrics.
//
// CPU usage is measured over a short 200ms interval for responsiveness.
// Load averages are collected on supported platforms (Linux, macOS).
// On Windows, load average collection is skipped gracefully — the fields
// remain at their zero values and are omitted from the JSON payload.
//
// The method respects context cancellation at each blocking point.
func (c *CPUCollector) Collect(ctx context.Context, metrics *Metrics) error {
	// Check if the context has already been cancelled before starting
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// --- CPU Usage ---
	// Run cpu.Percent in a goroutine so we can respect context cancellation
	// during the 200ms measurement interval.
	type cpuResult struct {
		percentages []float64
		err         error
	}

	cpuChan := make(chan cpuResult, 1)
	go func() {
		// Measure total CPU usage across all cores (false = aggregate, not per-core)
		percentages, err := cpu.Percent(200*time.Millisecond, false)
		cpuChan <- cpuResult{percentages: percentages, err: err}
	}()

	// Wait for CPU measurement or context cancellation
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

	// --- Load Averages ---
	// Load averages are a Unix concept and are not available on Windows.
	// We skip collection on Windows and leave the fields at zero (omitempty
	// ensures they are excluded from the JSON payload sent to the receiver).
	if runtime.GOOS == "windows" {
		log.Debug().Msg("Skipping load average collection: not supported on Windows")
		return nil
	}

	avg, err := load.Avg()
	if err != nil {
		// Non-fatal: log a warning and continue without load data.
		// This handles edge cases on exotic Linux configurations or containers
		// where /proc/loadavg may be unavailable.
		log.Warn().Err(err).Msg("Failed to collect load averages; skipping")
		return nil
	}

	metrics.CPU.Load1 = avg.Load1
	metrics.CPU.Load5 = avg.Load5
	metrics.CPU.Load15 = avg.Load15

	return nil
}
