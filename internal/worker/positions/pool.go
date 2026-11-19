package positions

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type PushGateway interface {
	RequestPosition(ctx context.Context, driverID, tripID uuid.UUID) error
}

type Job struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

type Pool struct {
	queue chan Job
	wg    sync.WaitGroup

	mu     sync.Mutex
	active map[uuid.UUID]struct{}

	gateway PushGateway
	timeout time.Duration
}

func NewPool(workers, queueSize int, gateway PushGateway, pushTimeout time.Duration) *Pool {
	return &Pool{
		queue:   make(chan Job, queueSize),
		active:  make(map[uuid.UUID]struct{}),
		gateway: gateway,
		timeout: pushTimeout,
	}
}

func (p *Pool) TryEnqueue(ctx context.Context, job Job) (queued, duplicate, dropped bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.active[job.TripID]; ok {
		recordStaleCheck(ctx, "duplicate")
		return false, true, false
	}

	p.active[job.TripID] = struct{}{}

	select {
	case p.queue <- job:
		queueSize.Add(ctx, 1)
		recordStaleCheck(ctx, "queued")
		return true, false, false
	default:
		delete(p.active, job.TripID)
		recordStaleCheck(ctx, "dropped")
		return false, false, true
	}
}

func (p *Pool) Cancel(tripID uuid.UUID) {
	p.mu.Lock()
	delete(p.active, tripID)
	p.mu.Unlock()
}

func (p *Pool) Start(ctx context.Context, numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		p.wg.Add(1)
		go p.worker(ctx)
	}
}

func (p *Pool) worker(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-p.queue:
			if !ok {
				return
			}
			p.process(ctx, job)
		}
	}
}

func (p *Pool) process(ctx context.Context, job Job) {
	p.mu.Lock()
	_, exists := p.active[job.TripID]
	if !exists {
		p.mu.Unlock()
		return
	}
	p.mu.Unlock()

	workersActive.Add(ctx, 1)
	defer workersActive.Add(ctx, -1)

	start := time.Now()
	pushCtx, cancel := context.WithTimeout(ctx, p.timeout)
	err := p.gateway.RequestPosition(pushCtx, job.DriverID, job.TripID)
	cancel()

	dur := time.Since(start)
	if err != nil {
		recordPushResult(ctx, "error", dur)
	} else {
		recordPushResult(ctx, "success", dur)
	}

	queueSize.Add(ctx, -1)

	p.mu.Lock()
	delete(p.active, job.TripID)
	p.mu.Unlock()
}

func (p *Pool) Stop() {
	close(p.queue)
	p.wg.Wait()
}