package trip

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidLatitude  = errors.New("latitude must be between -90 and 90")
	ErrInvalidLongitude = errors.New("longitude must be between -180 and 180")
	ErrRecordedInFuture = errors.New("recorded_at is in the future")
)

const maxFutureSkew = time.Minute

type Position struct {
	id         int64
	tripID     uuid.UUID
	latitude   float64
	longitude  float64
	recordedAt time.Time
}

func NewPosition(
	tripID uuid.UUID,
	latitude, longitude float64,
	recordedAt time.Time,
	now time.Time,
) (*Position, error) {
	if latitude < -90 || latitude > 90 {
		return nil, ErrInvalidLatitude
	}
	if longitude < -180 || longitude > 180 {
		return nil, ErrInvalidLongitude
	}
	if recordedAt.After(now.Add(maxFutureSkew)) {
		return nil, ErrRecordedInFuture
	}

	return &Position{
		tripID:     tripID,
		latitude:   latitude,
		longitude:  longitude,
		recordedAt: recordedAt,
	}, nil
}

func RestorePosition(
	id int64,
	tripID uuid.UUID,
	latitude, longitude float64,
	recordedAt time.Time,
) *Position {
	return &Position{
		id:         id,
		tripID:     tripID,
		latitude:   latitude,
		longitude:  longitude,
		recordedAt: recordedAt,
	}
}

func (p *Position) ID() int64             { return p.id }
func (p *Position) TripID() uuid.UUID     { return p.tripID }
func (p *Position) Latitude() float64     { return p.latitude }
func (p *Position) Longitude() float64    { return p.longitude }
func (p *Position) RecordedAt() time.Time { return p.recordedAt }