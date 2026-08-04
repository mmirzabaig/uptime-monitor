package main

import (
	"fmt"
	"net/http"

	"github.com/mmirzabaig/uptime-monitor/internal/api"
	"github.com/mmirzabaig/uptime-monitor/internal/monitor"
	"github.com/mmirzabaig/uptime-monitor/internal/scheduler"
)

func main() {

	websites := monitor.AllMonitors()

	scheduler.Run(websites)

	fmt.Println("Listening on port 8080!")
	router := api.NewRouter()

	http.ListenAndServe(":8080", router)

}
