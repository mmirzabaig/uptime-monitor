package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	schedulerspb "github.com/mmirzabaig/uptime-monitor/gen/scheduler"
	"github.com/mmirzabaig/uptime-monitor/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type API struct {
	db        *pgxpool.Pool
	scheduler schedulerspb.SchedulerClient
}

func NewRouter(db *pgxpool.Pool, scheduler schedulerspb.SchedulerClient) http.Handler {
	api := API{
		db:        db,
		scheduler: scheduler,
	}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/{id}", api.healthHandler)
	mux.HandleFunc("GET /results/{id}", api.getResultsHandler)
	// mux.HandleFunc("GET /results/", api.getAllResultsHandler)
	mux.HandleFunc("POST /new-monitor", api.addNewMonitor)
	mux.HandleFunc("GET /monitors", api.listMonitors)
	return mux
}

func (a *API) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.PathValue("id")

	monitorID, err := uuid.Parse(id)
	if err != nil {
		writeJSONError(w, "Invalid monitor ID", http.StatusBadRequest)
		return
	}

	result, err := storage.GetLastResult(
		r.Context(),
		a.db,
		monitorID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		writeJSONError(w, "No health result found for monitor", http.StatusNotFound)
		return
	}

	if err != nil {
		writeJSONError(w, "Failed to get health result", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func (a *API) listMonitors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	monitors, err := storage.GetAllMonitors(r.Context(), a.db)
	if err != nil {
		writeJSONError(w, "Failed to get monitors", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(monitors)
}

func (a *API) getResultsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	urlParam := r.PathValue("id")
	monitorID, err := uuid.Parse(urlParam)
	if err != nil {
		writeJSONError(w, "Invalid monitor ID", http.StatusBadRequest)
		return
	}
	query := r.URL.Query()

	limit := 50
	limitStr := query.Get("limit")

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			writeJSONError(w, "limit must be a number", http.StatusBadRequest)
			return
		}

		if parsedLimit <= 0 {
			writeJSONError(w, "limit must be greater than 0", http.StatusBadRequest)
			return
		}

		if parsedLimit > 100 {
			writeJSONError(w, "limit cannot be greater than 100", http.StatusBadRequest)
			return
		}

		limit = parsedLimit
	}

	offset := 0

	offsetStr := query.Get("offset")

	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil {
			writeJSONError(w, "offset must be a number", http.StatusBadRequest)
			return
		}

		if parsedOffset < 0 {
			writeJSONError(w, "offset must be at least 0", http.StatusBadRequest)
			return
		}

		offset = parsedOffset
	}

	exists, err := storage.MonitorExists(r.Context(), a.db, monitorID)
	if err != nil {
		writeJSONError(w, "Failed to check monitor", http.StatusInternalServerError)
		return
	}

	if !exists {
		writeJSONError(w, "Monitor not found", http.StatusNotFound)
		return
	}

	results, err := storage.GetResults(
		r.Context(),
		a.db,
		monitorID,
		limit,
		offset,
	)
	if err != nil {
		writeJSONError(w, "Failed to get results", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(results)
}

//	func (a *API) getAllResultsHandler(w http.ResponseWriter, r *http.Request) {
//		results := storage.GetAllResults()
//		json.NewEncoder(w).Encode(results)
//	}
func (a *API) addNewMonitor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()

	type Req struct {
		URL      string `json:"url"`
		Interval string `json:"interval"`
		Timeout  string `json:"timeout"`
	}
	var req Req
	decoder := json.NewDecoder(r.Body)
	// This line below will reject all json requests that dont have the exact json fields
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&req)

	if err != nil {
		writeJSONError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	parsedURL, err := url.ParseRequestURI(req.URL)
	if err != nil {
		writeJSONError(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		writeJSONError(w, "URL must use http or https", http.StatusBadRequest)
		return
	}
	if parsedURL.Host == "" {
		writeJSONError(w, "URL must include a host", http.StatusBadRequest)
		return
	}
	normalizedURL := parsedURL.String()

	interval, err := time.ParseDuration(req.Interval)
	if err != nil {
		writeJSONError(w, "Invalid interval", http.StatusBadRequest)
		return
	}
	if interval <= 0 {
		writeJSONError(w, "interval must be greater than 0", http.StatusBadRequest)
		return
	}

	timeout, err := time.ParseDuration(req.Timeout)
	if err != nil {
		writeJSONError(w, "Invalid timeout", http.StatusBadRequest)
		return
	}
	if timeout <= 0 {
		writeJSONError(w, "timeout must be greater than 0", http.StatusBadRequest)
		return
	}

	if timeout >= interval {
		writeJSONError(w, "timeout must be less than interval", http.StatusBadRequest)
		return
	}

	response, err := a.scheduler.CreateMonitor(
		r.Context(),
		&schedulerspb.CreateMonitorRequest{
			Url:      normalizedURL,
			Interval: int64(interval),
			Timeout:  int64(timeout),
		},
	)
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			writeJSONError(
				w,
				"A monitor with this URL already exists",
				http.StatusConflict,
			)
			return
		}

		writeJSONError(
			w,
			"Failed to create monitor",
			http.StatusInternalServerError,
		)
		return
	}

	// a.scheduler.AddMonitor(newMonitor)

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]string{
		"id":  response.Id,
		"url": normalizedURL,
	})
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
