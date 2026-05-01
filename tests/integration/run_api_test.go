//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/rinjold/go-etl-studio/api/handlers"
	"github.com/rinjold/go-etl-studio/internal/connections/manager"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
	"github.com/rinjold/go-etl-studio/internal/etl/engine"
	"github.com/rinjold/go-etl-studio/internal/jobs"
	"github.com/rinjold/go-etl-studio/internal/orchestrator"
	"github.com/rinjold/go-etl-studio/internal/xml/store"
)

// ──────────────────────────────────────────────────────────────────────────────
// mockJobRepo — implémentation en mémoire de jobs.Repository
// ──────────────────────────────────────────────────────────────────────────────

type mockJobRepo struct {
	mu   sync.Mutex
	runs map[string]*jobs.Run
	logs map[string][]jobs.RunLogEntry
	seq  int
}

func newMockJobRepo() *mockJobRepo {
	return &mockJobRepo{
		runs: make(map[string]*jobs.Run),
		logs: make(map[string][]jobs.RunLogEntry),
	}
}

func (m *mockJobRepo) Create(_ context.Context, projectID string) (*jobs.Run, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	run := &jobs.Run{
		ID:        fmt.Sprintf("run-%d", m.seq),
		ProjectID: projectID,
		Status:    jobs.Pending,
		StartedAt: time.Now(),
	}
	m.runs[run.ID] = run
	return run, nil
}

func (m *mockJobRepo) SetStatus(_ context.Context, runID string, status jobs.Status) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r, ok := m.runs[runID]; ok {
		r.Status = status
	}
	return nil
}

func (m *mockJobRepo) GetByID(_ context.Context, runID string) (*jobs.Run, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.runs[runID]
	if !ok {
		return nil, fmt.Errorf("run %q introuvable", runID)
	}
	return r, nil
}

func (m *mockJobRepo) ListByProject(_ context.Context, projectID string) ([]jobs.Run, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []jobs.Run
	for _, r := range m.runs {
		if r.ProjectID == projectID {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (m *mockJobRepo) AppendLog(_ context.Context, entry jobs.RunLogEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs[entry.RunID] = append(m.logs[entry.RunID], entry)
	return nil
}

func (m *mockJobRepo) GetLogs(_ context.Context, runID string) ([]jobs.RunLogEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.logs[runID], nil // slice nil → JSON [] grâce à json.Marshal
}

// ──────────────────────────────────────────────────────────────────────────────
// setup
// ──────────────────────────────────────────────────────────────────────────────

type runTestEnv struct {
	srv        *httptest.Server
	client     *http.Client
	projectID  string
	mockRepo   *mockJobRepo
}

func setupRunServer(t *testing.T) *runTestEnv {
	t.Helper()

	projectStore, err := store.NewProjectStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewProjectStore: %v", err)
	}
	connManager, err := manager.New(t.TempDir(), "dev")
	if err != nil {
		t.Fatalf("manager.New: %v", err)
	}

	mockRepo := newMockJobRepo()
	log := zerolog.Nop()
	exec := engine.NewExecutor(log, "dev")
	orch := orchestrator.NewService(exec, projectStore, mockRepo)
	rh := handlers.NewRunHandler(orch, mockRepo, log)

	router := handlers.NewRouter(log, projectStore, connManager, rh)
	srv := httptest.NewServer(router)

	// Créer un projet de test via l'API
	p := contracts.Project{Name: "run-integration-test"}
	body, _ := json.Marshal(p)
	resp, err := srv.Client().Post(
		srv.URL+"/api/v1/projects/",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		srv.Close()
		t.Fatalf("création projet: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		srv.Close()
		t.Fatalf("création projet: attendu 201, reçu %d", resp.StatusCode)
	}
	var created contracts.Project
	_ = json.NewDecoder(resp.Body).Decode(&created)

	return &runTestEnv{
		srv:       srv,
		client:    srv.Client(),
		projectID: created.ID,
		mockRepo:  mockRepo,
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// tests
// ──────────────────────────────────────────────────────────────────────────────

func TestRunPipeline(t *testing.T) {
	t.Parallel()
	env := setupRunServer(t)
	defer env.srv.Close()

	base := env.srv.URL
	projectID := env.projectID

	// ── 1. POST /api/v1/projects/{id}/runs → 201 ────────────────────────────
	resp1, err := env.client.Post(
		fmt.Sprintf("%s/api/v1/projects/%s/runs", base, projectID),
		"application/json",
		nil,
	)
	if err != nil {
		t.Fatalf("POST runs: %v", err)
	}
	defer resp1.Body.Close()

	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("POST runs: attendu 201, reçu %d", resp1.StatusCode)
	}

	// Le projet est vide (pas de blocs) donc l'exécution réussit avec un report vide.
	// On vérifie juste que le champ projectId est présent.
	var runResp map[string]any
	if err := json.NewDecoder(resp1.Body).Decode(&runResp); err != nil {
		t.Fatalf("POST runs: décodage: %v", err)
	}
	if runResp["projectId"] == "" || runResp["projectId"] == nil {
		t.Errorf("POST runs: champ projectId absent: %v", runResp)
	}

	// Récupérer le runID depuis le mock (le seul run créé)
	var runID string
	env.mockRepo.mu.Lock()
	for id := range env.mockRepo.runs {
		runID = id
	}
	env.mockRepo.mu.Unlock()
	if runID == "" {
		t.Fatal("aucun run trouvé dans le mock après StartRun")
	}

	// ── 2. GET /api/v1/projects/{id}/runs → 200 + run dans la liste ─────────
	resp2, err := env.client.Get(fmt.Sprintf("%s/api/v1/projects/%s/runs", base, projectID))
	if err != nil {
		t.Fatalf("GET runs list: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("GET runs list: attendu 200, reçu %d", resp2.StatusCode)
	}
	var runList []jobs.Run
	if err := json.NewDecoder(resp2.Body).Decode(&runList); err != nil {
		t.Fatalf("GET runs list: décodage: %v", err)
	}
	if len(runList) == 0 {
		t.Error("GET runs list: liste vide, attendu au moins 1 run")
	}

	// ── 3. GET /api/v1/runs/{runID} → 200 + détail ─────────────────────────
	resp3, err := env.client.Get(fmt.Sprintf("%s/api/v1/runs/%s", base, runID))
	if err != nil {
		t.Fatalf("GET run: %v", err)
	}
	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusOK {
		t.Fatalf("GET run: attendu 200, reçu %d", resp3.StatusCode)
	}
	var runDetail jobs.Run
	if err := json.NewDecoder(resp3.Body).Decode(&runDetail); err != nil {
		t.Fatalf("GET run: décodage: %v", err)
	}
	if runDetail.ID != runID {
		t.Errorf("GET run: id attendu %q, reçu %q", runID, runDetail.ID)
	}

	// ── 4. GET /api/v1/runs/{runID}/logs → 200 + tableau (vide ok) ────────
	resp4, err := env.client.Get(fmt.Sprintf("%s/api/v1/runs/%s/logs", base, runID))
	if err != nil {
		t.Fatalf("GET logs: %v", err)
	}
	defer resp4.Body.Close()

	if resp4.StatusCode != http.StatusOK {
		t.Fatalf("GET logs: attendu 200, reçu %d", resp4.StatusCode)
	}
	// Doit être un tableau JSON (même vide)
	var logEntries []json.RawMessage
	if err := json.NewDecoder(resp4.Body).Decode(&logEntries); err != nil {
		t.Fatalf("GET logs: décodage tableau: %v", err)
	}

	// ── 5. GET /api/v1/runs/{runID}/report → 200 + jobs.Run ───────────────
	resp5, err := env.client.Get(fmt.Sprintf("%s/api/v1/runs/%s/report", base, runID))
	if err != nil {
		t.Fatalf("GET report: %v", err)
	}
	defer resp5.Body.Close()

	if resp5.StatusCode != http.StatusOK {
		t.Fatalf("GET report: attendu 200, reçu %d", resp5.StatusCode)
	}
	var reportRun jobs.Run
	if err := json.NewDecoder(resp5.Body).Decode(&reportRun); err != nil {
		t.Fatalf("GET report: décodage: %v", err)
	}
	if reportRun.ProjectID != projectID {
		t.Errorf("GET report: projectId attendu %q, reçu %q", projectID, reportRun.ProjectID)
	}
}
