package collector

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/shirou/gopsutil/cpu"
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
// Load average semantics:
//   - On Linux/macOS: sourced from the OS kernel (/proc/loadavg or equivalent).
//     Represents the average number of runnable processes over 1, 5, and 15 minutes.
//   - On Windows: derived from the WMI Processor Queue Length metric using a
//     rolling sliding-window average over the same time intervals.
//     This is the closest Windows equivalent to Unix load average.
//
// Fields use `omitempty` so that zero values (e.g., on collection failure)
// are excluded from the JSON payload sent to the receiver.
type CPUStats struct {
	// UsagePercent is the overall CPU utilization as a percentage (0–100).
	UsagePercent float64 `json:"usage_percent"`

	// Load1 is the 1-minute load average (or Windows equivalent).
	Load1 float64 `json:"load_1,omitempty"`

	// Load5 is the 5-minute load average (or Windows equivalent).
	Load5 float64 `json:"load_5,omitempty"`

	// Load15 is the 15-minute load average (or Windows equivalent).
	Load15 float64 `json:"load_15,omitempty"`
}

// Collect gathers CPU usage and load average metrics.
//
// CPU usage is measured over a short 200ms interval for responsiveness.
// Load averages are collected via platform-specific implementations:
//   - Unix (Linux/macOS): uses gopsutil/load backed by /proc/loadavg
//   - Windows: uses WMI Processor Queue Length with a rolling average tracker
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
	// Delegated to platform-specific implementations:
	//   load_unix.go  → Linux / macOS / BSD
	//   load_windows.go → Windows (WMI-based rolling average)
	collectLoadAvg(&metrics.CPU)

	return nil
}
