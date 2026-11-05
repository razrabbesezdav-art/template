package get

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/razrabbesezdav-art/template/internal/domain/trip"
)

type UseCase struct {
	repo tripRepository
}

func New(repo tripRepository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Get(ctx context.Context, id uuid.UUID) (*trip.Trip, error) {
	t, err := uc.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("usecase get trip: %w", err)
	}
	return t, nil
}