package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmirzabaig/uptime-monitor/internal/monitor"
)

func NewPostgres(ctx context.Context) (*pgxpool.Pool, error) {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
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

func CreateMonitor(
	ctx context.Context,
	db *pgxpool.Pool,
	m monitor.Monitor,
) (bool, error) {
	var id uuid.UUID

	err := db.QueryRow(ctx, `
		INSERT INTO monitors (id, url, interval, timeout)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (url) DO NOTHING
		RETURNING id
	`,
		m.ID,
		m.URL,
		m.Interval,
		m.Timeout,
	).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
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

func MonitorExists(
	ctx context.Context,
	db *pgxpool.Pool,
	monitorID uuid.UUID,
) (bool, error) {
	var exists bool

	err := db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM monitors
			WHERE id = $1
		)
	`, monitorID).Scan(&exists)

	return exists, err
}
