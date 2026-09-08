package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	schedulerspb "github.com/mmirzabaig/uptime-monitor/gen/scheduler"
	"github.com/mmirzabaig/uptime-monitor/internal/api"
	"github.com/mmirzabaig/uptime-monitor/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	// ---------- GRPC Client Code -----------
	schedulerAddr := os.Getenv("SCHEDULER_ADDR")

	if schedulerAddr == "" {
		schedulerAddr = "localhost:50051"
	}

	conn, err := grpc.NewClient(
		schedulerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	schedulerClient := schedulerspb.NewSchedulerClient(conn)

	// ------------------------------------------------------
	serverAddr := os.Getenv("SERVER_ADDR")

	if serverAddr == "" {
		serverAddr = ":8080"
	}

	server := &http.Server{
		Addr:    serverAddr,
		Handler: api.NewRouter(db, schedulerClient),
	}

	go func() {
		log.Printf("Server listening on %s", serverAddr)

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	<-ctx.Done()

	log.Printf("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Printf("HTTP server stopped")
}
