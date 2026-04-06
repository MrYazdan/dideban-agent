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
func (c *CPUCollector) Name() string {
	return "cpu"
}

// CPUStats represents CPU-related metrics at a given point in time.
//
// Load average semantics:
//   - Linux/macOS: kernel-provided values from /proc/loadavg or equivalent syscall.
//   - Windows: derived from WMI Processor Queue Length using a rolling sliding-window
//     average over the same 1/5/15-minute intervals.
//
// NOTE: No omitempty tags are used here. The receiver validates all fields as
// `required`, so every field must be present in the JSON payload — even when
// the value is legitimately zero (e.g., idle system with zero queue length).
type CPUStats struct {
	// UsagePercent is the overall CPU utilization as a percentage (0–100).
	UsagePercent float64 `json:"usage_percent"`

	// Load1 is the 1-minute load average (or Windows processor queue rolling avg).
	Load1 float64 `json:"load_1"`

	// Load5 is the 5-minute load average (or Windows processor queue rolling avg).
	Load5 float64 `json:"load_5"`

	// Load15 is the 15-minute load average (or Windows processor queue rolling avg).
	Load15 float64 `json:"load_15"`
}

// cpuSampleInterval is the measurement window passed to cpu.Percent.
// A longer interval yields a more accurate reading but increases collection latency.
const cpuSampleInterval = 500 * time.Millisecond

// cpuRetryInterval is used for a second measurement attempt when the first
// returns zero. On Windows, the first call after process start always returns
// 0.0 because the OS needs two samples to compute a delta.
const cpuRetryInterval = 800 * time.Millisecond

// Collect gathers CPU usage and load average metrics.
//
// CPU usage strategy:
//   - Measured over cpuSampleInterval for a stable reading.
//   - On Windows (and occasionally on Linux), the very first call to cpu.Percent
//     returns 0.0 because the OS requires two ticks to compute a delta. If a zero
//     is returned, a single retry with a longer interval is performed automatically.
//     This ensures the receiver never receives a spurious zero that fails `required`
//     validation, at the cost of slightly longer collection time on cold start.
//
// Load average strategy:
//   - Delegated to platform-specific implementations (load_unix.go / load_windows.go).
//   - On Windows cold start, the rolling tracker may have only one sample; in that
//     case the current queue length is used directly for all three fields so that
//     no field is ever zero due to missing history.
//
// The method respects context cancellation at every blocking point.
func (c *CPUCollector) Collect(ctx context.Context, metrics *Metrics) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// --- CPU Usage ---
	usage, err := measureCPUPercent(ctx, cpuSampleInterval)
	if err != nil {
		return fmt.Errorf("failed to get CPU usage: %w", err)
	}

	// On Windows (and rarely on Linux), the first sample after process start is
	// always 0.0 because cpu.Percent needs two data points to compute a delta.
	// Retry once with a longer interval to get a real reading.
	if usage == 0 {
		usage, err = measureCPUPercent(ctx, cpuRetryInterval)
		if err != nil {
			return fmt.Errorf("failed to get CPU usage (retry): %w", err)
		}
	}

	metrics.CPU.UsagePercent = math.Round(usage)

	// --- Load Averages ---
	// Delegated to platform-specific implementations:
	//   load_unix.go    → Linux / macOS / BSD  (gopsutil/load)
	//   load_windows.go → Windows              (WMI + rolling window)
	collectLoadAvg(&metrics.CPU)

	return nil
}

// measureCPUPercent measures aggregate CPU usage over the given interval.
// It runs cpu.Percent in a goroutine so context cancellation is respected
// during the blocking measurement window.
func measureCPUPercent(ctx context.Context, interval time.Duration) (float64, error) {
	type result struct {
		pct float64
		err error
	}

	ch := make(chan result, 1)
	go func() {
		// false = aggregate across all cores (single value)
		percentages, err := cpu.Percent(interval, false)
		if err != nil {
			ch <- result{err: err}
			return
		}
		var pct float64
		if len(percentages) > 0 {
			pct = percentages[0]
		}
		ch <- result{pct: pct}
	}()

	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case r := <-ch:
		return r.pct, r.err
	}
}
