package app

import (
	"time"

	"github.com/rinjold/go-etl-studio/internal/config"
	"github.com/rinjold/go-etl-studio/internal/logger"
)

type WorkerApp struct {
	cfg config.Config
}

func NewWorkerApp() (*WorkerApp, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	_ = logger.New(cfg.AppEnv)
	return &WorkerApp{cfg: cfg}, nil
}

func (w *WorkerApp) Run() error {
	for {
		time.Sleep(5 * time.Second)
	}
}
