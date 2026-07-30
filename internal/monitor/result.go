package monitor

import (
	"time"
)

type Result struct {
	URL           string
	ResponseCode  int
	Latency       time.Duration
	Success       bool
	CheckedAt     time.Time
	FailureReason string
}
