package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mmirzabaig/uptime-monitor/internal/api"
	"github.com/mmirzabaig/uptime-monitor/internal/monitor"
	"github.com/mmirzabaig/uptime-monitor/internal/scheduler"
	"github.com/mmirzabaig/uptime-monitor/internal/storage"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	db, err := storage.NewPostgres(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Printf("Connected to PostgreSQL!")

	s := scheduler.NewScheduler(ctx, db)

	websites := monitor.AllMonitors()

	for _, m := range websites {
		created, err := storage.CreateMonitor(ctx, db, m)
		if err != nil {
			log.Fatal(err)
		}

		if created {
			log.Printf("Created monitor: %s", m.URL)
		} else {
			log.Printf("Monitor already exists: %s", m.URL)
		}
	}

	log.Println("Loading monitors from database...")

	monitors, err := storage.GetAllMonitors(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Loaded %d monitors", len(monitors))

	for _, m := range monitors {
		log.Printf(
			"Starting monitor: %s (ID: %s, interval: %s, timeout: %s)",
			m.URL,
			m.ID,
			m.Interval,
			m.Timeout,
		)
		s.AddMonitor(m)
	}

	serverAddr := os.Getenv("SERVER_ADDR")

	if serverAddr == "" {
		serverAddr = ":8080"
	}

	server := &http.Server{
		Addr:    serverAddr,
		Handler: api.NewRouter(db),
	}

	go func() {
		log.Printf("Server listening on %s", serverAddr)

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Printf("Server error:", err)
		}
	}()

	// Wait for Ctrl+C / SIGTERM
	<-ctx.Done()

	log.Printf("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error:", err)
	}

	s.Wait()

	log.Printf("HTTP server stopped")
}
