package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/rinjold/go-etl-studio/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	worker, err := app.NewWorkerApp()
	if err != nil {
		panic(err)
	}

	if err := worker.Run(ctx); err != nil {
		panic(err)
	}
}
