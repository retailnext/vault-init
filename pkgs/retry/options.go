package retry

import (
	"context"
	"time"
)

type Config struct {
	attempts uint
	delay    time.Duration
	context  context.Context
	retryIf  RetryIfFunc
}

// Option represents an option for retry.
type Option func(*Config)

type RetryIfFunc func(error) bool

func emptyOption(c *Config) {}

func RetryIf(retryIf RetryIfFunc) Option {
	if retryIf == nil {
		return emptyOption
	}
	return func(c *Config) {
		c.retryIf = retryIf
	}
}

func Delay(delay time.Duration) Option {
	return func(c *Config) {
		c.delay = delay
	}
}

func Attempts(attempts uint) Option {
	return func(c *Config) {
		c.attempts = attempts
	}
}
