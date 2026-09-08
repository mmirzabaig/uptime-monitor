package scheduler

import (
	"context"
	"time"

	schedulerspb "github.com/mmirzabaig/uptime-monitor/gen/scheduler"
	"github.com/mmirzabaig/uptime-monitor/internal/monitor"
	"github.com/mmirzabaig/uptime-monitor/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GRPCServer struct {
	schedulerspb.UnimplementedSchedulerServer
	db *pgxpool.Pool
}

func NewGRPCServer(db *pgxpool.Pool) *GRPCServer {
	return &GRPCServer{
		db: db,
	}
}

func (s *GRPCServer) CreateMonitor(
	ctx context.Context,
	req *schedulerspb.CreateMonitorRequest,
) (*schedulerspb.CreateMonitorResponse, error) {

	id := uuid.New()

	m := monitor.Monitor{
		ID:       id,
		URL:      req.Url,
		Interval: time.Duration(req.Interval),
		Timeout:  time.Duration(req.Timeout),
	}

	created, err := storage.CreateMonitor(ctx, s.db, m)
	if err != nil {
		return nil, err
	}

	if !created {
		return nil, status.Error(
			codes.AlreadyExists,
			"monitor with this URL already exists",
		)
	}

	return &schedulerspb.CreateMonitorResponse{
		Id: id.String(),
	}, nil
}
