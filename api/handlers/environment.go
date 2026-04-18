package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// EnvironmentHandler gère le switch d'environnement global.
type EnvironmentHandler struct {
	DB  *pgxpool.Pool
	Log zerolog.Logger
}

type switchEnvRequest struct {
	Env       string `json:"env"`       // "dev", "preprod", "prod"
	ChangedBy string `json:"changedBy"` // utilisateur qui fait le switch
}

// GetEnvironment GET /api/v1/environment
func (h *EnvironmentHandler) GetEnvironment(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var activeEnv string
	err := h.DB.QueryRow(ctx,
		`SELECT active_env FROM environment_context WHERE id = 1`,
	).Scan(&activeEnv)
	if err != nil {
		h.Log.Error().Err(err).Msg("GetEnvironment")
		http.Error(w, `{"error":"erreur lecture environnement"}`, http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"activeEnv": activeEnv}, http.StatusOK)
}

// SwitchEnvironment PUT /api/v1/environment
func (h *EnvironmentHandler) SwitchEnvironment(w http.ResponseWriter, r *http.Request) {
	var req switchEnvRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"JSON invalide"}`, http.StatusBadRequest)
		return
	}

	valid := map[string]bool{"dev": true, "preprod": true, "prod": true}
	if !valid[req.Env] {
		http.Error(w, `{"error":"env invalide : valeurs acceptées : dev, preprod, prod"}`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Lire l'env actuel pour l'historique.
	var previousEnv string
	h.DB.QueryRow(ctx, `SELECT active_env FROM environment_context WHERE id = 1`).Scan(&previousEnv)

	// Mise à jour atomique en base.
	_, err := h.DB.Exec(ctx,
		`UPDATE environment_context SET active_env = $1, updated_at = now(), updated_by = $2 WHERE id = 1`,
		req.Env, req.ChangedBy,
	)
	if err != nil {
		h.Log.Error().Err(err).Msg("SwitchEnvironment update")
		http.Error(w, `{"error":"erreur mise à jour environnement"}`, http.StatusInternalServerError)
		return
	}

	// Enregistrer dans l'historique.
	h.DB.Exec(ctx,
		`INSERT INTO environment_history (previous_env, new_env, changed_by) VALUES ($1, $2, $3)`,
		previousEnv, req.Env, req.ChangedBy,
	)

	h.Log.Info().Str("from", previousEnv).Str("to", req.Env).Str("by", req.ChangedBy).Msg("switch environnement")

	jsonResponse(w, map[string]string{
		"activeEnv": req.Env,
		"message":   fmt.Sprintf("environnement basculé de '%s' vers '%s'", previousEnv, req.Env),
	}, http.StatusOK)
}

// GetEnvironmentHistory GET /api/v1/environment/history
func (h *EnvironmentHandler) GetEnvironmentHistory(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := h.DB.Query(ctx,
		`SELECT id, previous_env, new_env, changed_at, changed_by
		 FROM environment_history
		 ORDER BY changed_at DESC
		 LIMIT 50`,
	)
	if err != nil {
		http.Error(w, `{"error":"erreur lecture historique"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type histEntry struct {
		ID          string    `json:"id"`
		PreviousEnv string    `json:"previousEnv"`
		NewEnv      string    `json:"newEnv"`
		ChangedAt   time.Time `json:"changedAt"`
		ChangedBy   string    `json:"changedBy"`
	}
	var entries []histEntry
	for rows.Next() {
		var e histEntry
		rows.Scan(&e.ID, &e.PreviousEnv, &e.NewEnv, &e.ChangedAt, &e.ChangedBy)
		entries = append(entries, e)
	}
	jsonResponse(w, entries, http.StatusOK)
}
