package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
	"github.com/rinjold/go-etl-studio/internal/xml/parser"
	"github.com/rinjold/go-etl-studio/internal/xml/serializer"
	"github.com/rinjold/go-etl-studio/internal/xml/store"
)

// ProjectHandler gère les endpoints de projets ETL.
type ProjectHandler struct {
	Store *store.Store
	Log   zerolog.Logger
}

// ListProjects GET /api/v1/projects
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	ids, err := h.Store.ListIDs()
	if err != nil {
		h.Log.Error().Err(err).Msg("ListProjects")
		http.Error(w, `{"error":"erreur lecture projets"}`, http.StatusInternalServerError)
		return
	}
	var projects []*contracts.Project
	for _, id := range ids {
		p, err := h.Store.Load(id)
		if err != nil {
			continue
		}
		projects = append(projects, p)
	}
	jsonResponse(w, projects, http.StatusOK)
}

// GetProject GET /api/v1/projects/{projectID}
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "projectID")
	p, err := h.Store.Load(id)
	if err != nil {
		http.Error(w, `{"error":"projet introuvable"}`, http.StatusNotFound)
		return
	}
	jsonResponse(w, p, http.StatusOK)
}

// CreateProject POST /api/v1/projects
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var p contracts.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, `{"error":"JSON invalide"}`, http.StatusBadRequest)
		return
	}
	if p.ID == "" || p.Name == "" {
		http.Error(w, `{"error":"id et name obligatoires"}`, http.StatusBadRequest)
		return
	}
	p.Version = 1
	sha, err := h.Store.Save(&p)
	if err != nil {
		h.Log.Error().Err(err).Msg("CreateProject")
		http.Error(w, `{"error":"erreur sauvegarde"}`, http.StatusInternalServerError)
		return
	}
	h.Log.Info().Str("id", p.ID).Str("sha", sha).Msg("projet créé")
	jsonResponse(w, p, http.StatusCreated)
}

// UpdateProject PUT /api/v1/projects/{projectID}
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "projectID")
	existing, err := h.Store.Load(id)
	if err != nil {
		http.Error(w, `{"error":"projet introuvable"}`, http.StatusNotFound)
		return
	}
	var p contracts.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, `{"error":"JSON invalide"}`, http.StatusBadRequest)
		return
	}
	p.ID = id
	p.Version = existing.Version + 1
	sha, err := h.Store.Save(&p)
	if err != nil {
		h.Log.Error().Err(err).Msg("UpdateProject")
		http.Error(w, `{"error":"erreur sauvegarde"}`, http.StatusInternalServerError)
		return
	}
	h.Log.Info().Str("id", p.ID).Int("version", p.Version).Str("sha", sha).Msg("projet mis à jour")
	jsonResponse(w, p, http.StatusOK)
}

// DeleteProject DELETE /api/v1/projects/{projectID}
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "projectID")
	if err := h.Store.Delete(id); err != nil {
		http.Error(w, `{"error":"erreur suppression"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ExportXML GET /api/v1/projects/{projectID}/xml
func (h *ProjectHandler) ExportXML(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "projectID")
	p, err := h.Store.Load(id)
	if err != nil {
		http.Error(w, `{"error":"projet introuvable"}`, http.StatusNotFound)
		return
	}
	data, err := serializer.Serialize(p)
	if err != nil {
		http.Error(w, `{"error":"erreur sérialisation"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Disposition", `attachment; filename="`+id+`.xml"`)
	w.Write(data)
}

// ImportXML POST /api/v1/projects/import
func (h *ProjectHandler) ImportXML(w http.ResponseWriter, r *http.Request) {
	body := make([]byte, r.ContentLength)
	if _, err := r.Body.Read(body); err != nil && r.ContentLength > 0 {
		http.Error(w, `{"error":"lecture body"}`, http.StatusBadRequest)
		return
	}
	p, err := parser.ParseBytes(body)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	sha, err := h.Store.Save(p)
	if err != nil {
		http.Error(w, `{"error":"erreur sauvegarde"}`, http.StatusInternalServerError)
		return
	}
	h.Log.Info().Str("id", p.ID).Str("sha", sha).Msg("projet importé via XML")
	jsonResponse(w, p, http.StatusCreated)
}

func jsonResponse(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
