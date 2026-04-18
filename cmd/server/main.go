package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/rinjold/go-etl-studio/api/handlers"
	"github.com/rinjold/go-etl-studio/internal/connections/manager"
	"github.com/rinjold/go-etl-studio/internal/xml/store"

	// Enregistrement de tous les blocs via leurs init()
	_ "github.com/rinjold/go-etl-studio/internal/etl/blocks/sources"
	_ "github.com/rinjold/go-etl-studio/internal/etl/blocks/targets"
	_ "github.com/rinjold/go-etl-studio/internal/etl/blocks/transforms"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	projectsDir := getEnv("PROJECTS_DIR", "./projects")
	connsDir := getEnv("CONNECTIONS_DIR", "./connections")
	activeEnv := getEnv("ACTIVE_ENV", "dev")
	addr := getEnv("SERVER_ADDR", ":8080")

	// Initialiser le store XML des projets.
	projectStore, err := store.NewProjectStore(projectsDir)
	if err != nil {
		log.Fatal().Err(err).Msg("impossible de créer le store de projets")
	}

	// Initialiser le gestionnaire de connexions.
	connManager, err := manager.New(connsDir, activeEnv)
	if err != nil {
		log.Fatal().Err(err).Msg("impossible de charger les connexions")
	}

	// Router Chi.
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type"},
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Handlers.
	ph := handlers.NewProjectHandler(projectStore, connManager, log.Logger)
	ch := handlers.NewConnectionHandler(connManager, log.Logger)

	r.Route("/api/v1", func(r chi.Router) {
		// Projets.
		r.Get("/projects", ph.List)
		r.Post("/projects", ph.Create)
		r.Get("/projects/{projectID}", ph.Get)
		r.Put("/projects/{projectID}", ph.Update)
		r.Delete("/projects/{projectID}", ph.Delete)
		r.Post("/projects/{projectID}/run", ph.Run)
		r.Get("/projects/{projectID}/xml", ph.ExportXML)
		r.Post("/projects/import", ph.ImportXML)

		// Catalogue de blocs (pour l'UI).
		r.Get("/catalogue", ph.Catalogue)

		// Connexions.
		r.Get("/connections", ch.List)
		r.Post("/connections", ch.Create)
		r.Get("/connections/{connID}", ch.Get)
		r.Put("/connections/{connID}", ch.Update)
		r.Delete("/connections/{connID}", ch.Delete)
		r.Post("/connections/{connID}/test", ch.Test)

		// Switch d'environnement global.
		r.Put("/environment", ch.SwitchEnv)
		r.Get("/environment", ch.GetEnv)
	})

	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		log.Info().Str("addr", addr).Str("env", activeEnv).Msg("GoLoadIt server démarré")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("erreur serveur")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Info().Msg("arrêt du serveur...")
	_ = srv.Shutdown(ctx)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
