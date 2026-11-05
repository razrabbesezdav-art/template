package finish

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/razrabbesezdav-art/template/internal/domain/trip"
)

type UseCase struct {
	repo tripRepository
	tx   txManager
}

func New(repo tripRepository, tx txManager) *UseCase {
	return &UseCase{repo: repo, tx: tx}
}

func (uc *UseCase) Finish(ctx context.Context, id uuid.UUID) (*trip.Trip, error) {
	var t *trip.Trip

	err := uc.tx.Do(ctx, func(txCtx context.Context) error {
		found, err := uc.repo.Get(txCtx, id)
		if err != nil {
			return err
		}

		if err := found.Finish(time.Now().UTC()); err != nil {
			return err
		}

		if err := uc.repo.Update(txCtx, found); err != nil {
			return err
		}

		t = found
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("usecase finish trip: %w", err)
	}

	return t, nil
}