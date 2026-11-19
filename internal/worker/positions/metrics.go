package positions

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	meter = otel.Meter("worker.positions")

	staleChecks, _ = meter.Int64Counter(
		"tripgo.stale_checks",
		metric.WithDescription("Результаты проверки устаревших поездок"),
	)

	positionRequests, _ = meter.Int64Counter(
		"tripgo.position_requests",
		metric.WithDescription("Результаты отправки пушей"),
	)

	positionRequestDuration, _ = meter.Int64Histogram(
		"tripgo.position_requests.duration",
		metric.WithUnit("ms"),
	)

	queueSize, _ = meter.Int64UpDownCounter(
		"tripgo.position_requests.queue.size",
	)

	workersActive, _ = meter.Int64UpDownCounter(
		"tripgo.position_requests.workers.active",
	)
)

func recordStaleCheck(ctx context.Context, result string) {
	staleChecks.Add(ctx, 1, metric.WithAttributes(
		attribute.String("result", result),
	))
}

func recordPushResult(ctx context.Context, result string, dur time.Duration) {
	positionRequests.Add(ctx, 1, metric.WithAttributes(
		attribute.String("result", result),
	))
	positionRequestDuration.Record(ctx, dur.Milliseconds())
}