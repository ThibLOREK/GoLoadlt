package app

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rinjold/go-etl-studio/internal/config"
	"github.com/rinjold/go-etl-studio/internal/logger"
	"github.com/rinjold/go-etl-studio/internal/services"
	"github.com/rinjold/go-etl-studio/internal/storage"
	"github.com/rinjold/go-etl-studio/internal/storage/memory"
	"github.com/rs/zerolog"
)

type Container struct {
	Config          config.Config
	Logger          zerolog.Logger
	PostgresPool    *pgxpool.Pool
	AuthService     *services.AuthService
	PipelineService *services.PipelineService
	RunService      *services.RunService
	ScheduleService *services.ScheduleService
}

func BuildContainer(ctx context.Context) (*Container, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	log := logger.New(cfg.AppEnv)

	var (
		userRepo     services.UserRepository
		pipelineRepo services.PipelineRepository
		runRepo      services.RunRepository
		scheduleRepo services.ScheduleRepository
		pool         *pgxpool.Pool
	)

	if cfg.AppEnv == "development" && cfg.PostgresDSN == "" {
		log.Warn().Msg("[dev mode] no POSTGRES_DSN set — using in-memory repositories (data will not persist)")
		userRepo = memory.NewUserRepository()
		pipelineRepo = memory.NewPipelineRepository()
		runRepo = memory.NewRunRepository()
		scheduleRepo = memory.NewScheduleRepository()
	} else {
		pool, err = storage.NewPostgresPool(ctx, cfg.PostgresDSN)
		if err != nil {
			return nil, err
		}
		userRepo = storage.NewUserRepository(pool)
		pipelineRepo = storage.NewPipelineRepository(pool)
		runRepo = storage.NewRunRepository(pool)
		scheduleRepo = storage.NewScheduleRepository(pool)
	}

	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	pipelineService := services.NewPipelineService(pipelineRepo)
	runService := services.NewRunService(runRepo, pipelineRepo)
	scheduleService := services.NewScheduleService(scheduleRepo, runService)

	return &Container{
		Config:          cfg,
		Logger:          log,
		PostgresPool:    pool,
		AuthService:     authService,
		PipelineService: pipelineService,
		RunService:      runService,
		ScheduleService: scheduleService,
	}, nil
}
