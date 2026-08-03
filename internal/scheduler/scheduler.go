package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mmirzabaig/uptime-monitor/internal/monitor"
	"github.com/mmirzabaig/uptime-monitor/internal/storage"
)

func Run(m []monitor.Monitor) {
	for _, val := range m {
		AddMonitor(val)
	}

}

func AddMonitor(m monitor.Monitor) {
	client := &http.Client{}
	go func(mm monitor.Monitor) {
		for {
			ctx, cancel := context.WithTimeout(context.Background(), mm.Timeout)
			result, err := mm.Check(ctx, client)
			cancel()
			if err != nil {
				fmt.Println(err)
				return
			}

			storage.StoreToMemory(result)

			time.Sleep(mm.Interval)
		}
	}(m)
}
