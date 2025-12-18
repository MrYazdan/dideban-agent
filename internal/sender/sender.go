package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"dideban-agent/internal/collector"

	"github.com/rs/zerolog/log"
)

// Sender defines the interface for sending metrics to remote endpoints.
// Implementations must handle retries, timeouts, and error recovery gracefully.
type Sender interface {
	// Send transmits metrics to the configured endpoint.
	// Returns an error if all retry attempts fail.
	Send(ctx context.Context, metrics *collector.Metrics) error

	// Close releases any resources held by the sender.
	// Should be called during application shutdown.
	Close() error
}

// HTTPSender implements the Sender interface using HTTP transport.
// It provides production-ready features including:
//   - Exponential backoff retry mechanism
//   - Configurable timeouts and connection pooling
//   - Structured error logging and metrics
//   - Graceful degradation on failures
type HTTPSender struct {
	client   *http.Client
	endpoint string
	token    string
	config   HTTPConfig
}

// HTTPConfig contains configuration parameters for HTTP sender behavior.
type HTTPConfig struct {
	// Maximum number of retry attempts for failed requests
	MaxRetries int

	// Initial retry delay (doubled on each retry)
	InitialRetryDelay time.Duration

	// Maximum retry delay (caps exponential backoff)
	MaxRetryDelay time.Duration

	// HTTP request timeout
	RequestTimeout time.Duration

	// HTTP client timeout (includes connection establishment)
	ClientTimeout time.Duration
}

// NewHTTPSender creates a new HTTP sender with the specified configuration.
// The sender is ready for immediate use and includes connection pooling.
func NewHTTPSender(endpoint, token string, config HTTPConfig) *HTTPSender {
	// Configure HTTP client with connection pooling and timeouts
	client := &http.Client{
		Timeout: config.ClientTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 2,
			IdleConnTimeout:     60 * time.Second,
		},
	}

	return &HTTPSender{
		client:   client,
		endpoint: endpoint,
		token:    token,
		config:   config,
	}
}

// Send transmits metrics to the configured endpoint with retry logic.
// The method implements exponential backoff and respects context cancellation.
func (s *HTTPSender) Send(ctx context.Context, metrics *collector.Metrics) error {
	// Serialize metrics to JSON
	payload, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	// Execute request with retry logic
	return s.sendWithRetry(ctx, payload)
}

// sendWithRetry implements the core retry logic with exponential backoff.
func (s *HTTPSender) sendWithRetry(ctx context.Context, payload []byte) error {
	var lastErr error
	retryDelay := s.config.InitialRetryDelay

	for attempt := 0; attempt <= s.config.MaxRetries; attempt++ {
		// Check for context cancellation before each attempt
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute HTTP request
		err := s.executeRequest(ctx, payload)
		if err == nil {
			// Success - log and return
			if attempt > 0 {
				log.Info().
					Int("attempts", attempt+1).
					Msg("📤 Metrics sent successfully after retries")
			}
			return nil
		}

		lastErr = err

		// Log retry attempt (except for the last failed attempt)
		if attempt < s.config.MaxRetries {
			log.Warn().
				Err(err).
				Int("attempt", attempt+1).
				Int("max_retries", s.config.MaxRetries).
				Dur("retry_delay", retryDelay).
				Msg("🔄 Request failed, retrying")

			// Wait before next retry (with context cancellation support)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay):
			}

			// Exponential backoff with maximum cap
			retryDelay *= 2
			if retryDelay > s.config.MaxRetryDelay {
				retryDelay = s.config.MaxRetryDelay
			}
		}
	}

	// All retries exhausted
	log.Error().
		Err(lastErr).
		Int("max_retries", s.config.MaxRetries).
		Msg("❌ Failed to send metrics after all retries")

	return fmt.Errorf("failed to send metrics after %d retries: %w", s.config.MaxRetries, lastErr)
}

// executeRequest performs a single HTTP request attempt.
func (s *HTTPSender) executeRequest(ctx context.Context, payload []byte) error {
	// Create request with timeout context
	reqCtx, cancel := context.WithTimeout(ctx, s.config.RequestTimeout)
	defer cancel()

	log.Debug().Str("endpoint", s.endpoint).Msg("Creating HTTP request")

	req, err := http.NewRequestWithContext(reqCtx, "POST", s.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("User-Agent", "dideban-agent/0.1.0")

	log.Debug().Msg("Executing HTTP request")

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	log.Debug().Int("status_code", resp.StatusCode).Msg("Received HTTP response")

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Success - drain response body to enable connection reuse
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	// Read error response for debugging
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
}

// Close releases resources held by the HTTP sender.
// This method should be called during application shutdown.
func (s *HTTPSender) Close() error {
	if s.client != nil {
		s.client.CloseIdleConnections()
	}
	return nil
}
