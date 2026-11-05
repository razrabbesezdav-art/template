package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/razrabbesezdav-art/template/internal/config"
	"github.com/razrabbesezdav-art/template/internal/container"
	api "github.com/razrabbesezdav-art/template/internal/generated"
	httplayer "github.com/razrabbesezdav-art/template/internal/http"
	"github.com/razrabbesezdav-art/template/internal/http/middleware"
	"github.com/razrabbesezdav-art/template/internal/http/trip/create"
	"github.com/razrabbesezdav-art/template/internal/http/trip/finish"
	"github.com/razrabbesezdav-art/template/internal/http/trip/get"
	"github.com/razrabbesezdav-art/template/internal/infra/telemetry"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger, shutdownTelemetry, err := telemetry.Init(ctx, "trip-service", cfg.LogLevel)
	if err != nil {
		log.Fatalf("init telemetry: %v", err)
	}
	defer func() {
		_ = shutdownTelemetry(context.Background())
	}()


	c, err := container.New(ctx, cfg)
	if err != nil {
		log.Fatalf("init container: %v", err)
	}
	defer c.DB.Close()


	createHandler := create.NewHandler(c.CreateUC)
	getHandler := get.NewHandler(c.GetUC)
	finishHandler := finish.NewHandler(c.FinishUC)

	router := httplayer.NewRouter(createHandler, getHandler, finishHandler)
	chiRouter := chi.NewRouter()


	chiRouter.Use(middleware.Logging(logger))
	chiRouter.Use(middleware.Metrics())
	

	tracedRouter := otelhttp.NewHandler(chiRouter, "http.server", otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents))

	api.HandlerFromMux(router, chiRouter)


	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           tracedRouter,
		ReadTimeout:       cfg.HTTPReadTimeout,
		ReadHeaderTimeout: cfg.HTTPReadTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("HTTP server listening", slog.String("addr", cfg.HTTPAddr))
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server error", slog.String("error", err.Error()))
		}
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}
}