package monitor

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Monitor struct {
	URL      string        `json:"url"`
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
}

func (m Monitor) Check(ctx context.Context, client *http.Client) (Result, error) {
	result := Result{}

	result.URL = m.URL

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
