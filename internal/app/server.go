package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rinjold/go-etl-studio/api/handlers"
	"github.com/rinjold/go-etl-studio/internal/config"
	"github.com/rinjold/go-etl-studio/internal/telemetry"
)

type ServerApp struct {
	cfg         config.Config
	server      *http.Server
	container   *Container
	telProvider *telemetry.Provider
}

func NewServerApp() (*ServerApp, error) {
	ctx := context.Background()
	container, err := BuildContainer(ctx)
	if err != nil {
		return nil, err
	}

	telProvider, err := telemetry.Init(ctx, container.Config.AppName)
	if err != nil {
		return nil, err
	}

	router := handlers.NewRouter(
		container.Logger,
		container.Config.JWTSecret,
		container.AuthService,
		container.PipelineService,
		container.RunService,
		container.ScheduleService,
	)

	// Observability endpoints (registered after NewRouter which handles its own middlewares)
	router.Handle("/metrics", telemetry.MetricsHandler())
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"time":   time.Now().UTC(),
			"app":    container.Config.AppName,
			"env":    container.Config.AppEnv,
		})
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", container.Config.HTTPPort),
		Handler: router,
	}

	return &ServerApp{
		cfg:         container.Config,
		server:      server,
		container:   container,
		telProvider: telProvider,
	}, nil
}

func (a *ServerApp) Run() error {
	a.container.Logger.Info().Str("addr", a.server.Addr).Msg("server started")
	return a.server.ListenAndServe()
}

func (a *ServerApp) Shutdown(ctx context.Context) error {
	if a.container.PostgresPool != nil {
		a.container.PostgresPool.Close()
	}
	if a.telProvider != nil {
		_ = a.telProvider.Shutdown(ctx)
	}
	return a.server.Shutdown(ctx)
}
