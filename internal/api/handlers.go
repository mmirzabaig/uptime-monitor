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

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/{id}", healthHandler)
	mux.HandleFunc("GET /results/{id}", getResultsHandler)
	mux.HandleFunc("GET /results/", getAllResultsHandler)
	mux.HandleFunc("POST /new-monitor", addNewMonitor)
	mux.HandleFunc("GET /monitors", listMonitors)
	return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	urlParam := r.PathValue("id")
	result := storage.GetLastResult(urlParam)
	json.NewEncoder(w).Encode(result)
}

func listMonitors(w http.ResponseWriter, r *http.Request) {
	monitors := monitor.AllMonitors()
	json.NewEncoder(w).Encode(monitors)
}

func getResultsHandler(w http.ResponseWriter, r *http.Request) {
	urlParam := r.PathValue("id")
	fmt.Println("ID", urlParam)
	result := storage.GetResults(urlParam)
	json.NewEncoder(w).Encode(result)
}
func getAllResultsHandler(w http.ResponseWriter, r *http.Request) {
	results := storage.GetAllResults()
	json.NewEncoder(w).Encode(results)
}
func addNewMonitor(w http.ResponseWriter, r *http.Request) {
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

	id := uuid.New().String()
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
	scheduler.AddMonitor(NewMonitor)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(NewMonitor)
}
