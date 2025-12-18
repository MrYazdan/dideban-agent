package sender

import (
	"context"
	"time"

	"dideban-agent/internal/collector"

	"github.com/rs/zerolog/log"
)

// MockSender implements the Sender interface for development and testing.
// It simulates network delays and provides configurable failure scenarios.
type MockSender struct {
	config MockConfig
}

// MockConfig contains configuration for mock sender behavior.
type MockConfig struct {
	// Simulated network delay
	Delay time.Duration

	// Failure rate (0.0 = never fail, 1.0 = always fail)
	FailureRate float64

	// Enable verbose logging of sent metrics
	VerboseLogging bool
}

// DefaultMockConfig returns sensible defaults for development.
func DefaultMockConfig() MockConfig {
	return MockConfig{
		Delay:          100 * time.Millisecond,
		FailureRate:    0.0,
		VerboseLogging: true,
	}
}

// NewMockSender creates a new mock sender for development use.
func NewMockSender(config MockConfig) *MockSender {
	return &MockSender{
		config: config,
	}
}

// Send simulates sending metrics with configurable delay and failure rate.
func (m *MockSender) Send(ctx context.Context, metrics *collector.Metrics) error {
	// Simulate network delay
	if m.config.Delay > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(m.config.Delay):
		}
	}

	// Log metrics if verbose logging is enabled
	if m.config.VerboseLogging {
		log.Info().
			Str("agent_id", metrics.AgentID).
			Int64("timestamp", metrics.Timestamp).
			Float64("cpu_usage_percent", metrics.CPU.UsagePercent).
			Float64("memory_usage_percent", metrics.Memory.UsagePercent).
			Float64("disk_usage_percent", metrics.Disk.UsagePercent).
			Int64("collect_duration_ms", metrics.CollectDuration).
			Msg("🧪 Mock sender:")
	}

	return nil
}

// Close is a no-op for the mock sender.
func (m *MockSender) Close() error {
	return nil
}
