package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmirzabaig/uptime-monitor/internal/monitor"
	"github.com/mmirzabaig/uptime-monitor/internal/storage"
)

type Scheduler struct {
	db *pgxpool.Pool
}

func NewScheduler(db *pgxpool.Pool) *Scheduler {
	return &Scheduler{
		db: db,
	}
}

func (s *Scheduler) Run(m []monitor.Monitor) {
	for _, val := range m {
		s.AddMonitor(val)
	}

}

func (s *Scheduler) AddMonitor(mm monitor.Monitor) {
	go func() {
		client := &http.Client{}

		for {
			ctx, cancel := context.WithTimeout(
				context.Background(),
				mm.Timeout,
			)

			result, err := mm.Check(ctx, client)
			cancel()

			if err != nil {
				fmt.Println(err)
				time.Sleep(mm.Interval)
				continue
			}

			err = storage.StoreResult(
				context.Background(),
				s.db,
				result,
			)

			if err != nil {
				fmt.Println("failed to store result:", err)
			}

			time.Sleep(mm.Interval)
		}
	}()
}
