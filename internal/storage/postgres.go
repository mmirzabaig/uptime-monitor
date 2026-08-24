package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmirzabaig/uptime-monitor/internal/monitor"
)

func NewPostgres(ctx context.Context) (*pgxpool.Pool, error) {
	connString := "postgres://postgres:password@localhost:5432/uptime_monitor"

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func StoreResult(
	ctx context.Context,
	db *pgxpool.Pool,
	result monitor.Result,
) error {
	_, err := db.Exec(ctx, `
		INSERT INTO results (
			monitor_id,
			url,
			response_code,
			latency,
			success,
			checked_at,
			failure_reason
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		result.ID,
		result.URL,
		result.ResponseCode,
		result.Latency,
		result.Success,
		result.CheckedAt,
		result.FailureReason,
	)

	return err
}
