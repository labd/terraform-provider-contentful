package utils

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/cenkalti/backoff/v5"
)

type mockResponse struct {
	statusCode int
}

func (m *mockResponse) StatusCode() int {
	return m.statusCode
}

func TestClassifyError_NilError(t *testing.T) {
	resp := &mockResponse{statusCode: 200}
	if err := ClassifyError(resp, nil); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestClassifyError_NilResponse(t *testing.T) {
	err := errors.New("network error")
	result := ClassifyError(nil, err)
	if result == nil {
		t.Fatal("expected error, got nil")
	}
	// Should be retryable (not permanent)
	var permanent *backoff.PermanentError
	if errors.As(result, &permanent) {
		t.Fatal("expected retryable error, got permanent")
	}
}

func TestClassifyError_429IsRetryable(t *testing.T) {
	resp := &mockResponse{statusCode: http.StatusTooManyRequests}
	err := errors.New("rate limited")
	result := ClassifyError(resp, err)
	var permanent *backoff.PermanentError
	if errors.As(result, &permanent) {
		t.Fatal("expected retryable error, got permanent")
	}
}

func TestClassifyError_500IsRetryable(t *testing.T) {
	resp := &mockResponse{statusCode: http.StatusInternalServerError}
	err := errors.New("server error")
	result := ClassifyError(resp, err)
	var permanent *backoff.PermanentError
	if errors.As(result, &permanent) {
		t.Fatal("expected retryable error, got permanent")
	}
}

func TestClassifyError_400NotInReadyStateIsRetryable(t *testing.T) {
	resp := &mockResponse{statusCode: http.StatusBadRequest}
	err := errors.New("resource not in ready state")
	result := ClassifyError(resp, err)
	var permanent *backoff.PermanentError
	if errors.As(result, &permanent) {
		t.Fatal("expected retryable error, got permanent")
	}
}

func TestClassifyError_400OtherIsNonRetryable(t *testing.T) {
	resp := &mockResponse{statusCode: http.StatusBadRequest}
	err := errors.New("validation failed")
	result := ClassifyError(resp, err)
	var permanent *backoff.PermanentError
	if !errors.As(result, &permanent) {
		t.Fatal("expected permanent error, got retryable")
	}
}

func TestClassifyError_404IsNonRetryable(t *testing.T) {
	resp := &mockResponse{statusCode: http.StatusNotFound}
	err := errors.New("not found")
	result := ClassifyError(resp, err)
	var permanent *backoff.PermanentError
	if !errors.As(result, &permanent) {
		t.Fatal("expected permanent error, got retryable")
	}
}

func TestWithRetry_SuccessOnFirstAttempt(t *testing.T) {
	calls := 0
	result, err := WithRetry(context.Background(), func() (string, error) {
		calls++
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ok" {
		t.Fatalf("expected 'ok', got '%s'", result)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestWithRetry_RetriesOnTransientError(t *testing.T) {
	calls := 0
	result, err := WithRetry(context.Background(), func() (string, error) {
		calls++
		if calls < 3 {
			return "", errors.New("transient")
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ok" {
		t.Fatalf("expected 'ok', got '%s'", result)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestWithRetry_StopsOnPermanentError(t *testing.T) {
	calls := 0
	_, err := WithRetry(context.Background(), func() (string, error) {
		calls++
		return "", backoff.Permanent(errors.New("fatal"))
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestCheckClientResponseWithRetry_Success(t *testing.T) {
	resp := &mockResponse{statusCode: 200}
	err := CheckClientResponseWithRetry(resp, nil, 200)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestCheckClientResponseWithRetry_TransportError(t *testing.T) {
	err := CheckClientResponseWithRetry(&mockResponse{statusCode: 0}, errors.New("connection refused"), 200)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var permanent *backoff.PermanentError
	if errors.As(err, &permanent) {
		t.Fatal("expected retryable error, got permanent")
	}
}

func TestCheckClientResponseWithRetry_404IsNonRetryable(t *testing.T) {
	resp := &mockResponse{statusCode: 404}
	err := CheckClientResponseWithRetry(resp, nil, 200)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var permanent *backoff.PermanentError
	if !errors.As(err, &permanent) {
		t.Fatal("expected permanent error, got retryable")
	}
}

func TestCheckClientResponseWithRetry_500IsRetryable(t *testing.T) {
	resp := &mockResponse{statusCode: 500}
	err := CheckClientResponseWithRetry(resp, nil, 200)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var permanent *backoff.PermanentError
	if errors.As(err, &permanent) {
		t.Fatal("expected retryable error, got permanent")
	}
}
