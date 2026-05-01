package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/rinjold/go-etl-studio/api/handlers"
	"github.com/rinjold/go-etl-studio/internal/connections/manager"
	"github.com/rinjold/go-etl-studio/internal/etl/engine"
	"github.com/rinjold/go-etl-studio/internal/orchestrator"
	"github.com/rinjold/go-etl-studio/internal/xml/store"

	// Enregistrement de tous les blocs via leurs init()
	_ "github.com/rinjold/go-etl-studio/internal/etl/blocks/sources"
	_ "github.com/rinjold/go-etl-studio/internal/etl/blocks/targets"
	_ "github.com/rinjold/go-etl-studio/internal/etl/blocks/transforms"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	projectsDir := getEnv("PROJECTS_DIR", "./projects")
	connsDir    := getEnv("CONNECTIONS_DIR", "./connections")
	activeEnv   := getEnv("ACTIVE_ENV", "dev")
	addr        := getEnv("SERVER_ADDR", ":8090")

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

	// Instancier l'Executor et l'Orchestrateur.
	// jobRepo est nil jusqu'à Sprint C (implémentation PostgreSQL + migration 005).
	// L'orchestrateur gère gracieusement un jobRepo nil en ne persistant pas les statuts.
	exec := engine.NewExecutor(log.Logger, activeEnv)
	orch := orchestrator.NewService(exec, projectStore.XMLStore(), nil)

	// RunHandler : jobRepo nil — GetLogs et GetReport retourneront 501 jusqu'à Sprint C.
	rh := handlers.NewRunHandler(orch, nil, log.Logger)

	// Construire le routeur via NewRouter (remplace le routeur inline précédent).
	r := handlers.NewRouter(log.Logger, projectStore, connManager, rh)

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
