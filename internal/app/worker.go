package app

import (
	"context"
	"time"
)

type WorkerApp struct {
	container *Container
}

func NewWorkerApp() (*WorkerApp, error) {
	container, err := BuildContainer(context.Background())
	if err != nil {
		return nil, err
	}
	return &WorkerApp{container: container}, nil
}

func (w *WorkerApp) Run() error {
	for {
		time.Sleep(5 * time.Second)
	}
}
