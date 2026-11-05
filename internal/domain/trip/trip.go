package trip

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
)

type Trip struct {
	id         uuid.UUID
	userID     uuid.UUID
	driverID   uuid.UUID
	startLat   float64
	startLon   float64
	endLat     float64
	endLon     float64
	price      int64
	status     Status
	startedAt  time.Time
	finishedAt *time.Time
}

func New(
	id, userID, driverID uuid.UUID,
	startLat, startLon, endLat, endLon float64,
	price int64,
	startedAt time.Time,
) *Trip {
	return &Trip{
		id:        id,
		userID:    userID,
		driverID:  driverID,
		startLat:  startLat,
		startLon:  startLon,
		endLat:    endLat,
		endLon:    endLon,
		price:     price,
		status:    StatusActive,
		startedAt: startedAt,
	}
}

func Restore(
	id, userID, driverID uuid.UUID,
	startLat, startLon, endLat, endLon float64,
	price int64,
	status Status,
	startedAt time.Time,
	finishedAt *time.Time,
) *Trip {
	return &Trip{
		id:         id,
		userID:     userID,
		driverID:   driverID,
		startLat:   startLat,
		startLon:   startLon,
		endLat:     endLat,
		endLon:     endLon,
		price:      price,
		status:     status,
		startedAt:  startedAt,
		finishedAt: finishedAt,
	}
}

func (t *Trip) Finish(at time.Time) error {
	if t.status != StatusActive {
		return ErrTripCompleted
	}
	t.status = StatusCompleted
	t.finishedAt = &at
	return nil
}

func (t *Trip) ID() uuid.UUID          { return t.id }
func (t *Trip) UserID() uuid.UUID      { return t.userID }
func (t *Trip) DriverID() uuid.UUID    { return t.driverID }
func (t *Trip) StartLat() float64      { return t.startLat }
func (t *Trip) StartLon() float64      { return t.startLon }
func (t *Trip) EndLat() float64        { return t.endLat }
func (t *Trip) EndLon() float64        { return t.endLon }
func (t *Trip) Price() int64           { return t.price }
func (t *Trip) Status() Status         { return t.status }
func (t *Trip) StartedAt() time.Time   { return t.startedAt }
func (t *Trip) FinishedAt() *time.Time { return t.finishedAt }