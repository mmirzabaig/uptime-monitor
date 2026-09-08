package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	schedulerspb "github.com/mmirzabaig/uptime-monitor/gen/scheduler"
	"github.com/mmirzabaig/uptime-monitor/internal/scheduler"
	"github.com/mmirzabaig/uptime-monitor/internal/storage"
	"google.golang.org/grpc"
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

	grpcServer := grpc.NewServer()

	schedulerGRPC := scheduler.NewGRPCServer(db)

	schedulerspb.RegisterSchedulerServer(grpcServer, schedulerGRPC)

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		log.Println("Scheduler gRPC server listening on :50051")

		if err := grpcServer.Serve(listener); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

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

	<-ctx.Done()

	log.Printf("Shutting down...")

	grpcServer.GracefulStop()
	listener.Close()

	s.Wait()

	log.Printf("Scheduler stopped")
}
