package retry

import (
	"fmt"
	"math"
	"time"

	"github.com/114windd/restapi/internal/logger"
	"github.com/sirupsen/logrus"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// DefaultRetryConfig returns sensible defaults for database operations
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    2 * time.Second,
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func() error

// ExecuteWithRetry executes a function with exponential backoff retry logic
func ExecuteWithRetry(operation string, fn RetryableFunc, config RetryConfig) error {
	var lastErr error

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Log the attempt
		LogRetry(operation, attempt, config.MaxAttempts).Debug("Executing operation")

		// Execute the function
		err := fn()
		if err == nil {
			// Success
			if attempt > 1 {
				LogRetry(operation, attempt, config.MaxAttempts).Info("Operation succeeded after retry")
			}
			return nil
		}

		lastErr = err

		// Don't sleep on the last attempt
		if attempt == config.MaxAttempts {
			LogRetry(operation, attempt, config.MaxAttempts).WithError(err).Error("Operation failed after all retries")
			break
		}

		// Calculate delay with exponential backoff
		delay := calculateDelay(attempt, config.BaseDelay, config.MaxDelay)

		LogRetry(operation, attempt, config.MaxAttempts).
			WithError(err).
			WithField("retry_delay_ms", delay.Milliseconds()).
			Warn("Operation failed, retrying")

		time.Sleep(delay)
	}

	return fmt.Errorf("operation '%s' failed after %d attempts: %w", operation, config.MaxAttempts, lastErr)
}

// calculateDelay calculates exponential backoff delay
func calculateDelay(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	// Exponential backoff: baseDelay * 2^(attempt-1)
	delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt-1)))

	// Cap at maxDelay
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

// LogRetry creates a structured log entry for retry operations
func LogRetry(operation string, attempt, maxAttempts int) *logrus.Entry {
	return logger.Log.WithFields(logrus.Fields{
		"operation":    operation,
		"attempt":      attempt,
		"max_attempts": maxAttempts,
		"type":         "retry",
	})
}
