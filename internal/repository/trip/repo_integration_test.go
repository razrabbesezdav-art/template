
package trip_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	
	"github.com/razrabbesezdav-art/template/internal/domain/trip"
	repoTrip "github.com/razrabbesezdav-art/template/internal/repository/trip" 
)

func TestRepo_Save_DriverBusyRollback(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	repo := repoTrip.NewRepo(pool, 5*time.Second)
	driverID := uuid.New()

	t1 := trip.New(uuid.New(), uuid.New(), driverID, 0, 0, 0, 0, 100, time.Now())
	err = repo.Save(context.Background(), t1)
	if err != nil {
		t.Fatalf("unexpected error on first save: %v", err)
	}
	
	saved1, err := repo.Get(context.Background(), t1.ID())
	if err != nil || saved1 == nil {
		t.Fatalf("expected first trip to be saved, got error: %v", err)
	}

	t2 := trip.New(uuid.New(), uuid.New(), driverID, 0, 0, 0, 0, 200, time.Now())
	err = repo.Save(context.Background(), t2)
	
	if !errors.Is(err, trip.ErrDriverBusy) {
		t.Fatalf("expected ErrDriverBusy, got: %v", err)
	}
	
	_, err = repo.Get(context.Background(), t2.ID())
	if !errors.Is(err, trip.ErrTripNotFound) {
		t.Fatalf("expected second trip to be rolled back (not found), got: %v", err)
	}
}