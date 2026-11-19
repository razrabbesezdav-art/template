package positions

import (
	"context"
	"log/slog"
	"time"

	"github.com/razrabbesezdav-art/template/internal/repository/trip"
)

type Scanner struct {
	repo     *trip.Repo
	pool     *Pool
	interval time.Duration
	staleAfter time.Duration
	logger   *slog.Logger
}

func NewScanner(repo *trip.Repo, pool *Pool, interval, staleAfter time.Duration, logger *slog.Logger) *Scanner {
	return &Scanner{
		repo:       repo,
		pool:       pool,
		interval:   interval,
		staleAfter: staleAfter,
		logger:     logger,
	}
}

func (s *Scanner) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("scanner stopped")
			return
		case <-ticker.C:
			s.scan(ctx)
		}
	}
}

func (s *Scanner) scan(ctx context.Context) {
	trips, err := s.repo.GetStaleTrips(ctx, s.staleAfter)
	if err != nil {
		s.logger.Error("failed to get stale trips", slog.String("error", err.Error()))
		return
	}

	for _, t := range trips {
		queued, dup, dropped := s.pool.TryEnqueue(ctx, Job{TripID: t.TripID, DriverID: t.DriverID})
		
		if queued {
		} else if dup {
		} else if dropped {
			s.logger.Warn("position queue is full, dropping trip", slog.String("trip_id", t.TripID.String()))
			break 
		}
	}
}