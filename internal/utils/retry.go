package utils

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
)

// WithRetry wraps an operation in exponential backoff retry logic.
// The function fn should return the result and an error. If the error is
// a PermanentError (wrapped via backoff.Permanent), retries stop immediately.
// Use ClassifyError to wrap errors from API responses appropriately.
func WithRetry[T any](ctx context.Context, fn func() (T, error), opts ...backoff.RetryOption) (T, error) {
	defaults := []backoff.RetryOption{
		backoff.WithMaxTries(3),
		backoff.WithMaxElapsedTime(20 * time.Second),
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
	}
	allOpts := append(defaults, opts...)
	return backoff.Retry(ctx, fn, allOpts...)
}

// ClassifyError examines an API response and classifies the error as retryable
// or non-retryable. Retryable errors are returned as-is; non-retryable errors
// are wrapped with backoff.Permanent to stop retry attempts immediately.
//
// Retryable: 429 (rate limit), 5xx (server errors), 400 with "not in ready state"
// Non-retryable: all other client errors (401, 403, 404, 409, other 4xx)
func ClassifyError(resp Response, err error) error {
	if err == nil {
		return nil
	}

	// If we don't have a response, the error is likely a network issue — retryable
	if resp == nil {
		return err
	}

	statusCode := resp.StatusCode()

	// 429 and 5xx are retryable
	if statusCode == http.StatusTooManyRequests || statusCode >= 500 {
		return err
	}

	// 400 with "not in ready state" is retryable
	if statusCode == http.StatusBadRequest {
		if strings.Contains(err.Error(), "not in ready state") {
			return err
		}
	}

	// All other errors are non-retryable
	return backoff.Permanent(err)
}

// CheckClientResponseWithRetry combines CheckClientResponse with ClassifyError.
// It checks the response status code and classifies the resulting error for retry.
// Use this inside a WithRetry callback.
func CheckClientResponseWithRetry(resp Response, err error, statusCode int) error {
	checkErr := CheckClientResponse(resp, err, statusCode)
	if checkErr == nil {
		return nil
	}
	// If the original call had a transport error (resp is nil-ish or err != nil),
	// treat as retryable
	if err != nil {
		return checkErr
	}
	return ClassifyError(resp, checkErr)
}
