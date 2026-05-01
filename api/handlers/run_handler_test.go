package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/rinjold/go-etl-studio/api/handlers"
	"github.com/rinjold/go-etl-studio/internal/etl/engine"
	"github.com/rinjold/go-etl-studio/internal/jobs"
	"github.com/rinjold/go-etl-studio/internal/orchestrator"
	"github.com/rinjold/go-etl-studio/internal/xml/store"
)

// ── helpers ──────────────────────────────────────────────────────────────────

// newTestRunHandler crée un RunHandler avec un vrai orchestrateur en mémoire
// (store temporaire, pas de jobRepo) pour les tests unitaires.
func newTestRunHandler(t *testing.T) (*handlers.RunHandler, *store.ProjectStore) {
	t.Helper()
	dir := t.TempDir()
	ps, err := store.NewProjectStore(dir)
	if err != nil {
		t.Fatalf("NewProjectStore: %v", err)
	}
	exec := engine.NewExecutor(zerolog.Nop(), "dev")
	orch := orchestrator.NewService(exec, ps, nil)
	rh := handlers.NewRunHandler(orch, nil, zerolog.Nop())
	return rh, ps
}

// chiCtx injecte un param URL Chi dans le contexte de la requête.
func chiCtx(r *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// ── tests ────────────────────────────────────────────────────────────────────

// TestStartRun_ProjectNotFound vérifie que StartRun retourne 500 si le projet n'existe pas.
func TestStartRun_ProjectNotFound(t *testing.T) {
	rh, _ := newTestRunHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/unknown/runs", nil)
	req = chiCtx(req, "projectID", "unknown")
	w := httptest.NewRecorder()

	rh.StartRun(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("StartRun: attendu 500, reçu %d", w.Code)
	}
	var body map[string]string
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["error"] == "" {
		t.Error("StartRun: champ 'error' absent du corps")
	}
}

// TestCancelRun_NoJobRepo vérifie que CancelRun retourne 500 quand jobRepo est nil (Sprint B).
func TestCancelRun_NoJobRepo(t *testing.T) {
	rh, _ := newTestRunHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/runs/run-42", nil)
	req = chiCtx(req, "runID", "run-42")
	w := httptest.NewRecorder()

	rh.CancelRun(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("CancelRun: attendu 500 (jobRepo nil), reçu %d", w.Code)
	}
}

// TestListRuns_NoJobRepo vérifie que ListRuns retourne une liste vide (pas d'erreur) sans jobRepo.
func TestListRuns_NoJobRepo(t *testing.T) {
	rh, _ := newTestRunHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/p1/runs", nil)
	req = chiCtx(req, "projectID", "p1")
	w := httptest.NewRecorder()

	rh.ListRuns(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ListRuns: attendu 200, reçu %d", w.Code)
	}
	var runs []jobs.Run
	if err := json.NewDecoder(w.Body).Decode(&runs); err != nil {
		t.Fatalf("ListRuns: décodage JSON: %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("ListRuns: attendu slice vide, reçu %d éléments", len(runs))
	}
}

// TestGetRun_NoJobRepo vérifie que GetRun retourne 404 quand jobRepo est nil.
func TestGetRun_NoJobRepo(t *testing.T) {
	rh, _ := newTestRunHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/runs/run-42", nil)
	req = chiCtx(req, "runID", "run-42")
	w := httptest.NewRecorder()

	rh.GetRun(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GetRun: attendu 404, reçu %d", w.Code)
	}
}

// TestGetLogs_Returns501_WhenNoJobRepo vérifie le comportement 501 Sprint B.
func TestGetLogs_Returns501_WhenNoJobRepo(t *testing.T) {
	rh, _ := newTestRunHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/runs/run-42/logs", nil)
	req = chiCtx(req, "runID", "run-42")
	w := httptest.NewRecorder()

	rh.GetLogs(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("GetLogs: attendu 501, reçu %d", w.Code)
	}
}

// TestGetReport_Returns501_WhenNoJobRepo vérifie le comportement 501 Sprint B.
func TestGetReport_Returns501_WhenNoJobRepo(t *testing.T) {
	rh, _ := newTestRunHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/runs/run-42/report", nil)
	req = chiCtx(req, "runID", "run-42")
	w := httptest.NewRecorder()

	rh.GetReport(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("GetReport: attendu 501, reçu %d", w.Code)
	}
}
