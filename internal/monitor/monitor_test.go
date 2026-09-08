package monitor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMonitorCheckSuccess(t *testing.T) {
	// Arrange
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	defer server.Close()

	monitor := Monitor{
		URL: server.URL,
		ID:  uuid.New(),
	}

	client := &http.Client{}

	// Act
	result, err := monitor.Check(
		context.Background(),
		client,
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected monitor check to succeed")
	}

	if result.ResponseCode != http.StatusOK {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusOK,
			result.ResponseCode,
		)
	}

	if result.URL != server.URL {
		t.Errorf(
			"expected URL %s, got %s",
			server.URL,
			result.URL,
		)
	}

	if result.MonitorID != monitor.ID {
		t.Errorf(
			"expected monitor ID %s, got %s",
			monitor.ID,
			result.MonitorID,
		)
	}

	if result.Latency <= 0 {
		t.Error("expected latency to be greater than 0")
	}

	if result.CheckedAt.IsZero() {
		t.Error("expected CheckedAt to be set")
	}
}

func TestMonitorCheckHTTPFailure(t *testing.T) {
	// Arrange
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}),
	)
	defer server.Close()

	m := Monitor{
		URL: server.URL,
		ID:  uuid.New(),
	}

	client := &http.Client{}

	// Act
	result, err := m.Check(context.Background(), client)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Success {
		t.Error("expected monitor check to fail")
	}

	if result.ResponseCode != http.StatusNotFound {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusNotFound,
			result.ResponseCode,
		)
	}

	expectedReason := "404"

	if result.FailureReason != expectedReason {
		t.Errorf(
			"expected failure reason %s, got %s",
			expectedReason,
			result.FailureReason,
		)
	}
}

func TestMonitorCheckTimeout(t *testing.T) {
	// Arrange
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)

			w.WriteHeader(http.StatusOK)
		}),
	)
	defer server.Close()

	m := Monitor{
		URL: server.URL,
		ID:  uuid.New(),
	}

	client := &http.Client{}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Millisecond,
	)
	defer cancel()

	// Act
	result, err := m.Check(ctx, client)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Success {
		t.Error("expected monitor check to fail")
	}

	if result.FailureReason != "Timed Out" {
		t.Errorf(
			"expected failure reason %q, got %q",
			"Timed Out",
			result.FailureReason,
		)
	}
}

func TestMonitorCheckInvalidURL(t *testing.T) {
	m := Monitor{
		URL: "://invalid-url",
		ID:  uuid.New(),
	}

	result, err := m.Check(
		context.Background(),
		&http.Client{},
	)

	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if result.Success {
		t.Error("expected monitor check to fail")
	}

	if result.FailureReason == "" {
		t.Error("expected failure reason to be set")
	}
}
