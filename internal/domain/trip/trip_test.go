package trip

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTrip_Finish(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		status     Status
		finishedAt *time.Time
		wantErr    error
		wantStatus Status
	}{
		{
			name:       "success finish active trip",
			status:     StatusActive,
			finishedAt: nil,
			wantErr:    nil,
			wantStatus: StatusCompleted,
		},
		{
			name:       "error finish already completed trip",
			status:     StatusCompleted,
			finishedAt: &now, 
			wantErr:    ErrTripCompleted,
			wantStatus: StatusCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trip := Restore(
				uuid.New(), uuid.New(), uuid.New(),
				0, 0, 0, 0, 100, tt.status, now, tt.finishedAt,
			)

			err := trip.Finish(time.Now())

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Finish() error = %v, wantErr %v", err, tt.wantErr)
			}
			if trip.Status() != tt.wantStatus {
				t.Errorf("Finish() status = %v, want %v", trip.Status(), tt.wantStatus)
			}
			if tt.wantStatus == StatusCompleted && trip.FinishedAt() == nil {
				t.Error("Finish() finishedAt is nil")
			}
		})
	}
}