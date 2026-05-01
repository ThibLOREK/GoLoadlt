package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib" // enregistre le driver "pgx" pour database/sql
	"github.com/rs/zerolog"

	"github.com/rinjold/go-etl-studio/internal/connections"
	"github.com/rinjold/go-etl-studio/internal/connections/manager"
	"github.com/rinjold/go-etl-studio/internal/connections/resolver"
)

// validEnvs liste les valeurs d'environnement acceptées.
var validEnvs = map[string]bool{"dev": true, "preprod": true, "prod": true}

// ConnectionHandler gère les opérations CRUD sur les connexions et le switch d'env.
type ConnectionHandler struct {
	mgr *manager.Manager
	log zerolog.Logger
}

func NewConnectionHandler(m *manager.Manager, log zerolog.Logger) *ConnectionHandler {
	return &ConnectionHandler{mgr: m, log: log}
}

func (h *ConnectionHandler) List(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.mgr.List())
}

func (h *ConnectionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "connID")
	conn, err := h.mgr.Get(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, conn)
}

func (h *ConnectionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var conn connections.Connection
	if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if conn.ID == "" {
		conn.ID = uuid.NewString()
	}
	if err := h.mgr.Save(&conn); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, conn)
}

func (h *ConnectionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "connID")
	var conn connections.Connection
	if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	conn.ID = id
	if err := h.mgr.Save(&conn); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, conn)
}

func (h *ConnectionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "connID")
	if err := h.mgr.Delete(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Test vérifie qu'une connexion est réellement atteignable sur l'environnement actif.
func (h *ConnectionHandler) Test(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "connID")
	rc, err := resolver.Resolve(h.mgr, id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	start := time.Now()
	if err := pingConnection(ctx, rc); err != nil {
		h.log.Warn().Str("conn", id).Str("type", rc.Type).Err(err).Msg("ping échoué")
		writeError(w, http.StatusBadGateway, fmt.Sprintf("ping échoué (%s): %v", rc.Type, err))
		return
	}
	latency := time.Since(start).Milliseconds()

	h.log.Info().Str("conn", id).Str("type", rc.Type).Int64("ms", latency).Msg("ping ok")
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"type":    rc.Type,
		"host":    rc.Host,
		"db":      rc.Database,
		"env":     h.mgr.ActiveEnv,
		"latency": fmt.Sprintf("%dms", latency),
	})
}

// pingConnection tente une vraie connexion à la cible selon le type.
// - postgres : driver pgx/stdlib enregistré, sql.Open + PingContext.
// - mysql/mssql : drivers absents du module — erreur explicite cohérente avec les stubs sources.
// - rest : HTTP GET minimal avec le contexte fourni.
func pingConnection(ctx context.Context, rc *resolver.ResolvedConn) error {
	switch rc.Type {
	case "postgres":
		db, err := sql.Open("pgx", rc.DSN)
		if err != nil {
			return fmt.Errorf("postgres: ouverture driver: %w", err)
		}
		defer db.Close()
		if err := db.PingContext(ctx); err != nil {
			return fmt.Errorf("postgres: ping: %w", err)
		}
		return nil

	case "mysql":
		// go-sql-driver/mysql n'est pas dans le module — comportement explicite.
		return fmt.Errorf("mysql: driver non disponible dans ce build (ajoutez github.com/go-sql-driver/mysql)")

	case "mssql":
		// microsoft/go-mssqldb n'est pas dans le module — comportement explicite.
		return fmt.Errorf("mssql: driver non disponible dans ce build (ajoutez github.com/microsoft/go-mssqldb)")

	case "rest":
		if rc.Host == "" {
			return fmt.Errorf("rest: URL de base (host) manquante")
		}
		u, err := url.ParseRequestURI(rc.Host)
		if err != nil || u.Scheme == "" {
			return fmt.Errorf("rest: URL invalide: %q", rc.Host)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rc.Host, nil)
		if err != nil {
			return fmt.Errorf("rest: construction requête: %w", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("rest: requête: %w", err)
		}
		resp.Body.Close()
		return nil

	default:
		return fmt.Errorf("type de connexion non supporté: %q", rc.Type)
	}
}

// SwitchEnv bascule l'environnement actif globalement.
func (h *ConnectionHandler) SwitchEnv(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Env string `json:"env"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Env == "" {
		writeError(w, http.StatusBadRequest, "champ 'env' manquant (dev|preprod|prod)")
		return
	}
	if !validEnvs[body.Env] {
		writeError(w, http.StatusBadRequest,
			fmt.Sprintf("env invalide '%s' : valeurs acceptées dev|preprod|prod", body.Env))
		return
	}
	if err := h.mgr.SwitchEnv(body.Env); err != nil {
		h.log.Error().Err(err).Str("env", body.Env).Msg("impossible de persister l'env")
		writeError(w, http.StatusInternalServerError, "impossible de persister l'env: "+err.Error())
		return
	}
	h.log.Info().Str("env", body.Env).Msg("environnement basculé et persisté")
	writeJSON(w, http.StatusOK, map[string]string{"activeEnv": body.Env})
}

// GetEnv retourne l'environnement actif.
func (h *ConnectionHandler) GetEnv(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"activeEnv": h.mgr.ActiveEnv})
}
