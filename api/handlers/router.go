package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rinjold/go-etl-studio/api/middleware"
	"github.com/rinjold/go-etl-studio/internal/services"
	"github.com/rs/zerolog"
)

func NewRouter(
	log zerolog.Logger,
	jwtSecret string,
	authService *services.AuthService,
	pipelineService *services.PipelineService,
	runService *services.RunService,
	scheduleService *services.ScheduleService,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Public routes
	r.Post("/api/v1/auth/login", Login(authService))
	r.Post("/api/v1/auth/register", Register(authService))

	// Protected routes
	r.Group(func(protected chi.Router) {
		protected.Use(middleware.Auth(jwtSecret))

		protected.Route("/api/v1/pipelines", func(pr chi.Router) {
			pr.Get("/", ListPipelines(pipelineService))
			pr.Post("/", CreatePipeline(pipelineService))
			pr.Get("/{pipelineID}", GetPipeline(pipelineService))
			pr.Put("/{pipelineID}", UpdatePipeline(pipelineService))
			pr.Delete("/{pipelineID}", DeletePipeline(pipelineService))

			pr.Post("/{pipelineID}/runs", ScheduleRun(runService))
			pr.Get("/{pipelineID}/runs", ListRuns(runService))
			pr.Get("/{pipelineID}/runs/{runID}", GetRun(runService))

			pr.Put("/{pipelineID}/schedule", UpsertSchedule(scheduleService))
			pr.Get("/{pipelineID}/schedule", GetSchedule(scheduleService))
			pr.Delete("/{pipelineID}/schedule", DeleteSchedule(scheduleService))
		})
	})

	_ = log
	return r
}
