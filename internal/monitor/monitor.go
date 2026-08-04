package monitor

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Monitor struct {
	URL      string        `json:"url"`
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
	ID       string        `json:"id"`
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
			ID:       "a1",
		},
		{
			URL:      "https://youtube.com",
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
			ID:       "a2",
		},
		{
			URL:      "https://go.dev/tour/list",
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
			ID:       "a3",
		},
		{
			URL:      "https://amazon.com",
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
			ID:       "a4",
		},
	},
}

func (m Monitor) Check(ctx context.Context, client *http.Client) (Result, error) {
	result := Result{}

	result.URL = m.URL
	result.ID = m.ID

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.URL, nil)

	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		result.FailureReason = err.Error()
	}

	start := time.Now()
	result.CheckedAt = start

	resp, err := client.Do(req)
	if ctx.Err() == context.DeadlineExceeded {
		result.FailureReason = "Timed Out"
		return result, err
	}
	if err != nil {
		// If the timeout is reached, client.Do returns an error
		fmt.Printf("Request failed (possibly due to timeout): %v\n", err)
		return result, err
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

func AddMonitor(m Monitor) {
	monitors.mu.Lock()
	defer monitors.mu.Unlock()
	monitors.monitors = append(monitors.monitors, m)
}
func AllMonitors() []Monitor {
	monitors.mu.RLock()
	defer monitors.mu.RUnlock()
	return monitors.monitors
}
