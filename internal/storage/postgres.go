package storage

import (
	"context"

	"github.com/google/uuid"
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
		result.MonitorID,
		result.URL,
		result.ResponseCode,
		result.Latency,
		result.Success,
		result.CheckedAt,
		result.FailureReason,
	)

	return err
}

func CreateMonitor(ctx context.Context, db *pgxpool.Pool, m monitor.Monitor) error {
	_, err := db.Exec(ctx, `
		INSERT INTO monitors (id, url, interval, timeout)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (url) DO NOTHING
	`,
		m.ID,
		m.URL,
		m.Interval,
		m.Timeout,
	)

	return err
}

func GetAllMonitors(ctx context.Context, db *pgxpool.Pool) ([]monitor.Monitor, error) {
	rows, err := db.Query(ctx, `
		SELECT id, url, interval, timeout
		FROM monitors
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monitors []monitor.Monitor

	for rows.Next() {
		var m monitor.Monitor

		err := rows.Scan(
			&m.ID,
			&m.URL,
			&m.Interval,
			&m.Timeout,
		)
		if err != nil {
			return nil, err
		}

		monitors = append(monitors, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return monitors, nil
}

func GetResults(
	ctx context.Context,
	db *pgxpool.Pool,
	monitorID uuid.UUID,
	limit int,
	offset int,
) ([]monitor.Result, error) {
	rows, err := db.Query(ctx, `
		SELECT
			url,
			monitor_id,
			response_code,
			latency,
			success,
			checked_at,
			failure_reason
		FROM results
		WHERE monitor_id = $1
		ORDER BY checked_at DESC
		LIMIT $2 OFFSET $3
	`, monitorID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []monitor.Result

	for rows.Next() {
		var result monitor.Result

		err := rows.Scan(
			&result.URL,
			&result.MonitorID,
			&result.ResponseCode,
			&result.Latency,
			&result.Success,
			&result.CheckedAt,
			&result.FailureReason,
		)
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func GetLastResult(
	ctx context.Context,
	db *pgxpool.Pool,
	monitorID uuid.UUID,
) (monitor.Result, error) {
	var result monitor.Result

	err := db.QueryRow(ctx, `
		SELECT
			url,
			monitor_id,
			response_code,
			latency,
			success,
			checked_at,
			failure_reason
		FROM results
		WHERE monitor_id = $1
		ORDER BY checked_at DESC
		LIMIT 1
	`, monitorID).Scan(
		&result.URL,
		&result.MonitorID,
		&result.ResponseCode,
		&result.Latency,
		&result.Success,
		&result.CheckedAt,
		&result.FailureReason,
	)

	if err != nil {
		return monitor.Result{}, err
	}

	return result, nil
}
