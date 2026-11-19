package container

import (
	"context"
	"log/slog"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/razrabbesezdav-art/template/internal/config"
	"github.com/razrabbesezdav-art/template/internal/gateway/push"
	"github.com/razrabbesezdav-art/template/internal/infra/postgres/transaction"
	"github.com/razrabbesezdav-art/template/internal/repository/trip"
	usecaseCreate "github.com/razrabbesezdav-art/template/internal/usecase/trip/create"
	usecaseFinish "github.com/razrabbesezdav-art/template/internal/usecase/trip/finish"
	usecaseGet "github.com/razrabbesezdav-art/template/internal/usecase/trip/get"
	usecasePosition "github.com/razrabbesezdav-art/template/internal/usecase/trip/position"
	"github.com/razrabbesezdav-art/template/internal/worker/positions"
)

type Container struct {
	CreateUC       *usecaseCreate.UseCase
	GetUC          *usecaseGet.UseCase
	FinishUC       *usecaseFinish.UseCase
	PositionSaveUC *usecasePosition.UseCase
	PositionGetUC  *usecasePosition.GetUseCase

	PositionPool    *positions.Pool
	PositionScanner *positions.Scanner

	DB *pgxpool.Pool
}

func New(ctx context.Context, cfg config.Config, logger *slog.Logger) (*Container, error) {
	pgxCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	pgxCfg.ConnConfig.Tracer = otelpgx.NewTracer()

	dbPool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, err
	}

	txManager := transaction.NewManager(dbPool)
	tripRepo := trip.NewRepo(dbPool, cfg.DatabaseQueryTimeout)

	pushClient := push.NewClient(cfg.PushHTTPURL, cfg.PushRequestTimeout)

	workerPool := positions.NewPool(
		cfg.PositionWorkers,
		cfg.PositionQueueSize,
		pushClient,
		cfg.PushRequestTimeout,
	)

	createUC := usecaseCreate.New(tripRepo, txManager)
	getUC := usecaseGet.New(tripRepo)
	finishUC := usecaseFinish.New(tripRepo, txManager)
	positionSaveUC := usecasePosition.New(tripRepo, txManager, workerPool, time.Now)
	positionGetUC := usecasePosition.NewGet(tripRepo)

	scanner := positions.NewScanner(
		tripRepo,
		workerPool,
		cfg.PositionCheckInterval,
		cfg.PositionStaleAfter,
		logger,
	)

	return &Container{
		CreateUC:        createUC,
		GetUC:           getUC,
		FinishUC:        finishUC,
		PositionSaveUC:  positionSaveUC,
		PositionGetUC:   positionGetUC,
		PositionPool:    workerPool,
		PositionScanner: scanner,
		DB:              dbPool,
	}, nil
}