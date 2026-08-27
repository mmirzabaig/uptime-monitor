package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmirzabaig/uptime-monitor/internal/monitor"
	"github.com/mmirzabaig/uptime-monitor/internal/scheduler"
	"github.com/mmirzabaig/uptime-monitor/internal/storage"
)

type API struct {
	scheduler *scheduler.Scheduler
	db        *pgxpool.Pool
}

func NewRouter(s *scheduler.Scheduler, db *pgxpool.Pool) http.Handler {
	api := API{
		scheduler: s,
		db:        db,
	}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/{id}", api.healthHandler)
	mux.HandleFunc("GET /results/{id}", api.getResultsHandler)
	mux.HandleFunc("GET /results/", api.getAllResultsHandler)
	mux.HandleFunc("POST /new-monitor", api.addNewMonitor)
	mux.HandleFunc("GET /monitors", api.listMonitors)
	return mux
}

func (a *API) healthHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	monitorID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid monitor ID", http.StatusBadRequest)
		return
	}

	result, err := storage.GetLastResult(
		r.Context(),
		a.db,
		monitorID,
	)
	if err != nil {
		http.Error(w, "Failed to get health result", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (a *API) listMonitors(w http.ResponseWriter, r *http.Request) {
	monitors, err := storage.GetAllMonitors(r.Context(), a.db)
	if err != nil {
		http.Error(w, "Failed to get monitors", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(monitors)
}

func (a *API) getResultsHandler(w http.ResponseWriter, r *http.Request) {
	urlParam := r.PathValue("id")
	monitorID, err := uuid.Parse(urlParam)
	if err != nil {
		http.Error(w, "Invalid monitor ID", http.StatusBadRequest)
		return
	}
	query := r.URL.Query()

	limit := 50
	limitStr := query.Get("limit")

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "limit must be a number", http.StatusBadRequest)
			return
		}

		if parsedLimit <= 0 {
			http.Error(w, "limit must be greater than 0", http.StatusBadRequest)
			return
		}

		if parsedLimit > 100 {
			http.Error(w, "limit cannot be greater than 100", http.StatusBadRequest)
			return
		}

		limit = parsedLimit
	}

	offset := 0

	offsetStr := query.Get("offset")

	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil {
			http.Error(w, "offset must be a number", http.StatusBadRequest)
			return
		}

		if parsedOffset < 0 {
			http.Error(w, "offset must be atleast 0", http.StatusBadRequest)
			return
		}

		offset = parsedOffset
	}

	results, err := storage.GetResults(r.Context(), a.db, monitorID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get results", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(results)
}
func (a *API) getAllResultsHandler(w http.ResponseWriter, r *http.Request) {
	results := storage.GetAllResults()
	json.NewEncoder(w).Encode(results)
}
func (a *API) addNewMonitor(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type Req struct {
		URL      string `json:"url"`
		Interval string `json:"interval"`
		Timeout  string `json:"timeout"`
	}
	var req Req
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	var NewMonitor monitor.Monitor

	id := uuid.New()
	NewMonitor.ID = id

	parsedURL, err := url.ParseRequestURI(req.URL)
	if err != nil {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	NewMonitor.URL = parsedURL.String()

	NewMonitor.Interval, err = time.ParseDuration(req.Interval)
	if err != nil {
		http.Error(w, "Invalid interval", http.StatusBadRequest)
		return
	}

	NewMonitor.Timeout, err = time.ParseDuration(req.Timeout)
	if err != nil {
		http.Error(w, "Invalid interval", http.StatusBadRequest)
		return
	}

	err = storage.CreateMonitor(r.Context(), a.db, NewMonitor)
	if err != nil {
		http.Error(w, "Failed to create monitor", http.StatusInternalServerError)
		return
	}

	a.scheduler.AddMonitor(NewMonitor)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(NewMonitor)
}
