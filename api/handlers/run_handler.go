package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/rinjold/go-etl-studio/internal/jobs"
	"github.com/rinjold/go-etl-studio/internal/orchestrator"
	"github.com/rinjold/go-etl-studio/pkg/dto"
)

// RunHandler gère les endpoints de lancement, suivi et historique des runs.
type RunHandler struct {
	orch    *orchestrator.Service
	jobRepo jobs.Repository
	log     zerolog.Logger
}

// NewRunHandler crée un RunHandler.
func NewRunHandler(orch *orchestrator.Service, jobRepo jobs.Repository, log zerolog.Logger) *RunHandler {
	return &RunHandler{orch: orch, jobRepo: jobRepo, log: log}
}

// StartRun lance l'exécution d'un projet ETL.
// POST /api/v1/projects/{projectID}/runs
func (h *RunHandler) StartRun(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")

	report, err := h.orch.RunProject(r.Context(), projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	status := "succeeded"
	if report != nil && !report.Success {
		status = "failed"
	}

	writeJSON(w, http.StatusCreated, dto.RunResponse{
		ProjectID: projectID,
		Status:    status,
		Report:    report,
	})
}

// CancelRun annule un run en cours.
// DELETE /api/v1/runs/{runID}
func (h *RunHandler) CancelRun(w http.ResponseWriter, r *http.Request) {
	runID := chi.URLParam(r, "runID")

	if err := h.orch.CancelRun(r.Context(), runID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListRuns retourne l'historique des runs d'un projet.
// GET /api/v1/projects/{projectID}/runs
func (h *RunHandler) ListRuns(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")

	runs, err := h.orch.ListRuns(r.Context(), projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

// GetRun retourne le détail d'un run.
// GET /api/v1/runs/{runID}
func (h *RunHandler) GetRun(w http.ResponseWriter, r *http.Request) {
	runID := chi.URLParam(r, "runID")

	run, err := h.orch.GetRunStatus(r.Context(), runID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, run)
}

// GetLogs retourne les logs structurés d'un run par bloc.
// GET /api/v1/runs/{runID}/logs
// Note Sprint B : GetLogs est dans l'interface jobs.Repository mais l'implémentation
// PostgreSQL (Sprint C) n'est pas encore disponible — retourne 501 si jobRepo est nil.
func (h *RunHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	if h.jobRepo == nil {
		writeError(w, http.StatusNotImplemented, "logs non disponibles avant Sprint C")
		return
	}

	runID := chi.URLParam(r, "runID")

	logs, err := h.jobRepo.GetLogs(r.Context(), runID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

// GetReport retourne le rapport d'exécution complet d'un run.
// GET /api/v1/runs/{runID}/report
// Note Sprint B : sans implémentation PostgreSQL, on retourne les métadonnées du run.
func (h *RunHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	if h.jobRepo == nil {
		writeError(w, http.StatusNotImplemented, "report non disponible avant Sprint C")
		return
	}

	runID := chi.URLParam(r, "runID")

	run, err := h.jobRepo.GetByID(r.Context(), runID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, run)
}
