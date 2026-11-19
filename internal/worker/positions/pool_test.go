package positions

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/goleak"
)

type mockGateway struct {
	mu    sync.Mutex
	calls map[uuid.UUID]int
	delay time.Duration
}

func newMockGateway() *mockGateway {
	return &mockGateway{calls: make(map[uuid.UUID]int)}
}

func (m *mockGateway) RequestPosition(ctx context.Context, driverID, tripID uuid.UUID) error {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	m.mu.Lock()
	m.calls[tripID]++
	m.mu.Unlock()
	return nil
}

func (m *mockGateway) getCalls(tripID uuid.UUID) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls[tripID]
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestPool_DeduplicationAndCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gateway := newMockGateway()
	gateway.delay = 50 * time.Millisecond 

	pool := NewPool(1, 10, gateway, 100*time.Millisecond)
	pool.Start(ctx, 1)
	
	tripID := uuid.New()
	driverID := uuid.New()
	job := Job{TripID: tripID, DriverID: driverID}

	queued, dup, dropped := pool.TryEnqueue(ctx, job)
	if !queued || dup || dropped {
		t.Fatalf("expected queued, got q=%v d=%v dr=%v", queued, dup, dropped)
	}
	queued, dup, dropped = pool.TryEnqueue(ctx, job)
	if queued || !dup || dropped {
		t.Fatalf("expected duplicate, got q=%v d=%v dr=%v", queued, dup, dropped)
	}

	pool.Cancel(tripID)
	time.Sleep(100 * time.Millisecond)

	if calls := gateway.getCalls(tripID); calls != 0 {
		t.Fatalf("expected 0 calls, got %d", calls)
	}

	pool.Stop()
}

func TestPool_Overflow(t *testing.T) {
	ctx := context.Background()
	gateway := newMockGateway()
	gateway.delay = 100 * time.Millisecond
	
	pool := NewPool(1, 1, gateway, 100*time.Millisecond)
	pool.Start(ctx, 1)

	job1 := Job{TripID: uuid.New(), DriverID: uuid.New()}
	job2 := Job{TripID: uuid.New(), DriverID: uuid.New()}

	pool.TryEnqueue(ctx, job1)
	pool.TryEnqueue(ctx, job2)
	
	job3 := Job{TripID: uuid.New(), DriverID: uuid.New()}
	queued, _, dropped := pool.TryEnqueue(ctx, job3)
	if !dropped {
		t.Fatalf("expected dropped due to overflow, got queued=%v", queued)
	}

	pool.Stop()
}

func TestPool_Stop_And_NoGoroutines(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gateway := newMockGateway()
	pool := NewPool(3, 10, gateway, 100*time.Millisecond)
	pool.Start(ctx, 3)

	for i := 0; i < 5; i++ {
		pool.TryEnqueue(ctx, Job{TripID: uuid.New(), DriverID: uuid.New()})
	}

	pool.Stop()

}