package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"

	"github.com/razrabbesezdav-art/template/internal/config"
	"github.com/razrabbesezdav-art/template/internal/container"
	api "github.com/razrabbesezdav-art/template/internal/generated"
	httplayer "github.com/razrabbesezdav-art/template/internal/http"
	"github.com/razrabbesezdav-art/template/internal/http/trip/create"
	"github.com/razrabbesezdav-art/template/internal/http/trip/finish"
	"github.com/razrabbesezdav-art/template/internal/http/trip/get"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
	api.HandlerFromMux(router, chiRouter)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           chiRouter,
		ReadTimeout:       cfg.HTTPReadTimeout,
		ReadHeaderTimeout: cfg.HTTPReadTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("HTTP server listening on %s", cfg.HTTPAddr)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server: %v", err)
		}
	case <-ctx.Done():
		log.Println("shutdown signal received")
		
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
			_ = server.Close()
		}
		log.Println("server stopped")
	}
}