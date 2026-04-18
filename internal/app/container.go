package app

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rinjold/go-etl-studio/internal/config"
	"github.com/rinjold/go-etl-studio/internal/logger"
	"github.com/rinjold/go-etl-studio/internal/services"
	"github.com/rinjold/go-etl-studio/internal/storage"
	"github.com/rs/zerolog"
)

type Container struct {
	Config          config.Config
	Logger          zerolog.Logger
	PostgresPool    *pgxpool.Pool
	PipelineService *services.PipelineService
}

func BuildContainer(ctx context.Context) (*Container, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	log := logger.New(cfg.AppEnv)

	pool, err := storage.NewPostgresPool(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, err
	}

	repo := storage.NewPipelineRepository(pool)
	pipelineService := services.NewPipelineService(repo)

	return &Container{
		Config:          cfg,
		Logger:          log,
		PostgresPool:    pool,
		PipelineService: pipelineService,
	}, nil
}
