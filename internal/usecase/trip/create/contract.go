
package create

import (
	"context"

	"github.com/razrabbesezdav-art/template/internal/domain/trip"
)

type tripRepository interface {
	Save(ctx context.Context, t *trip.Trip) error
}

type txManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}