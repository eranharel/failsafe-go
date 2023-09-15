package test

import (
	"errors"
	"testing"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/fallback"
	"github.com/failsafe-go/failsafe-go/internal/testutil"
)

// Tests Fallback.WithResult
func TestFallbackOfResult(t *testing.T) {
	fb := fallback.WithResult(true)

	testutil.TestGetSuccess(t, failsafe.NewExecutor[bool](fb),
		func(execution failsafe.Execution[bool]) (bool, error) {
			return false, testutil.ErrInvalidArgument
		},
		1, 1, true)
}

// Tests Fallback.WithError
func TestShouldFallbackOfError(t *testing.T) {
	fb := fallback.WithError[bool](testutil.ErrInvalidArgument)

	testutil.TestGetFailure(t, failsafe.NewExecutor[bool](fb),
		func(execution failsafe.Execution[bool]) (bool, error) {
			return false, testutil.ErrInvalidArgument
		},
		1, 1, testutil.ErrInvalidArgument)
}

// Tests Fallback.WithFn
func TestShouldFallbackOfFn(t *testing.T) {
	fb := fallback.WithFn(func(exec failsafe.Execution[bool]) (bool, error) {
		return false, &testutil.CompositeError{
			Cause: exec.LastError(),
		}
	})

	testutil.TestGetFailure(t, failsafe.NewExecutor[bool](fb),
		func(execution failsafe.Execution[bool]) (bool, error) {
			return false, testutil.ErrConnecting
		},
		1, 1, &testutil.CompositeError{
			Cause: testutil.ErrConnecting,
		})
}

// Tests a successful execution that does not fallback
func TestShouldNotFallback(t *testing.T) {
	testutil.TestGetSuccess(t, failsafe.NewExecutor[bool](fallback.WithResult(true)),
		func(execution failsafe.Execution[bool]) (bool, error) {
			return false, nil
		},
		1, 1, false)
}

// Tests a fallback with failure conditions
func TestShouldFallbackWithFailureConditions(t *testing.T) {
	fb := fallback.BuilderWithResult[bool](true).
		HandleErrors(testutil.ErrInvalidState).
		Build()

	// Fallback should not handle
	testutil.TestGetFailure(t, failsafe.NewExecutor[bool](fb),
		func(execution failsafe.Execution[bool]) (bool, error) {
			return false, testutil.ErrInvalidArgument
		},
		1, 1, testutil.ErrInvalidArgument)

	// Fallback should handle
	testutil.TestGetSuccess(t, failsafe.NewExecutor[bool](fb),
		func(execution failsafe.Execution[bool]) (bool, error) {
			return false, testutil.ErrInvalidState
		},
		1, 1, true)
}

// Asserts that the fallback result itself can cause an execution to be considered a failure.
func TestShouldVerifyFallbackResult(t *testing.T) {
	// Assert a failure is still a failure
	fb := fallback.WithError[any](testutil.ErrInvalidArgument)
	testutil.TestGetFailure[any](t, failsafe.NewExecutor[any](fb),
		func(execution failsafe.Execution[any]) (any, error) {
			return false, errors.New("test")
		}, 1, 1, testutil.ErrInvalidArgument)

	// Assert a success after a failure is a success
	fb = fallback.WithResult[any](true)
	testutil.TestGetSuccess[any](t, failsafe.NewExecutor[any](fb),
		func(execution failsafe.Execution[any]) (any, error) {
			return false, errors.New("test")
		}, 1, 1, true)
}
