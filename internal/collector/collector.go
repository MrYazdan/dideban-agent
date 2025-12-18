package collector

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// MetricCollector defines a common interface for all metric collectors
// (CPU, Memory, Disk, etc).
//
// Each collector:
//   - Must provide a name (used for logging and debugging)
//   - Must collect its own metrics and populate the shared Metrics struct
type MetricCollector interface {
	// Name returns the unique name of the collector (e.g. "cpu", "memory").
	Name() string

	// Collect gathers metrics and writes them into the provided Metrics struct.
	// The context should be respected for cancellation and timeouts.
	Collect(ctx context.Context, m *Metrics) error
}

// Collector orchestrates all MetricCollector implementations.
// It is responsible for running collectors concurrently,
// handling errors, and measuring collection duration.
type Collector struct {
	collectors []MetricCollector
}

// New creates and initializes a new Collector instance
// with all default system metric collectors registered.
func New() *Collector {
	return &Collector{
		collectors: []MetricCollector{
			&CPUCollector{},
			&MemoryCollector{},
			&DiskCollector{},
		},
	}
}

// CollectAll runs all registered metric collectors concurrently.
// Partial results are returned even if one or more collectors fail.
//
// Returns:
//   - *Metrics: collected metrics (maybe partial)
//   - error: combined error if any collector fails
func (c *Collector) CollectAll(ctx context.Context, agentID string) (*Metrics, error) {
	start := time.Now()

	metrics := &Metrics{
		AgentID:   agentID,
		Timestamp: time.Now().UnixMilli(),
	}

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	// Run each collector in its own goroutine
	for _, col := range c.collectors {
		wg.Add(1)

		go func(col MetricCollector) {
			defer wg.Done()

			if err := col.Collect(ctx, metrics); err != nil {
				log.Warn().
					Err(err).
					Str("collector", col.Name()).
					Msg("metric collection failed")

				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(col)
	}

	// Wait for all collectors to finish
	wg.Wait()

	// Measure total collection duration
	metrics.CollectDuration = time.Since(start).Milliseconds()

	// Return partial metrics with a combined error (if any)
	if len(errs) > 0 {
		return metrics, errors.Join(errs...)
	}

	return metrics, nil
}

// Metrics represents a snapshot of all collected system metrics.
type Metrics struct {
	AgentID         string `json:"agent_id"`
	Timestamp       int64  `json:"timestamp_ms"`
	CollectDuration int64  `json:"collect_duration_ms"`

	CPU    CPUStats  `json:"cpu"`
	Memory MemStats  `json:"memory"`
	Disk   DiskStats `json:"disk"`
}
