package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmirzabaig/uptime-monitor/internal/monitor"
	"github.com/mmirzabaig/uptime-monitor/internal/storage"
)

type Scheduler struct {
	db  *pgxpool.Pool
	ctx context.Context
	wg  sync.WaitGroup
}

func (s *Scheduler) Wait() {
	s.wg.Wait()
}

func NewScheduler(ctx context.Context, db *pgxpool.Pool) *Scheduler {
	return &Scheduler{
		ctx: ctx,
		db:  db,
	}
}

// func (s *Scheduler) Run(m []monitor.Monitor) {
// 	for _, val := range m {
// 		s.AddMonitor(val)
// 	}

// }

func (s *Scheduler) AddMonitor(mm monitor.Monitor) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		client := &http.Client{}

		ticker := time.NewTicker(mm.Interval)
		defer ticker.Stop()

		for {
			checkCtx, cancel := context.WithTimeout(s.ctx, mm.Timeout)

			result, err := mm.Check(checkCtx, client)

			cancel()

			if err != nil {
				fmt.Println(err)
			} else {
				err = storage.StoreResult(
					s.ctx,
					s.db,
					result,
				)

				if err != nil {
					fmt.Println("failed to store result:", err)
				}
			}

			select {
			case <-s.ctx.Done():
				return

			case <-ticker.C:
				// Continue to next iteration
			}
		}
	}()
}
