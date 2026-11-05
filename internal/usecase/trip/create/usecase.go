package create

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/razrabbesezdav-art/template/internal/domain/trip"
)

// Request — параметры для создания поездки.
type Request struct {
	UserID   uuid.UUID
	DriverID uuid.UUID
	StartLat float64
	StartLon float64
	EndLat   float64
	EndLon   float64
	Price    int64
}

type UseCase struct {
	repo tripRepository
	tx   txManager
}

func New(repo tripRepository, tx txManager) *UseCase {
	return &UseCase{
		repo: repo,
		tx:   tx,
	}
}


func (uc *UseCase) Create(ctx context.Context, req Request) (*trip.Trip, error) {
	t := trip.New(
		uuid.New(),
		req.UserID,
		req.DriverID,
		req.StartLat,
		req.StartLon,
		req.EndLat,
		req.EndLon,
		req.Price,
		time.Now().UTC(),
	)


	err := uc.tx.Do(ctx, func(txCtx context.Context) error {
		if err := uc.repo.Save(txCtx, t); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("usecase create trip: %w", err)
	}

	return t, nil
}