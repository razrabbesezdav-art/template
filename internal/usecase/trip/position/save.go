package position

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/razrabbesezdav-art/template/internal/domain/trip"
	"github.com/razrabbesezdav-art/template/internal/infra/postgres/transaction"
)

type TripRepo interface {
	Get(ctx context.Context, id uuid.UUID) (*trip.Trip, error)
	SavePosition(ctx context.Context, tx transaction.Tx, p *trip.Position) (*trip.Position, error)
	UpdateLastPosition(ctx context.Context, tx transaction.Tx, tripID uuid.UUID, recordedAt time.Time) error
}

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type Canceller interface {
	Cancel(tripID uuid.UUID)
}

type Request struct {
	TripID     uuid.UUID
	Latitude   float64
	Longitude  float64
	RecordedAt time.Time
}

type UseCase struct {
	repo      TripRepo
	txManager TxManager
	canceller Canceller
	now       func() time.Time
}

func New(repo TripRepo, txManager TxManager, canceller Canceller, now func() time.Time) *UseCase {
	if now == nil {
		now = time.Now
	}
	return &UseCase{
		repo:      repo,
		txManager: txManager,
		canceller: canceller,
		now:       now,
	}
}

func (uc *UseCase) Save(ctx context.Context, req Request) (*trip.Position, error) {
	p, err := trip.NewPosition(req.TripID, req.Latitude, req.Longitude, req.RecordedAt, uc.now())
	if err != nil {
		return nil, err
	}

	t, err := uc.repo.Get(ctx, req.TripID)
	if err != nil {
		return nil, err
	}
	if t.Status() != trip.StatusActive {
		return nil, trip.ErrTripCompleted
	}

	var saved *trip.Position
	err = uc.txManager.Do(ctx, func(ctx context.Context) error {
		tx, ok := transaction.TxFromContext(ctx)
		if !ok {
			return errors.New("no transaction in context")
		}

		sp, err := uc.repo.SavePosition(ctx, tx, p)
		if err != nil {
			return fmt.Errorf("save position: %w", err)
		}
		saved = sp

		if err := uc.repo.UpdateLastPosition(ctx, tx, req.TripID, req.RecordedAt); err != nil {
			return fmt.Errorf("update last position: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if uc.canceller != nil {
		uc.canceller.Cancel(req.TripID)
	}

	return saved, nil
}