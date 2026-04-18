package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func NewRouter(log zerolog.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/pipelines", ListPipelines)
		api.Post("/pipelines", CreatePipeline)
		api.Post("/runs/{pipelineID}", RunPipeline)
	})

	_ = log
	return r
}
