package monitor

import (
	"time"
)

type Result struct {
	URL           string        `json:"url"`
	ResponseCode  int           `json:"response_code"`
	Latency       time.Duration `json:"latency"`
	Success       bool          `json:"success"`
	CheckedAt     time.Time     `json:"checked_at"`
	FailureReason string        `json:"failure_reason"`
}
