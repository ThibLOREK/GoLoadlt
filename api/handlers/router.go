package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rinjold/go-etl-studio/internal/services"
	"github.com/rs/zerolog"
)

func NewRouter(log zerolog.Logger, pipelineService *services.PipelineService) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(api chi.Router) {
		api.Route("/pipelines", func(pr chi.Router) {
			pr.Get("/", ListPipelines(pipelineService))
			pr.Post("/", CreatePipeline(pipelineService))
			pr.Get("/{pipelineID}", GetPipeline(pipelineService))
			pr.Put("/{pipelineID}", UpdatePipeline(pipelineService))
			pr.Delete("/{pipelineID}", DeletePipeline(pipelineService))
			pr.Post("/{pipelineID}/runs", RunPipeline)
		})
	})

	_ = log
	return r
}
