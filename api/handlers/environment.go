package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/rinjold/go-etl-studio/internal/connections/manager"
)

// EnvironmentHandler gère le switch d'environnement global via le manager de connexions.
type EnvironmentHandler struct {
	mgr *manager.Manager
	log zerolog.Logger
}

func NewEnvironmentHandler(m *manager.Manager, log zerolog.Logger) *EnvironmentHandler {
	return &EnvironmentHandler{mgr: m, log: log}
}

type switchEnvRequest struct {
	Env       string `json:"env"`
	ChangedBy string `json:"changedBy"`
}

// GetEnvironment GET /api/v1/environment
func (h *EnvironmentHandler) GetEnvironment(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"activeEnv": h.mgr.ActiveEnv})
}

// SwitchEnvironment PUT /api/v1/environment
func (h *EnvironmentHandler) SwitchEnvironment(w http.ResponseWriter, r *http.Request) {
	var req switchEnvRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Env == "" {
		writeError(w, http.StatusBadRequest, "champ 'env' manquant (dev|preprod|prod)")
		return
	}
	valid := map[string]bool{"dev": true, "preprod": true, "prod": true}
	if !valid[req.Env] {
		writeError(w, http.StatusBadRequest, "env invalide : valeurs acceptées : dev, preprod, prod")
		return
	}

	prev := h.mgr.ActiveEnv
	h.mgr.SwitchEnv(req.Env)
	h.log.Info().Str("from", prev).Str("to", req.Env).Msg("switch environnement")

	writeJSON(w, http.StatusOK, map[string]string{
		"activeEnv": req.Env,
		"message":   fmt.Sprintf("environnement basculé de '%s' vers '%s'", prev, req.Env),
	})
}

// GetEnvironmentHistory GET /api/v1/environment/history
func (h *EnvironmentHandler) GetEnvironmentHistory(w http.ResponseWriter, r *http.Request) {
	// Historique non persisté en base dans cette version — retourne une liste vide.
	type histEntry struct {
		Env       string    `json:"env"`
		ChangedAt time.Time `json:"changedAt"`
		ChangedBy string    `json:"changedBy"`
	}
	writeJSON(w, http.StatusOK, []histEntry{})
}
