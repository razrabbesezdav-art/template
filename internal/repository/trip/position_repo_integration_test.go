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
	"github.com/razrabbesezdav-art/template/internal/infra/postgres/transaction"
	repoTrip "github.com/razrabbesezdav-art/template/internal/repository/trip"
)

func withTx(t *testing.T, pool *pgxpool.Pool, fn func(ctx context.Context, tx transaction.Tx)) {
	t.Helper()
	tx, err := pool.Begin(context.Background())
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
	})
	ctx := transaction.WithTxForTest(context.Background(), tx)
	fn(ctx, tx)
}

func TestRepo_SavePosition_AtomicAndMonotonic(t *testing.T) {
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
	txMgr := transaction.NewManager(pool)

	driverID := uuid.New()
	tr := trip.New(uuid.New(), uuid.New(), driverID, 0, 0, 0, 0, 100, time.Now().Add(-time.Hour))

	withTx(t, pool, func(ctx context.Context, tx transaction.Tx) {
		if err := repo.Save(ctx, tr); err != nil {
			t.Fatalf("save trip: %v", err)
		}

		freshAt := time.Now().Add(-10 * time.Minute)
		p1, err := trip.NewPosition(tr.ID(), 59.93, 30.31, freshAt, time.Now())
		if err != nil {
			t.Fatalf("new position: %v", err)
		}
		if _, err := repo.SavePosition(ctx, tx, p1); err != nil {
			t.Fatalf("save position: %v", err)
		}
		if err := repo.UpdateLastPosition(ctx, tx, tr.ID(), freshAt); err != nil {
			t.Fatalf("update last position: %v", err)
		}

		got, err := repo.Get(ctx, tr.ID())
		if err != nil {
			t.Fatalf("get trip: %v", err)
		}
		if got.LastPositionAt() == nil || !got.LastPositionAt().Equal(freshAt) {
			t.Fatalf("last_position_at = %v, want %v", got.LastPositionAt(), freshAt)
		}

		oldAt := freshAt.Add(-time.Hour)
		p2, err := trip.NewPosition(tr.ID(), 59.94, 30.32, oldAt, time.Now())
		if err != nil {
			t.Fatalf("new old position: %v", err)
		}
		if _, err := repo.SavePosition(ctx, tx, p2); err != nil {
			t.Fatalf("save old position: %v", err)
		}
		if err := repo.UpdateLastPosition(ctx, tx, tr.ID(), oldAt); err != nil {
			t.Fatalf("update last position with old: %v", err)
		}

		got2, err := repo.Get(ctx, tr.ID())
		if err != nil {
			t.Fatalf("get trip after old: %v", err)
		}
		if !got2.LastPositionAt().Equal(freshAt) {
			t.Fatalf("last_position_at = %v, want %v (old coordinate must not regress)", got2.LastPositionAt(), freshAt)
		}
	})

	_ = txMgr
}

func TestRepo_SavePosition_RollbackOnError(t *testing.T) {
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
	txMgr := transaction.NewManager(pool)

	tr := trip.New(uuid.New(), uuid.New(), uuid.New(), 0, 0, 0, 0, 100, time.Now().Add(-time.Hour))
	if err := repo.Save(context.Background(), tr); err != nil {
		t.Fatalf("save trip: %v", err)
	}

	p, err := trip.NewPosition(tr.ID(), 59.93, 30.31, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("new position: %v", err)
	}

	sentinel := errors.New("boom")
	err = txMgr.Do(context.Background(), func(ctx context.Context) error {
		tx, ok := transaction.TxFromContext(ctx)
		if !ok {
			t.Fatal("no tx in context")
		}
		if _, err := repo.SavePosition(ctx, tx, p); err != nil {
			return err
		}
		if err := repo.UpdateLastPosition(ctx, tx, tr.ID(), p.RecordedAt()); err != nil {
			return err
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel, got %v", err)
	}

	positions, err := repo.GetPositions(context.Background(), tr.ID())
	if err != nil {
		t.Fatalf("get positions: %v", err)
	}
	if len(positions) != 0 {
		t.Fatalf("positions saved despite rollback: %d", len(positions))
	}

	got, err := repo.Get(context.Background(), tr.ID())
	if err != nil {
		t.Fatalf("get trip: %v", err)
	}
	if got.LastPositionAt() != nil {
		t.Fatalf("last_position_at not rolled back: %v", got.LastPositionAt())
	}
}