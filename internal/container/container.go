package container

import (
	"context"
	"time" // <-- Добавить

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/exaring/otelpgx"
	"github.com/razrabbesezdav-art/template/internal/config"
	"github.com/razrabbesezdav-art/template/internal/infra/postgres/transaction"
	"github.com/razrabbesezdav-art/template/internal/repository/trip"
	usecaseCreate "github.com/razrabbesezdav-art/template/internal/usecase/trip/create"
	usecaseFinish "github.com/razrabbesezdav-art/template/internal/usecase/trip/finish"
	usecaseGet "github.com/razrabbesezdav-art/template/internal/usecase/trip/get"
	usecasePosition "github.com/razrabbesezdav-art/template/internal/usecase/trip/position"
)

type Container struct {
	CreateUC       *usecaseCreate.UseCase
	GetUC          *usecaseGet.UseCase
	FinishUC       *usecaseFinish.UseCase
	PositionSaveUC *usecasePosition.UseCase
	PositionGetUC  *usecasePosition.GetUseCase

	DB *pgxpool.Pool
}

func New(ctx context.Context, cfg config.Config) (*Container, error) {
	pgxCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	pgxCfg.ConnConfig.Tracer = otelpgx.NewTracer()

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, err
	}

	txManager := transaction.NewManager(pool)
	tripRepo := trip.NewRepo(pool, cfg.DatabaseQueryTimeout)

	createUC := usecaseCreate.New(tripRepo, txManager)
	getUC := usecaseGet.New(tripRepo)
	finishUC := usecaseFinish.New(tripRepo, txManager)

	positionSaveUC := usecasePosition.New(tripRepo, txManager, nil, time.Now)
	positionGetUC := usecasePosition.NewGet(tripRepo)

	return &Container{
		CreateUC:       createUC,
		GetUC:          getUC,
		FinishUC:       finishUC,
		PositionSaveUC: positionSaveUC,
		PositionGetUC:  positionGetUC,
		DB:             pool,
	}, nil
}