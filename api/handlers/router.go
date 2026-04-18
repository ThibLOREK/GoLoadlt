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

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// --- Projets ETL ---
	ph := NewProjectHandler(projectStore, connManager, log)
	r.Route("/api/v1/projects", func(pr chi.Router) {
		pr.Get("/", ph.List)
		pr.Post("/", ph.Create)
		pr.Post("/import", ph.ImportXML)
		pr.Get("/{projectID}", ph.Get)
		pr.Put("/{projectID}", ph.Update)
		pr.Delete("/{projectID}", ph.Delete)
		pr.Post("/{projectID}/run", ph.Run)
		pr.Get("/{projectID}/xml", ph.ExportXML)
	})

	// --- Catalogue des blocs ---
	r.Get("/api/v1/catalogue", ph.Catalogue)

	// --- Connexions multi-env ---
	ch := NewConnectionHandler(connManager, log)
	r.Route("/api/v1/connections", func(cr chi.Router) {
		cr.Get("/", ch.List)
		cr.Post("/", ch.Create)
		cr.Get("/{connID}", ch.Get)
		cr.Put("/{connID}", ch.Update)
		cr.Delete("/{connID}", ch.Delete)
		cr.Post("/{connID}/test", ch.Test)
	})

	// --- Switch d'environnement ---
	r.Get("/api/v1/environment", ch.GetEnv)
	r.Put("/api/v1/environment", ch.SwitchEnv)

	return r
}
