package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/rinjold/go-etl-studio/internal/connections/manager"
	"github.com/rinjold/go-etl-studio/internal/etl/engine"
	"github.com/rinjold/go-etl-studio/internal/xml/store"

	// Enregistrement de tous les blocs.
	_ "github.com/rinjold/go-etl-studio/internal/etl/blocks/sources"
	_ "github.com/rinjold/go-etl-studio/internal/etl/blocks/targets"
	_ "github.com/rinjold/go-etl-studio/internal/etl/blocks/transforms"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	projectsDir := getEnv("PROJECTS_DIR", "./projects")
	connsDir := getEnv("CONNECTIONS_DIR", "./connections")
	activeEnv := getEnv("ACTIVE_ENV", "dev")
	pollInterval := 5 * time.Second

	projectStore, err := store.NewProjectStore(projectsDir)
	if err != nil {
		log.Fatal().Err(err).Msg("store projets")
	}

	connManager, err := manager.New(connsDir, activeEnv)
	if err != nil {
		log.Fatal().Err(err).Msg("connexions")
	}

	executor := engine.NewExecutor(log.Logger, activeEnv)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info().Str("env", activeEnv).Dur("poll", pollInterval).Msg("GoLoadIt worker démarré")

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("arrêt du worker")
			return
		case <-ticker.C:
			runPendingJobs(ctx, projectStore, connManager, executor)
		}
	}
}

func runPendingJobs(
	ctx context.Context,
	ps *store.ProjectStore,
	mgr *manager.Manager,
	exec *engine.Executor,
) {
	// Lire les projets marqués comme "pending" dans un fichier d'index simple.
	// Pour le MVP le worker exécute les projets dont un fichier .run existe.
	entries, _ := os.ReadDir(ps.ProjectsDir())
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		runFile := ps.ProjectsDir() + "/" + e.Name() + "/run.trigger"
		if _, err := os.Stat(runFile); os.IsNotExist(err) {
			continue
		}
		_ = os.Remove(runFile)
		project, err := ps.Load(e.Name())
		if err != nil {
			log.Error().Str("project", e.Name()).Err(err).Msg("chargement projet")
			continue
		}
		// Injecter les DSN résolus dans les params des blocs sources/targets.
		injectDSNs(project, mgr)

		log.Info().Str("project", project.ID).Msg("exécution projet")
		report, err := exec.Execute(ctx, project)
		if err != nil {
			log.Error().Str("project", project.ID).Err(err).Msg("erreur exécution")
		} else {
			log.Info().Str("project", project.ID).
				Bool("success", report.Success).
				Dur("durée", report.EndedAt.Sub(report.StartedAt)).
				Msg("projet terminé")
		}
	}
}

// injectDSNs résout les connexions référencées et injecte le DSN dans les params des blocs.
func injectDSNs(project interface{ GetNodes() interface{} }, mgr *manager.Manager) {
	// L'injection est gérée au niveau du runner via le resolver.
	// Cette fonction est un placeholder pour l'évolution vers la résolution dynamique.
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
