package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chirouter "github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Logging возвращает middleware для логирования запросов через slog.
func Logging(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}

			level := slog.LevelInfo
			if status >= 500 {
				level = slog.LevelError
			}

			route := chiRoute(r)
			if route == "" {
				route = r.URL.Path
			}

			log.LogAttrs(r.Context(), level, "request handled",
				slog.String("method", r.Method),
				slog.String("route", route),
				slog.Int("status", status),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()),
			)
		})
	}
}

// Метрики (Пункт 6.3)
var (
	meter           = otel.Meter("http.server")
	httpDuration, _ = meter.Int64Histogram("http.server.request.duration", metric.WithUnit("ms"))
	httpActive, _   = meter.Int64UpDownCounter("http.server.active_requests", metric.WithUnit("{request}"))
)

// Metrics возвращает middleware для сбора HTTP-метрик.
func Metrics() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			route := chiRoute(r)
			if route == "" {
				route = r.URL.Path
			}

			attrs := metric.WithAttributes(
				attribute.String("http.request.method", r.Method),
				attribute.String("http.route", route),
			)

			httpActive.Add(ctx, 1, attrs)
			defer httpActive.Add(ctx, -1, attrs)

			start := time.Now()
			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}

			httpDuration.Record(ctx, time.Since(start).Milliseconds(), metric.WithAttributes(
				attribute.String("http.request.method", r.Method),
				attribute.String("http.route", route),
				attribute.Int("http.response.status_code", status),
			))
		})
	}
}

// Recover ловит паники и логирует их.
func Recover(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered", slog.Any("panic", rec))
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func chiRoute(r *http.Request) string {
	if routeCtx := chirouter.RouteContext(r.Context()); routeCtx != nil {
		return routeCtx.RoutePattern()
	}
	return ""
}