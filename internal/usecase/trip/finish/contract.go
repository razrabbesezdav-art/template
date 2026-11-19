
package finish

import (
	"context"

	"github.com/google/uuid"
	"github.com/razrabbesezdav-art/template/internal/domain/trip"
)

type tripRepository interface {
	Get(ctx context.Context, id uuid.UUID) (*trip.Trip, error)
	Update(ctx context.Context, t *trip.Trip) error
}

type txManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}