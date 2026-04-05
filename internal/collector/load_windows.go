//go:build windows

package collector

import (
	"sync"
	"time"

	"github.com/yusufpapurcu/wmi"
	"github.com/rs/zerolog/log"
)

// win32PerfRawDataPerfOSProcessorQueue maps to the WMI class
// Win32_PerfRawData_PerfOS_System, which exposes the Processor Queue Length —
// the number of threads waiting to be executed by the CPU.
//
// This is the closest Windows equivalent to Unix load average:
// it reflects instantaneous CPU run-queue depth rather than a time-averaged value.
type win32PerfRawDataPerfOSProcessorQueue struct {
	ProcessorQueueLength uint32
}

// windowsLoadTracker maintains a rolling history of processor queue length
// samples to compute 1-minute, 5-minute, and 15-minute moving averages.
//
// It is safe for concurrent use.
type windowsLoadTracker struct {
	mu      sync.Mutex
	samples []timedSample
}

// timedSample holds a single processor queue length reading with its timestamp.
type timedSample struct {
	value float64
	at    time.Time
}

// globalLoadTracker is the singleton tracker used by CPUCollector on Windows.
// It is initialized once and shared across all Collect calls.
var globalLoadTracker = &windowsLoadTracker{}

// record appends a new sample and prunes entries older than 15 minutes.
func (t *windowsLoadTracker) record(value float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	t.samples = append(t.samples, timedSample{value: value, at: now})

	// Prune samples older than 15 minutes — they are no longer needed
	cutoff := now.Add(-15 * time.Minute)
	start := 0
	for start < len(t.samples) && t.samples[start].at.Before(cutoff) {
		start++
	}
	t.samples = t.samples[start:]
}

// averageOver returns the mean of all samples within the given duration window.
// Returns 0 if no samples exist in the window.
func (t *windowsLoadTracker) averageOver(window time.Duration) float64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	cutoff := time.Now().Add(-window)
	var sum float64
	var count int

	for _, s := range t.samples {
		if s.at.After(cutoff) {
			sum += s.value
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

// collectLoadAvg queries the Windows WMI for the current processor queue length,
// records it in the rolling tracker, and populates the CPUStats load fields.
//
// Mapping to Unix load average semantics:
//   - Load1  → rolling average of queue length over the last 1 minute
//   - Load5  → rolling average of queue length over the last 5 minutes
//   - Load15 → rolling average of queue length over the last 15 minutes
//
// On the very first call (cold start), Load5 and Load15 will equal Load1
// since there is no historical data yet. They converge to true averages
// as samples accumulate over time.
func collectLoadAvg(stats *CPUStats) {
	var result []win32PerfRawDataPerfOSProcessorQueue

	// Query WMI for the current processor queue length.
	// The WHERE clause is required by WMI for this class.
	err := wmi.Query(
		"SELECT ProcessorQueueLength FROM Win32_PerfRawData_PerfOS_System",
		&result,
	)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to query WMI for processor queue length; skipping load averages")
		return
	}

	if len(result) == 0 {
		log.Warn().Msg("WMI returned no rows for processor queue length")
		return
	}

	queueLen := float64(result[0].ProcessorQueueLength)

	// Record the sample for rolling average computation
	globalLoadTracker.record(queueLen)

	// Populate load fields using rolling averages over standard Unix windows
	stats.Load1 = globalLoadTracker.averageOver(1 * time.Minute)
	stats.Load5 = globalLoadTracker.averageOver(5 * time.Minute)
	stats.Load15 = globalLoadTracker.averageOver(15 * time.Minute)
}
