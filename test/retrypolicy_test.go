package test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/internal/policytesting"
	"github.com/failsafe-go/failsafe-go/internal/testutil"
	"github.com/failsafe-go/failsafe-go/retrypolicy"
)

// Tests a simple execution that retries.
func TestShouldRetryOnFailure(t *testing.T) {
	// Given
	rp := retrypolicy.WithDefaults[bool]()

	// When / Then
	testutil.TestGetFailure(t, failsafe.With[bool](rp),
		func(exec failsafe.Execution[bool]) (bool, error) {
			return false, testutil.ErrConnecting
		},
		3, 3, testutil.ErrConnecting)
}

func TestShouldReturnRetriesExceededError(t *testing.T) {
	// Given
	rp := retrypolicy.WithDefaults[bool]()

	// When / Then
	testutil.TestGetFailure(t, failsafe.With[bool](rp),
		func(exec failsafe.Execution[bool]) (bool, error) {
			return false, testutil.ErrConnecting
		},
		3, 3, &retrypolicy.RetriesExceededError{})
}

// Tests a simple execution that does not retry.
func TestShouldNotRetryOnSuccess(t *testing.T) {
	// Given
	rp := retrypolicy.WithDefaults[bool]()

	// When / Then
	testutil.TestGetSuccess(t, failsafe.With[bool](rp),
		func(exec failsafe.Execution[bool]) (bool, error) {
			return false, nil
		},
		1, 1, false)
}

// Asserts that a non-handled error does not trigger retries.
func TestShouldNotRetryOnNonRetriableFailure(t *testing.T) {
	// Given
	rp := retrypolicy.Builder[any]().
		WithMaxRetries(-1).
		HandleErrors(testutil.ErrConnecting).
		Build()

	// When / Then
	testutil.TestRunFailure(t, failsafe.With[any](rp),
		func(exec failsafe.Execution[any]) error {
			if exec.Attempts() <= 2 {
				return testutil.ErrConnecting
			}
			return testutil.ErrTimeout
		},
		3, 3, testutil.ErrTimeout)
}

// Asserts that an execution is failed when the max duration is exceeded.
func TestShouldCompleteWhenMaxDurationExceeded(t *testing.T) {
	// Given
	stats := &policytesting.Stats{}
	rp := policytesting.WithRetryStats(retrypolicy.Builder[bool]().
		HandleResult(false).
		WithMaxDuration(100*time.Millisecond), stats).
		Build()

	// When / Then
	testutil.TestGetFailure(t, failsafe.With[bool](rp),
		func(exec failsafe.Execution[bool]) (bool, error) {
			time.Sleep(120 * time.Millisecond)
			return false, errors.New("Asdf")
		},
		1, 1, &retrypolicy.RetriesExceededError{})
}

// Asserts that the ExecutionScheduledEvent.getDelay is as expected.
func TestScheduledRetryDelay(t *testing.T) {
	// Given
	delay := 10 * time.Millisecond
	rp := retrypolicy.Builder[any]().
		WithDelay(delay).
		OnRetryScheduled(func(e failsafe.ExecutionScheduledEvent[any]) {
			assert.Equal(t, delay, e.Delay)
		}).
		Build()

	// When / Then
	failsafe.With[any](rp).Run(func() error {
		return testutil.ErrConnecting
	})
}
