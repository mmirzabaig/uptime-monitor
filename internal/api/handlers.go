package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/mmirzabaig/uptime-monitor/internal/monitor"
	"github.com/mmirzabaig/uptime-monitor/internal/scheduler"
	"github.com/mmirzabaig/uptime-monitor/internal/storage"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/{url}", healthHandler)
	mux.HandleFunc("GET /results/{url}", getResultsHandler)
	mux.HandleFunc("GET /results/", getAllResultsHandler)
	mux.HandleFunc("POST /new-monitor", addNewMonitor)
	return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	urlParam := r.PathValue("url")
	result := storage.GetLastResult("https://" + urlParam)
	json.NewEncoder(w).Encode(result)
}

func getResultsHandler(w http.ResponseWriter, r *http.Request) {
	urlParam := r.PathValue("url")
	result := storage.GetResults("https://" + urlParam)
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

	parsedURL, err := url.ParseRequestURI(req.URL)
	if err != nil {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	NewMonitor.URL = parsedURL.String()

	NewMonitor.Interval, err = time.ParseDuration(req.Interval)
	if err != nil {
		http.Error(w, "Ivalid interval", http.StatusBadRequest)
	}

	NewMonitor.Timeout, err = time.ParseDuration(req.Timeout)
	if err != nil {
		http.Error(w, "Ivalid interval", http.StatusBadRequest)
	}

	scheduler.AddMonitor(NewMonitor)
}
