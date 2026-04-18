package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/rinjold/go-etl-studio/internal/connections/manager"
	connresolver "github.com/rinjold/go-etl-studio/internal/connections/resolver"
	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
	"github.com/rinjold/go-etl-studio/internal/etl/engine"
	"github.com/rinjold/go-etl-studio/internal/xml/parser"
	"github.com/rinjold/go-etl-studio/internal/xml/serializer"
	"github.com/rinjold/go-etl-studio/internal/xml/store"
)

// ProjectHandler gère les opérations CRUD et d'exécution sur les projets ETL.
type ProjectHandler struct {
	store   *store.ProjectStore
	mgr     *manager.Manager
	log     zerolog.Logger
}

func NewProjectHandler(s *store.ProjectStore, m *manager.Manager, log zerolog.Logger) *ProjectHandler {
	return &ProjectHandler{store: s, mgr: m, log: log}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.store.ListAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "projectID")
	p, err := h.store.Load(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var p contracts.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "body JSON invalide: "+err.Error())
		return
	}
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	p.Version = 1
	if err := h.store.Save(&p); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.log.Info().Str("id", p.ID).Str("name", p.Name).Msg("projet créé")
	writeJSON(w, http.StatusCreated, p)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "projectID")
	existing, err := h.store.Load(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	var updated contracts.Project
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	updated.ID = existing.ID
	updated.Version = existing.Version + 1
	if err := h.store.Save(&updated); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "projectID")
	if err := h.store.Delete(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Run exécute un projet immédiatement avec injection des connexions résolues.
func (h *ProjectHandler) Run(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "projectID")
	p, err := h.store.Load(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err := engine.InjectResolvedConnections(p, func(connID string) (*connresolver.ResolvedConn, error) {
		return connresolver.Resolve(h.mgr, connID)
	}); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	exec := engine.NewExecutor(h.log, h.mgr.ActiveEnv)
	start := time.Now()
	report, err := exec.Execute(r.Context(), p)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"projectID": id,
		"success":   report.Success,
		"duration":  time.Since(start).String(),
		"results":   report.Results,
	})
}

func (h *ProjectHandler) ExportXML(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "projectID")
	p, err := h.store.Load(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	data, err := serializer.SerializeProject(p)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/xml")
	w.Write(data)
}

func (h *ProjectHandler) ImportXML(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	p, err := parser.ParseProjectBytes(data)
	if err != nil {
		writeError(w, http.StatusBadRequest, "XML invalide: "+err.Error())
		return
	}
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	if err := h.store.Save(p); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *ProjectHandler) Catalogue(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, blocks.Catalogue())
}
