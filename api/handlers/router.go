package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"

	"github.com/rinjold/go-etl-studio/internal/connections/manager"
	"github.com/rinjold/go-etl-studio/internal/xml/store"
)

// NewRouter construit le routeur Chi et branche tous les handlers.
func NewRouter(
	log zerolog.Logger,
	projectStore *store.ProjectStore,
	connManager *manager.Manager,
	// runHandler est optionnel (nil tant que Sprint C n'a pas injecté jobRepo)
	runHandler *RunHandler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	ph := NewProjectHandler(projectStore, connManager, log)
	r.Route("/api/v1/projects", func(pr chi.Router) {
		pr.Get("/", ph.List)
		pr.Post("/", ph.Create)
		pr.Post("/import", ph.ImportXML)
		pr.Get("/{projectID}", ph.Get)
		pr.Put("/{projectID}", ph.Update)
		pr.Delete("/{projectID}", ph.Delete)
		pr.Post("/{projectID}/run", ph.Run)   // existant — exécution synchrone directe
		pr.Get("/{projectID}/xml", ph.ExportXML)
		pr.Get("/{projectID}/csv-preview", ph.CSVPreview)
		pr.Get("/{projectID}/csv-scan", ph.CSVScan)

		// Phase 7 — Runs via orchestrateur
		if runHandler != nil {
			pr.Post("/{projectID}/runs", runHandler.StartRun)
			pr.Get("/{projectID}/runs", runHandler.ListRuns)
		}
	})

	r.Get("/api/v1/catalogue", ph.Catalogue)

	ch := NewConnectionHandler(connManager, log)
	r.Route("/api/v1/connections", func(cr chi.Router) {
		cr.Get("/", ch.List)
		cr.Post("/", ch.Create)
		cr.Get("/{connID}", ch.Get)
		cr.Put("/{connID}", ch.Update)
		cr.Delete("/{connID}", ch.Delete)
		cr.Post("/{connID}/test", ch.Test)
	})

	r.Get("/api/v1/environment", ch.GetEnv)
	r.Put("/api/v1/environment", ch.SwitchEnv)

	// Phase 7 — Routes runs standalone
	if runHandler != nil {
		r.Route("/api/v1/runs", func(rr chi.Router) {
			rr.Get("/{runID}", runHandler.GetRun)
			rr.Delete("/{runID}", runHandler.CancelRun)
			rr.Get("/{runID}/logs", runHandler.GetLogs)
			rr.Get("/{runID}/report", runHandler.GetReport)
		})
	}

	return r
}
