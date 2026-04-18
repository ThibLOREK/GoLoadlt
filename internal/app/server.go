package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rinjold/go-etl-studio/api/handlers"
	"github.com/rinjold/go-etl-studio/api/middleware"
	"github.com/rinjold/go-etl-studio/internal/config"
	"github.com/rinjold/go-etl-studio/internal/logger"
)

type ServerApp struct {
	cfg    config.Config
	server *http.Server
}

func NewServerApp() (*ServerApp, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	log := logger.New(cfg.AppEnv)
	router := handlers.NewRouter(log)
	middleware.Apply(router)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: router,
	}

	return &ServerApp{cfg: cfg, server: server}, nil
}

func (a *ServerApp) Run() error {
	return a.server.ListenAndServe()
}

func (a *ServerApp) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}
