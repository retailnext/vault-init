package retry

import (
	"context"
	"log/slog"
	"time"
)

// Function signature of retryable function
type RetryableFunc func() error

func Do(retryableFunc RetryableFunc, opts ...Option) (err error) {
	var n uint

	// default
	config := newDefaultRetryConfig()

	// apply opts
	for _, opt := range opts {
		opt(config)
	}

	if err := config.context.Err(); err != nil {
		return err
	}

	for n < config.attempts {
		err = retryableFunc()
		if err == nil {
			break
		}

		if !config.retryIf(err) {
			break
		}

		// last attempt. exit right away
		if n == config.attempts-1 {
			break
		}
		slog.Info("iteration_err_result", "err", err)

		select {
		case <-time.After(config.delay):
		case <-config.context.Done():
			return config.context.Err()
		}

		n++
	}
	return err
}

func newDefaultRetryConfig() *Config {
	return &Config{
		attempts: uint(10),
		delay:    100 * time.Millisecond,
		retryIf:  AlwaysRetry,
		context:  context.Background(),
	}
}

func AlwaysRetry(_ error) bool {
	return true
}
