package app

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	connMgr "github.com/rinjold/go-etl-studio/internal/connections/manager"
	"github.com/rinjold/go-etl-studio/internal/connections/resolver"
	"github.com/rinjold/go-etl-studio/internal/config"
	"github.com/rinjold/go-etl-studio/internal/logger"
	"github.com/rinjold/go-etl-studio/internal/services"
	"github.com/rinjold/go-etl-studio/internal/storage"
	"github.com/rinjold/go-etl-studio/internal/storage/memory"
	"github.com/rinjold/go-etl-studio/internal/xml/store"
)

type Container struct {
	Config          config.Config
	Logger          zerolog.Logger
	PostgresPool    *pgxpool.Pool
	AuthService     *services.AuthService
	PipelineService *services.PipelineService
	RunService      *services.RunService
	ScheduleService *services.ScheduleService
	ProjectStore    *store.Store
	ConnManager     *connMgr.Manager
	ConnResolver    *resolver.Resolver
}

func BuildContainer(ctx context.Context) (*Container, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	log := logger.New(cfg.AppEnv)

	projectsDir := getEnvOr("PROJECTS_DIR", "./projects")
	connectionsDir := getEnvOr("CONNECTIONS_DIR", "./connections")

	projectStore := store.New(projectsDir)
	connManager := connMgr.New(connectionsDir)

	var (
		userRepo     services.UserRepository
		pipelineRepo services.PipelineRepository
		runRepo      services.RunRepository
		scheduleRepo services.ScheduleRepository
		pool         *pgxpool.Pool
	)

	if cfg.AppEnv == "development" && cfg.PostgresDSN == "" {
		log.Warn().Msg("[dev mode] no POSTGRES_DSN — using in-memory repositories")
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

	// Lire l'env actif depuis la DB (ou fallback "dev")
	activeEnv := "dev"
	if pool != nil {
		var envFromDB string
		err := pool.QueryRow(ctx, `SELECT active_env FROM environment_context WHERE id = 1`).Scan(&envFromDB)
		if err == nil && envFromDB != "" {
			activeEnv = envFromDB
		}
	}
	log.Info().Str("activeEnv", activeEnv).Msg("environnement actif chargé")

	connResolver := resolver.New(connManager, activeEnv)

	return &Container{
		Config:          cfg,
		Logger:          log,
		PostgresPool:    pool,
		AuthService:     services.NewAuthService(userRepo, cfg.JWTSecret),
		PipelineService: services.NewPipelineService(pipelineRepo),
		RunService:      services.NewRunService(runRepo, pipelineRepo),
		ScheduleService: services.NewScheduleService(scheduleRepo, services.NewRunService(runRepo, pipelineRepo)),
		ProjectStore:    projectStore,
		ConnManager:     connManager,
		ConnResolver:    connResolver,
	}, nil
}

func getEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}