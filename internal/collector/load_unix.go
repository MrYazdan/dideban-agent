//go:build !windows

package collector

import (
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/load"
)

// minLoad is the minimum value emitted for any load field.
//
// The receiver uses Go's `validate:"required"` tag on float64 fields, which
// treats 0.0 as "not provided" and rejects the payload. On a completely idle
// system, load averages can legitimately be 0.0, so we clamp to this minimum
// to satisfy validation without misrepresenting the system state.
const minLoad = 0.01

// clampLoad ensures a load value is never exactly 0.0.
func clampLoad(v float64) float64 {
	if v < minLoad {
		return minLoad
	}
	return v
}

// collectLoadAvg retrieves system load averages on Unix-like systems
// (Linux, macOS, BSD) via /proc/loadavg or the equivalent syscall.
//
// Load averages represent the average number of runnable processes over
// the last 1, 5, and 15 minutes respectively.
//
// All values are clamped to minLoad (0.01) to satisfy the receiver's
// `required` validation on float64 fields, which rejects 0.0 as absent.
//
// On failure (e.g., containers without /proc/loadavg), a warning is logged
// and all fields are set to minLoad so the payload remains valid.
func collectLoadAvg(stats *CPUStats) {
	avg, err := load.Avg()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to collect load averages; using minLoad fallback")
		stats.Load1 = minLoad
		stats.Load5 = minLoad
		stats.Load15 = minLoad
		return
	}

	stats.Load1 = clampLoad(avg.Load1)
	stats.Load5 = clampLoad(avg.Load5)
	stats.Load15 = clampLoad(avg.Load15)
}
