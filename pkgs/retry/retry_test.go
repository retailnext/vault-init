package retry

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDoRetryAlways(t *testing.T) {
	count := 0
	startTime := time.Now()
	err := Do(func() error {
		count++
		if count < 3 {
			return fmt.Errorf("less than 3")
		}
		return nil
	},
		RetryIf(func(_ error) bool {
			// retry always
			return true
		}),
		Delay(time.Second),
	)
	duration := math.Round(time.Since(startTime).Seconds())
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
	assert.Equal(t, 2.0, duration)
}

func TestDoRetryCertainError(t *testing.T) {
	count := 0
	errorRetry := "do it"
	errorDontRetry := "don't do it"
	err := Do(func() error {
		count++
		if count < 5 {
			return fmt.Errorf("%s", errorRetry)
		}
		return fmt.Errorf("%s", errorDontRetry)
	},
		RetryIf(func(err error) bool {
			return err.Error() == errorRetry
		}),
	)
	assert.Equal(t, err.Error(), errorDontRetry)
	assert.Equal(t, 5, count)
}

func TestDoRetryMax(t *testing.T) {
	count := 0
	errorMsg := "error msg"
	err := Do(func() error {
		count++
		return fmt.Errorf("%s", errorMsg)
	},
		Attempts(10),
	)
	assert.Equal(t, err.Error(), errorMsg)
	assert.Equal(t, 10, count)
}
