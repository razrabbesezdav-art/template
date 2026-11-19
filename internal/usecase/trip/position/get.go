package position

import (
	"context"

	"github.com/google/uuid"
	"github.com/razrabbesezdav-art/template/internal/domain/trip"
)

type PositionRepo interface {
	GetPositions(ctx context.Context, tripID uuid.UUID) ([]*trip.Position, error)
	Get(ctx context.Context, id uuid.UUID) (*trip.Trip, error)
}

type GetUseCase struct {
	repo PositionRepo
}

func NewGet(repo PositionRepo) *GetUseCase {
	return &GetUseCase{repo: repo}
}

func (uc *GetUseCase) GetPositions(ctx context.Context, tripID uuid.UUID) ([]*trip.Position, error) {
	if _, err := uc.repo.Get(ctx, tripID); err != nil {
		return nil, err
	}
	return uc.repo.GetPositions(ctx, tripID)
}