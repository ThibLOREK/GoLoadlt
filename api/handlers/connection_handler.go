package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/rinjold/go-etl-studio/internal/connections"
	"github.com/rinjold/go-etl-studio/internal/connections/manager"
	"github.com/rinjold/go-etl-studio/internal/connections/resolver"
)

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

// Test vérifie qu'une connexion est atteignable sur l'environnement actif.
func (h *ConnectionHandler) Test(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "connID")
	rc, err := resolver.Resolve(h.mgr, id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"type":   rc.Type,
		"host":   rc.Host,
		"db":     rc.Database,
		"env":    h.mgr.ActiveEnv,
	})
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
	h.mgr.SwitchEnv(body.Env)
	h.log.Info().Str("env", body.Env).Msg("environnement basculé")
	writeJSON(w, http.StatusOK, map[string]string{"activeEnv": body.Env})
}

// GetEnv retourne l'environnement actif.
func (h *ConnectionHandler) GetEnv(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"activeEnv": h.mgr.ActiveEnv})
}
