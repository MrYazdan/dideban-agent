//go:build !windows

package collector

import (
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/load"
)

// collectLoadAvg retrieves the system load averages on Unix-like systems
// (Linux, macOS, BSD) using /proc/loadavg or the equivalent syscall.
//
// Load averages represent the average number of runnable processes
// over the last 1, 5, and 15 minutes respectively.
//
// On exotic Linux configurations (e.g., some containers without /proc/loadavg),
// this call may fail. In that case a warning is logged and the fields remain
// at their zero values (omitted from the JSON payload via omitempty).
func collectLoadAvg(stats *CPUStats) {
	avg, err := load.Avg()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to collect load averages; skipping")
		return
	}

	stats.Load1 = avg.Load1
	stats.Load5 = avg.Load5
	stats.Load15 = avg.Load15
}
