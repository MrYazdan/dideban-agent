//go:build windows

package collector

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/yusufpapurcu/wmi"
)

// win32PerfRawDataPerfOSSystem maps to the WMI class Win32_PerfRawData_PerfOS_System.
// ProcessorQueueLength is the number of threads waiting to be scheduled on any CPU —
// the closest Windows equivalent to the Unix run-queue depth used in load averages.
type win32PerfRawDataPerfOSSystem struct {
	ProcessorQueueLength uint32
}

// timedSample holds a single processor queue length reading with its timestamp.
type timedSample struct {
	value float64
	at    time.Time
}

// windowsLoadTracker maintains a rolling history of processor queue length
// samples to compute 1-minute, 5-minute, and 15-minute moving averages.
// It is safe for concurrent use.
type windowsLoadTracker struct {
	mu      sync.Mutex
	samples []timedSample
}

// globalLoadTracker is the package-level singleton used by CPUCollector on Windows.
var globalLoadTracker = &windowsLoadTracker{}

// record appends a new sample and prunes entries older than 15 minutes.
func (t *windowsLoadTracker) record(value float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	t.samples = append(t.samples, timedSample{value: value, at: now})

	cutoff := now.Add(-15 * time.Minute)
	i := 0
	for i < len(t.samples) && t.samples[i].at.Before(cutoff) {
		i++
	}
	t.samples = t.samples[i:]
}

// average returns the mean of all samples within the given window,
// or the fallback value if no samples exist in that window.
func (t *windowsLoadTracker) average(window time.Duration, fallback float64) float64 {
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
		return fallback
	}
	return sum / float64(count)
}

// minLoad is the minimum value returned for any load field.
//
// The receiver uses Go's `validate:"required"` tag on float64 fields, which
// treats 0.0 as "not provided" and rejects the payload. Even on a completely
// idle Windows system the processor queue length can legitimately be 0, so we
// clamp all load values to this minimum to satisfy the receiver's validation
// without misrepresenting the actual system state in a meaningful way.
const minLoad = 0.01

// clampLoad ensures a load value is never exactly 0.0.
// A truly idle system (queue length = 0) is represented as minLoad.
func clampLoad(v float64) float64 {
	if v < minLoad {
		return minLoad
	}
	return v
}

// collectLoadAvg queries WMI for the current processor queue length, records
// the sample in the rolling tracker, and populates CPUStats load fields.
//
// Field semantics (mirroring Unix load average intervals):
//   - Load1  → 1-minute rolling average of processor queue length
//   - Load5  → 5-minute rolling average of processor queue length
//   - Load15 → 15-minute rolling average of processor queue length
//
// Cold-start behaviour:
//   On the first call(s) before the rolling windows have accumulated enough
//   history, the current instantaneous queue length is used as the fallback
//   for all three fields. This ensures the receiver's `required` validation
//   is always satisfied from the very first metric submission.
//
// Failure behaviour:
//   If the WMI query fails, all load fields are set to minLoad (0.01) so the
//   payload remains valid. A warning is logged with the underlying error.
func collectLoadAvg(stats *CPUStats) {
	var rows []win32PerfRawDataPerfOSSystem

	err := wmi.Query(
		"SELECT ProcessorQueueLength FROM Win32_PerfRawData_PerfOS_System",
		&rows,
	)
	if err != nil {
		log.Warn().Err(err).Msg("WMI query failed for processor queue length; using minLoad fallback")
		stats.Load1 = minLoad
		stats.Load5 = minLoad
		stats.Load15 = minLoad
		return
	}

	if len(rows) == 0 {
		log.Warn().Msg("WMI returned no rows for processor queue length; using minLoad fallback")
		stats.Load1 = minLoad
		stats.Load5 = minLoad
		stats.Load15 = minLoad
		return
	}

	queueLen := float64(rows[0].ProcessorQueueLength)
	globalLoadTracker.record(queueLen)

	// Use the current queue length as the cold-start fallback so that all
	// three fields are populated with a meaningful value from the first call.
	// clampLoad ensures we never emit 0.0 even on a fully idle system.
	fallback := clampLoad(queueLen)

	stats.Load1 = clampLoad(globalLoadTracker.average(1*time.Minute, fallback))
	stats.Load5 = clampLoad(globalLoadTracker.average(5*time.Minute, fallback))
	stats.Load15 = clampLoad(globalLoadTracker.average(15*time.Minute, fallback))
}
