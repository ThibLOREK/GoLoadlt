package main

import (
	"github.com/rinjold/go-etl-studio/internal/app"
)

func main() {
	worker, err := app.NewWorkerApp()
	if err != nil {
		panic(err)
	}

	if err := worker.Run(); err != nil {
		panic(err)
	}
}
