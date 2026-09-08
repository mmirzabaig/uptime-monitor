package monitor

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Monitor struct {
	URL      string        `json:"url"`
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
	ID       uuid.UUID     `json:"id"`
}

type MonitorStorage struct {
	mu       sync.RWMutex
	monitors []Monitor
}

var monitors = MonitorStorage{
	monitors: []Monitor{
		{
			URL:      "https://google.com",
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
			ID:       uuid.New(),
		},
		{
			URL:      "https://youtube.com",
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
			ID:       uuid.New(),
		},
		{
			URL:      "https://go.dev/tour/list",
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
			ID:       uuid.New(),
		},
		{
			URL:      "https://amazon.com",
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
			ID:       uuid.New(),
		},
		{
			URL:      "https://open.spotify.com/",
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
			ID:       uuid.New(),
		},
	},
}

func (m Monitor) Check(ctx context.Context, client *http.Client) (Result, error) {
	result := Result{}

	result.URL = m.URL
	result.MonitorID = m.ID

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.URL, nil)

	if err != nil {
		result.FailureReason = err.Error()
		return result, err
	}

	start := time.Now()
	result.CheckedAt = start

	resp, err := client.Do(req)
	if err != nil {
		result.Latency = time.Since(start)

		if ctx.Err() == context.DeadlineExceeded {
			result.FailureReason = "Timed Out"
		} else {
			result.FailureReason = err.Error()
		}

		result.Success = false

		return result, nil
	}

	defer resp.Body.Close()

	latency := time.Since(start)
	result.ResponseCode = resp.StatusCode
	result.Latency = latency
	result.Success = resp.StatusCode >= 200 && resp.StatusCode < 400
	if !result.Success {
		result.FailureReason = strconv.Itoa(result.ResponseCode)
	}
	return result, nil
}

func AllMonitors() []Monitor {
	monitors.mu.RLock()
	defer monitors.mu.RUnlock()
	// return a copy(slice) of monitors isntead of returning the actual monitors object so no one can modify it
	return append([]Monitor(nil), monitors.monitors...)
}
