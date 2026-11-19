package trip

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/razrabbesezdav-art/template/internal/domain/trip"
)

func (r *Repo) Save(ctx context.Context, t *trip.Trip) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query, args, err := r.builder.
		Insert("trips").
		Columns(
			"id", "user_id", "driver_id",
			"start_latitude", "start_longitude",
			"end_latitude", "end_longitude",
			"price", "status", "started_at",
		).
		Values(
			t.ID(), t.UserID(), t.DriverID(),
			t.StartLat(), t.StartLon(),
			t.EndLat(), t.EndLon(),
			t.Price(), t.Status(), t.StartedAt(),
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("build save trip query: %w", err)
	}

	_, err = r.queryer(ctx).Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "trips_one_active_per_driver_idx" {
			return trip.ErrDriverBusy
		}
		return fmt.Errorf("save trip: %w", err)
	}

	return nil
}