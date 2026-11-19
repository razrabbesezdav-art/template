package trip

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/razrabbesezdav-art/template/internal/domain/trip"
	"github.com/razrabbesezdav-art/template/internal/infra/postgres/transaction"
)

func (r *Repo) SavePosition(ctx context.Context, tx transaction.Tx, p *trip.Position) (*trip.Position, error) {
	const query = `
		INSERT INTO trip_positions (trip_id, latitude, longitude, recorded_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64
	err := tx.QueryRow(ctx, query,
		p.TripID(), p.Latitude(), p.Longitude(), p.RecordedAt(),
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("insert position: %w", err)
	}

	return trip.RestorePosition(
		id, p.TripID(), p.Latitude(), p.Longitude(), p.RecordedAt(),
	), nil
}

func (r *Repo) UpdateLastPosition(ctx context.Context, tx transaction.Tx, tripID uuid.UUID, recordedAt time.Time) error {
	const query = `
		UPDATE trips
		SET last_position_at = GREATEST(COALESCE(last_position_at, 'epoch'::timestamptz), $2)
		WHERE id = $1
	`

	_, err := tx.Exec(ctx, query, tripID, recordedAt)
	if err != nil {
		return fmt.Errorf("update last_position_at: %w", err)
	}
	return nil
}

func (r *Repo) GetPositions(ctx context.Context, tripID uuid.UUID) ([]*trip.Position, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	const query = `
		SELECT id, trip_id, latitude, longitude, recorded_at
		FROM trip_positions
		WHERE trip_id = $1
		ORDER BY recorded_at ASC, id ASC
	`

	rows, err := r.queryer(ctx).Query(ctx, query, tripID)
	if err != nil {
		return nil, fmt.Errorf("query positions: %w", err)
	}
	defer rows.Close()

	var result []*trip.Position
	for rows.Next() {
		var (
			id         int64
			tid        uuid.UUID
			lat, lon   float64
			recordedAt time.Time
		)
		if err := rows.Scan(&id, &tid, &lat, &lon, &recordedAt); err != nil {
			return nil, fmt.Errorf("scan position: %w", err)
		}
		result = append(result, trip.RestorePosition(id, tid, lat, lon, recordedAt))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate positions: %w", err)
	}
	return result, nil
}

type StaleTrip struct {
	TripID         uuid.UUID
	DriverID       uuid.UUID
	LastActivityAt time.Time
}

func (r *Repo) GetStaleTrips(ctx context.Context, staleAfter time.Duration) ([]StaleTrip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	threshold := time.Now().Add(-staleAfter)

	const query = `
		SELECT id, driver_id, COALESCE(last_position_at, started_at) AS last_activity
		FROM trips
		WHERE status = 'active'
		  AND COALESCE(last_position_at, started_at) < $1
		ORDER BY last_activity ASC
	`

	rows, err := r.queryer(ctx).Query(ctx, query, threshold)
	if err != nil {
		return nil, fmt.Errorf("query stale trips: %w", err)
	}
	defer rows.Close()

	var result []StaleTrip
	for rows.Next() {
		var s StaleTrip
		if err := rows.Scan(&s.TripID, &s.DriverID, &s.LastActivityAt); err != nil {
			return nil, fmt.Errorf("scan stale trip: %w", err)
		}
		result = append(result, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stale trips: %w", err)
	}
	return result, nil
}