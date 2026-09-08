package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmirzabaig/uptime-monitor/internal/monitor"
)

// --------------------Health Handler Tests-------------------
func TestHealthHandlerInvalidUUID(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(
		http.MethodGet,
		"/health/not-a-uuid",
		nil,
	)

	recorder := httptest.NewRecorder()

	api := API{}

	// Act
	api.healthHandler(recorder, req)

	// Assert
	if recorder.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func testDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()

	databaseURL := "postgres://postgres:password@localhost:5432/uptime_monitor_test?sslmode=disable"

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to create database pool: %v", err)
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		t.Fatalf("failed to ping test database: %v", err)
	}

	return db
}

func TestHealthHandlerSuccess(t *testing.T) {
	db := testDB(t)

	// this function runs at the end of the function after everything is done running or after return.
	t.Cleanup(func() {
		_, err := db.Exec(
			context.Background(),
			"DELETE FROM results",
		)
		if err != nil {
			t.Errorf("failed to clean up results: %v", err)
		}

		_, err = db.Exec(
			context.Background(),
			"DELETE FROM monitors",
		)
		if err != nil {
			t.Errorf("failed to clean up monitors: %v", err)
		}
		defer db.Close()
	})

	ctx := context.Background()

	// Create test monitor
	monitorID := uuid.New()

	_, err := db.Exec(ctx, `
		INSERT INTO monitors (
			id,
			url,
			interval,
			timeout
		)
		VALUES ($1, $2, $3, $4)
	`,
		monitorID,
		"https://example.com",
		int64(30*time.Second),
		int64(5*time.Second),
	)

	if err != nil {
		t.Fatalf("failed to insert test monitor: %v", err)
	}

	// Create test result
	_, err = db.Exec(ctx, `
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
		monitorID,
		"https://example.com",
		200,
		int64(50*time.Millisecond),
		true,
		time.Now(),
		"",
	)

	if err != nil {
		t.Fatalf("failed to insert test result: %v", err)
	}

	// Create API
	api := API{
		db: db,
	}

	// Create request
	req := httptest.NewRequest(
		http.MethodGet,
		"/health/"+monitorID.String(),
		nil,
	)

	req.SetPathValue("id", monitorID.String())

	recorder := httptest.NewRecorder()

	// Call handler
	api.healthHandler(recorder, req)

	// Check status
	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}
}

func TestHealthHandlerMonitorNotFound(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	// This UUID is valid, but we haven't inserted a monitor with it.
	monitorID := uuid.New()

	api := API{
		db: db,
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/health/"+monitorID.String(),
		nil,
	)

	req.SetPathValue("id", monitorID.String())

	recorder := httptest.NewRecorder()

	api.healthHandler(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusNotFound,
			recorder.Code,
		)
	}
}

// ---------------------- Results per ID Tests ---------------------------------
func TestGetResultsInvalidLimit(t *testing.T) {
	monitorID := uuid.New()

	api := API{}

	req := httptest.NewRequest(
		http.MethodGet,
		"/results/"+monitorID.String()+"?limit=abc",
		nil,
	)

	req.SetPathValue("id", monitorID.String())

	recorder := httptest.NewRecorder()

	// Act
	api.getResultsHandler(recorder, req)

	// Assert
	if recorder.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}
func TestGetResultsSuccess(t *testing.T) {
	db := testDB(t)

	t.Cleanup(func() {
		ctx := context.Background()

		_, err := db.Exec(ctx, "DELETE FROM results")
		if err != nil {
			t.Errorf("failed to clean up results: %v", err)
		}

		_, err = db.Exec(ctx, "DELETE FROM monitors")
		if err != nil {
			t.Errorf("failed to clean up monitors: %v", err)
		}

		db.Close()
	})

	ctx := context.Background()
	monitorID := uuid.New()

	_, err := db.Exec(ctx, `
		INSERT INTO monitors (
			id,
			url,
			interval,
			timeout
		)
		VALUES ($1, $2, $3, $4)
	`,
		monitorID,
		"https://example.com",
		int64(30*time.Second),
		int64(5*time.Second),
	)

	if err != nil {
		t.Fatalf("failed to insert test monitor: %v", err)
	}

	// Create test result
	for i := 0; i < 5; i++ {
		_, err = db.Exec(ctx, `
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
			monitorID,
			"https://example.com",
			200+i,
			int64(50*time.Millisecond),
			true,
			time.Now().Add(time.Duration(i)*time.Second),
			"",
		)

		if err != nil {
			t.Fatalf("failed to insert test result: %v", err)
		}
	}
	// Create API
	api := API{
		db: db,
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/results/"+monitorID.String()+"?limit=2&offset=2",
		nil,
	)

	req.SetPathValue("id", monitorID.String())

	recorder := httptest.NewRecorder()

	api.getResultsHandler(recorder, req)

	// Check status
	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}
	var results []monitor.Result

	err = json.NewDecoder(recorder.Body).Decode(&results)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Check number of results
	if len(results) != 2 {
		t.Fatalf(
			"expected 2 results, got %d",
			len(results),
		)
	}
	if results[0].ResponseCode != 202 {
		t.Errorf(
			"expected first result to have response code 202, got %d",
			results[0].ResponseCode,
		)
	}

	if results[1].ResponseCode != 201 {
		t.Errorf(
			"expected second result to have response code 201, got %d",
			results[1].ResponseCode,
		)
	}

}

// ------------------------new-mointor route tests-------------------------
func TestAddNewMonitorInvalidJSON(t *testing.T) {
	body := strings.NewReader(`{
		"url": "https://google.com"
	`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/new-monitor",
		body,
	)

	recorder := httptest.NewRecorder()

	api := API{}

	api.addNewMonitor(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}
func TestAddNewMonitorInvalidURL(t *testing.T) {
	body := strings.NewReader(`{
		"url": "not-a-valid-url",
		"interval": "30s",
		"timeout": "5s"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/new-monitor",
		body,
	)

	recorder := httptest.NewRecorder()

	api := API{}

	api.addNewMonitor(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}
func TestAddNewMonitorInvalidInterval(t *testing.T) {
	body := strings.NewReader(`{
		"url": "https://google.com",
		"interval": "not-a-duration",
		"timeout": "5s"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/new-monitor",
		body,
	)

	recorder := httptest.NewRecorder()

	api := API{}

	api.addNewMonitor(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}
func TestAddNewMonitorInvalidTimeout(t *testing.T) {
	body := strings.NewReader(`{
		"url": "https://google.com",
		"interval": "30s",
		"timeout": "not-a-duration"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/new-monitor",
		body,
	)

	recorder := httptest.NewRecorder()

	api := API{}

	api.addNewMonitor(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}
func TestAddNewMonitorDuplicateURL(t *testing.T) {
	db := testDB(t)

	t.Cleanup(func() {
		ctx := context.Background()

		_, err := db.Exec(ctx, "DELETE FROM results")
		if err != nil {
			t.Errorf("failed to clean results: %v", err)
		}

		_, err = db.Exec(ctx, "DELETE FROM monitors")
		if err != nil {
			t.Errorf("failed to clean monitors: %v", err)
		}

		db.Close()
	})

	ctx := context.Background()

	// Create an existing monitor.
	existingMonitor := monitor.Monitor{
		ID:       uuid.New(),
		URL:      "https://google.com",
		Interval: 30 * time.Second,
		Timeout:  5 * time.Second,
	}

	_, err := db.Exec(ctx, `
		INSERT INTO monitors (id, url, interval, timeout)
		VALUES ($1, $2, $3, $4)
	`,
		existingMonitor.ID,
		existingMonitor.URL,
		existingMonitor.Interval,
		existingMonitor.Timeout,
	)
	if err != nil {
		t.Fatalf("failed to insert existing monitor: %v", err)
	}

	// Try to create another monitor with the same URL.
	body := strings.NewReader(`{
		"url": "https://google.com",
		"interval": "30s",
		"timeout": "5s"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/new-monitor",
		body,
	)

	recorder := httptest.NewRecorder()

	api := API{
		db: db,
	}

	api.addNewMonitor(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusConflict,
			recorder.Code,
		)
	}
}

type fakeScheduler struct {
	added []monitor.Monitor
}

func (f *fakeScheduler) AddMonitor(m monitor.Monitor) {
	f.added = append(f.added, m)
}
func TestAddNewMonitorSuccess(t *testing.T) {
	db := testDB(t)

	t.Cleanup(func() {
		ctx := context.Background()

		_, err := db.Exec(ctx, "DELETE FROM results")
		if err != nil {
			t.Errorf("failed to clean results: %v", err)
		}

		_, err = db.Exec(ctx, "DELETE FROM monitors")
		if err != nil {
			t.Errorf("failed to clean monitors: %v", err)
		}

		db.Close()
	})

	body := strings.NewReader(`{
		"url": "https://google.com",
		"interval": "30s",
		"timeout": "5s"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/new-monitor",
		body,
	)

	recorder := httptest.NewRecorder()

	fake := &fakeScheduler{}

	api := API{
		db:        db,
		scheduler: fake,
	}

	api.addNewMonitor(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			recorder.Code,
		)
	}

	var created monitor.Monitor

	err := json.NewDecoder(recorder.Body).Decode(&created)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if created.URL != "https://google.com" {
		t.Errorf("expected URL https://google.com, got %s", created.URL)
	}

	if created.Interval != 30*time.Second {
		t.Errorf(
			"expected interval %s, got %s",
			30*time.Second,
			created.Interval,
		)
	}

	if created.Timeout != 5*time.Second {
		t.Errorf(
			"expected timeout %s, got %s",
			5*time.Second,
			created.Timeout,
		)
	}

	if created.ID == uuid.Nil {
		t.Error("expected monitor ID to be generated")
	}

	if len(fake.added) != 1 {
		t.Fatalf(
			"expected scheduler to receive 1 monitor, got %d",
			len(fake.added),
		)
	}

	if fake.added[0].ID != created.ID {
		t.Errorf("scheduler received different monitor ID")
	}
}
func TestListMonitorsSuccess(t *testing.T) {
	db := testDB(t)

	t.Cleanup(func() {
		ctx := context.Background()

		_, err := db.Exec(ctx, "DELETE FROM results")
		if err != nil {
			t.Errorf("failed to clean results: %v", err)
		}

		_, err = db.Exec(ctx, "DELETE FROM monitors")
		if err != nil {
			t.Errorf("failed to clean monitors: %v", err)
		}

		db.Close()
	})

	ctx := context.Background()

	monitor1 := monitor.Monitor{
		ID:       uuid.New(),
		URL:      "https://google.com",
		Interval: 30 * time.Second,
		Timeout:  5 * time.Second,
	}

	monitor2 := monitor.Monitor{
		ID:       uuid.New(),
		URL:      "https://youtube.com",
		Interval: 60 * time.Second,
		Timeout:  10 * time.Second,
	}

	_, err := db.Exec(ctx, `
		INSERT INTO monitors (id, url, interval, timeout)
		VALUES ($1, $2, $3, $4), ($5, $6, $7, $8)
	`,
		monitor1.ID,
		monitor1.URL,
		monitor1.Interval,
		monitor1.Timeout,
		monitor2.ID,
		monitor2.URL,
		monitor2.Interval,
		monitor2.Timeout,
	)
	if err != nil {
		t.Fatalf("failed to insert monitors: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/monitors",
		nil,
	)

	recorder := httptest.NewRecorder()

	api := API{
		db: db,
	}

	api.listMonitors(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	var monitors []monitor.Monitor

	err = json.NewDecoder(recorder.Body).Decode(&monitors)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(monitors) != 2 {
		t.Fatalf(
			"expected 2 monitors, got %d",
			len(monitors),
		)
	}

	found := map[uuid.UUID]bool{}

	for _, m := range monitors {
		found[m.ID] = true
	}

	if !found[monitor1.ID] {
		t.Errorf("monitor 1 was not returned")
	}

	if !found[monitor2.ID] {
		t.Errorf("monitor 2 was not returned")
	}
}
