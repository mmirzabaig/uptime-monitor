package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mmirzabaig/uptime-monitor/internal/monitor"
)

func Run(m []monitor.Monitor) {
	client := &http.Client{}

	for _, val := range m {

		go func(mm monitor.Monitor) {
			for {
				ctx, cancel := context.WithTimeout(context.Background(), mm.Timeout)
				_, err := mm.Check(ctx, client)
				cancel()
				if err != nil {
					fmt.Println(err)
					return
				}
				time.Sleep(mm.Interval)
			}
		}(val)
	}
	for {
		time.Sleep(time.Hour)
	}

}
