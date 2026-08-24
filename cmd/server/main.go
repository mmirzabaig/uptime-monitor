package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/mmirzabaig/uptime-monitor/internal/api"
	"github.com/mmirzabaig/uptime-monitor/internal/monitor"
	"github.com/mmirzabaig/uptime-monitor/internal/scheduler"
	"github.com/mmirzabaig/uptime-monitor/internal/storage"
)

func main() {
	ctx := context.Background()

	db, err := storage.NewPostgres(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Connected to PostgreSQL!")

	s := scheduler.NewScheduler(db)

	websites := monitor.AllMonitors()

	for _, m := range websites {
		err := storage.CreateMonitor(ctx, db, m)
		if err != nil {
			log.Fatal(err)
		}
	}

	monitors, err := storage.GetAllMonitors(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range monitors {
		s.AddMonitor(m)
	}

	fmt.Println("Listening on port 8080!")
	router := api.NewRouter(s, db)

	http.ListenAndServe(":8080", router)

}
