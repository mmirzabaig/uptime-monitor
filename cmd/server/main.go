package main

import (
	"context"
	"fmt"
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

	fmt.Println("Connected to PostgreSQL!")

	s := scheduler.NewScheduler(ctx, db)

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

	server := &http.Server{
		Addr:    ":8080",
		Handler: api.NewRouter(s, db),
	}

	go func() {
		fmt.Println("Server listening on :8080")

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Println("Server error:", err)
		}
	}()

	// Wait for Ctrl+C / SIGTERM
	<-ctx.Done()

	fmt.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Println("Server shutdown error:", err)
	}

	s.Wait()

	fmt.Println("HTTP server stopped")
}
