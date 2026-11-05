package trip

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/razrabbesezdav-art/template/internal/domain/trip"
	"github.com/Masterminds/squirrel"
)

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (*trip.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query, args, err := r.builder.
		Select(
			"id", "user_id", "driver_id",
			"start_latitude", "start_longitude",
			"end_latitude", "end_longitude",
			"price", "status", "started_at", "finished_at",
		).
		From("trips").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get trip query: %w", err)
	}

	row := r.queryer(ctx).QueryRow(ctx, query, args...)
	
	var (
		tID, uID, dID uuid.UUID
		sLat, sLon, eLat, eLon float64
		price int64
		status string
		startedAt time.Time
		finishedAt *time.Time
	)

	err = row.Scan(&tID, &uID, &dID, &sLat, &sLon, &eLat, &eLon, &price, &status, &startedAt, &finishedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, trip.ErrTripNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan trip: %w", err)
	}

	return trip.Restore(tID, uID, dID, sLat, sLon, eLat, eLon, price, trip.Status(status), startedAt, finishedAt), nil
}

func (r *Repo) Update(ctx context.Context, t *trip.Trip) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query, args, err := r.builder.
		Update("trips").
		Set("status", t.Status()).
		Set("finished_at", t.FinishedAt()).
		Set("updated_at", squirrel.Expr("now()")).
		Where(squirrel.Eq{"id": t.ID()}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update trip query: %w", err)
	}

	_, err = r.queryer(ctx).Exec(ctx, query, args...)
	return err
}