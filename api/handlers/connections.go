package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	connMgr "github.com/rinjold/go-etl-studio/internal/connections/manager"
	"github.com/rinjold/go-etl-studio/internal/connections/resolver"
)

// ConnectionHandler gère les endpoints de connexions.
type ConnectionHandler struct {
	Manager  *connMgr.Manager
	Resolver *resolver.Resolver
	Log      zerolog.Logger
}

// ListConnections GET /api/v1/connections
func (h *ConnectionHandler) ListConnections(w http.ResponseWriter, r *http.Request) {
	conns, err := h.Manager.List()
	if err != nil {
		h.Log.Error().Err(err).Msg("ListConnections")
		http.Error(w, `{"error":"erreur lecture connexions"}`, http.StatusInternalServerError)
		return
	}
	jsonResponse(w, conns, http.StatusOK)
}

// GetConnection GET /api/v1/connections/{connID}
func (h *ConnectionHandler) GetConnection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "connID")
	c, err := h.Manager.Load(id)
	if err != nil {
		http.Error(w, `{"error":"connexion introuvable"}`, http.StatusNotFound)
		return
	}
	jsonResponse(w, c, http.StatusOK)
}

// CreateConnection POST /api/v1/connections
func (h *ConnectionHandler) CreateConnection(w http.ResponseWriter, r *http.Request) {
	var c connMgr.Connection
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, `{"error":"JSON invalide"}`, http.StatusBadRequest)
		return
	}
	if c.ID == "" || c.Name == "" || c.Type == "" {
		http.Error(w, `{"error":"id, name et type obligatoires"}`, http.StatusBadRequest)
		return
	}
	if err := h.Manager.Save(&c); err != nil {
		h.Log.Error().Err(err).Msg("CreateConnection")
		http.Error(w, `{"error":"erreur sauvegarde connexion"}`, http.StatusInternalServerError)
		return
	}
	h.Log.Info().Str("id", c.ID).Str("type", c.Type).Msg("connexion créée")
	jsonResponse(w, c, http.StatusCreated)
}

// UpdateConnection PUT /api/v1/connections/{connID}
func (h *ConnectionHandler) UpdateConnection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "connID")
	var c connMgr.Connection
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, `{"error":"JSON invalide"}`, http.StatusBadRequest)
		return
	}
	c.ID = id
	if err := h.Manager.Save(&c); err != nil {
		http.Error(w, `{"error":"erreur sauvegarde connexion"}`, http.StatusInternalServerError)
		return
	}
	jsonResponse(w, c, http.StatusOK)
}

// DeleteConnection DELETE /api/v1/connections/{connID}
func (h *ConnectionHandler) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "connID")
	if err := h.Manager.Delete(id); err != nil {
		http.Error(w, `{"error":"erreur suppression connexion"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// TestConnection POST /api/v1/connections/{connID}/test
func (h *ConnectionHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "connID")
	resolved, err := h.Resolver.Resolve(id)
	if err != nil {
		h.Log.Warn().Str("id", id).Err(err).Msg("TestConnection échéc résolution")
		jsonResponse(w, map[string]any{
			"success": false,
			"error":   err.Error(),
		}, http.StatusOK)
		return
	}
	// TODO : implémenter le test réel de connexion SQL/REST selon resolved.Type
	h.Log.Info().Str("id", id).Str("host", resolved.Host).Msg("TestConnection résolution OK")
	jsonResponse(w, map[string]any{
		"success":  true,
		"host":     resolved.Host,
		"database": resolved.Database,
		"user":     resolved.User,
		"type":     resolved.Type,
	}, http.StatusOK)
}
