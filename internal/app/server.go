package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rinjold/go-etl-studio/api/handlers"
	"github.com/rinjold/go-etl-studio/api/middleware"
	"github.com/rinjold/go-etl-studio/internal/config"
)

type ServerApp struct {
	cfg       config.Config
	server    *http.Server
	container *Container
}

func NewServerApp() (*ServerApp, error) {
	ctx := context.Background()
	container, err := BuildContainer(ctx)
	if err != nil {
		return nil, err
	}

	router := handlers.NewRouter(container.Logger, container.PipelineService)
	middleware.Apply(router)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", container.Config.HTTPPort),
		Handler: router,
	}

	return &ServerApp{cfg: container.Config, server: server, container: container}, nil
}

func (a *ServerApp) Run() error {
	return a.server.ListenAndServe()
}

func (a *ServerApp) Shutdown(ctx context.Context) error {
	if a.container != nil && a.container.PostgresPool != nil {
		a.container.PostgresPool.Close()
	}
	return a.server.Shutdown(ctx)
}
