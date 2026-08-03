package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mmirzabaig/uptime-monitor/internal/api"
	"github.com/mmirzabaig/uptime-monitor/internal/monitor"
	"github.com/mmirzabaig/uptime-monitor/internal/scheduler"
)

func main() {

	websites := []monitor.Monitor{
		{
			URL:      "https://google.com",
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
		},
		{
			URL:      "https://youtube.com",
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
		},
		{
			URL:      "https://go.dev/tour/list",
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
		},
		{
			URL:      "https://amazon.com",
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
		},
	}
	scheduler.Run(websites)
	fmt.Println("Listening on port 8080!")
	router := api.NewRouter()

	http.ListenAndServe(":8080", router)

}
