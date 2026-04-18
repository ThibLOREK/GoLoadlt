package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	apimiddleware "github.com/rinjold/go-etl-studio/api/middleware"
	connMgr "github.com/rinjold/go-etl-studio/internal/connections/manager"
	"github.com/rinjold/go-etl-studio/internal/connections/resolver"
	"github.com/rinjold/go-etl-studio/internal/services"
	"github.com/rinjold/go-etl-studio/internal/xml/store"
)

func NewRouter(
	log zerolog.Logger,
	jwtSecret string,
	db *pgxpool.Pool,
	authService *services.AuthService,
	pipelineService *services.PipelineService,
	runService *services.RunService,
	scheduleService *services.ScheduleService,
	projectStore *store.Store,
	connManager *connMgr.Manager,
	connResolver *resolver.Resolver,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Public
	r.Post("/api/v1/auth/login", Login(authService))
	r.Post("/api/v1/auth/register", Register(authService))

	// Protected
	r.Group(func(protected chi.Router) {
		protected.Use(apimiddleware.Auth(jwtSecret))

		// --- Projets ETL (graphe de blocs + XML) ---
		ph := &ProjectHandler{Store: projectStore, Log: log}
		protected.Route("/api/v1/projects", func(pr chi.Router) {
			pr.Get("/", ph.ListProjects)
			pr.Post("/", ph.CreateProject)
			pr.Post("/import", ph.ImportXML)
			pr.Get("/{projectID}", ph.GetProject)
			pr.Put("/{projectID}", ph.UpdateProject)
			pr.Delete("/{projectID}", ph.DeleteProject)
			pr.Get("/{projectID}/xml", ph.ExportXML)
		})

		// --- Connexions multi-env ---
		ch := &ConnectionHandler{Manager: connManager, Resolver: connResolver, Log: log}
		protected.Route("/api/v1/connections", func(cr chi.Router) {
			cr.Get("/", ch.ListConnections)
			cr.Post("/", ch.CreateConnection)
			cr.Get("/{connID}", ch.GetConnection)
			cr.Put("/{connID}", ch.UpdateConnection)
			cr.Delete("/{connID}", ch.DeleteConnection)
			cr.Post("/{connID}/test", ch.TestConnection)
		})

		// --- Switch d'environnement global ---
		eh := &EnvironmentHandler{DB: db, Log: log}
		protected.Route("/api/v1/environment", func(er chi.Router) {
			er.Get("/", eh.GetEnvironment)
			er.Put("/", eh.SwitchEnvironment)
			er.Get("/history", eh.GetEnvironmentHistory)
		})

		// --- Catalogue des blocs disponibles ---
		protected.Get("/api/v1/blocks", ListBlocks())

		// --- Pipelines legacy (conservé pour compatibilité) ---
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