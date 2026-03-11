package app

import (
	grpcapp "gw-exchanger/internal/app/grpc"
	"gw-exchanger/internal/config"
	"gw-exchanger/internal/storage/postgres"
	pgxdriver "gw-exchanger/pkg/pgx-driver"
	"log/slog"
)

type App struct {
	GRPCServer *grpcapp.App

	Postgres *pgxdriver.Postgres
}

func New(cfg *config.Config, log *slog.Logger) *App {
	pCfg := cfg.StorageConfig.Postgres

	// INIT CONNECTION DB

	storage, err := pgxdriver.New(
		pCfg.DSN,
		log,
		pgxdriver.MaxPoolSize(pCfg.MaxOpenConns),
		pgxdriver.MinConns(pCfg.MaxIdleConns),
		pgxdriver.MaxConnIdleTime(pCfg.MaxIdleTime))

	if err != nil {
		panic(err)
	}

	// INIT REPOSITORY

	exRepository := postgres.NewExchangeRatesRepository(log, storage)

	// INIT GRPC server

	grpcApp := grpcapp.New(log, exRepository, cfg.GRPCConfig.Port)

	return &App{
		GRPCServer: grpcApp,
		Postgres:   storage,
	}
}
