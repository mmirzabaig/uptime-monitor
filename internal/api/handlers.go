package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/mmirzabaig/uptime-monitor/internal/monitor"
	"github.com/mmirzabaig/uptime-monitor/internal/scheduler"
	"github.com/mmirzabaig/uptime-monitor/internal/storage"
)

type API struct {
	scheduler *scheduler.Scheduler
}

func NewRouter(s *scheduler.Scheduler) http.Handler {
	api := API{
		scheduler: s,
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
	urlParam := r.PathValue("id")
	result := storage.GetLastResult(urlParam)
	json.NewEncoder(w).Encode(result)
}

func (a *API) listMonitors(w http.ResponseWriter, r *http.Request) {
	monitors := monitor.AllMonitors()
	json.NewEncoder(w).Encode(monitors)
}

func (a *API) getResultsHandler(w http.ResponseWriter, r *http.Request) {
	urlParam := r.PathValue("id")
	fmt.Println("ID", urlParam)
	result := storage.GetResults(urlParam)
	json.NewEncoder(w).Encode(result)
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
	monitor.AddMonitor(NewMonitor)
	a.scheduler.AddMonitor(NewMonitor)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(NewMonitor)
}
