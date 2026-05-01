//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	"github.com/rinjold/go-etl-studio/api/handlers"
	"github.com/rinjold/go-etl-studio/internal/connections/manager"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
	"github.com/rinjold/go-etl-studio/internal/etl/engine"
	"github.com/rinjold/go-etl-studio/internal/orchestrator"
	"github.com/rinjold/go-etl-studio/internal/xml/store"
)

// setupProjectServer construit un httptest.Server avec les vraies dépendances
// en mémoire/tmpdir. Aucune base PostgreSQL nécessaire.
func setupProjectServer(t *testing.T) *httptest.Server {
	t.Helper()

	projectStore, err := store.NewProjectStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewProjectStore: %v", err)
	}
	connManager, err := manager.New(t.TempDir(), "dev")
	if err != nil {
		t.Fatalf("manager.New: %v", err)
	}

	log := zerolog.Nop()
	exec := engine.NewExecutor(log, "dev")
	orch := orchestrator.NewService(exec, projectStore, nil)
	rh := handlers.NewRunHandler(orch, nil, log)

	router := handlers.NewRouter(log, projectStore, connManager, rh)
	return httptest.NewServer(router)
}

func TestProjectCRUD(t *testing.T) {
	t.Parallel()
	srv := setupProjectServer(t)
	defer srv.Close()

	client := srv.Client()
	base := srv.URL + "/api/v1/projects"

	// ── 1. POST /api/v1/projects/ ────────────────────────────────────────────
	p := contracts.Project{Name: "integration-test-project"}
	body, _ := json.Marshal(p)

	resp, err := client.Post(base+"/", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST projects: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST projects: attendu 201, reçu %d", resp.StatusCode)
	}

	var created contracts.Project
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("POST projects: décodage: %v", err)
	}
	if created.ID == "" {
		t.Fatal("POST projects: id absent de la réponse")
	}
	if created.Name != p.Name {
		t.Errorf("POST projects: name attendu %q, reçu %q", p.Name, created.Name)
	}

	projectID := created.ID
	projectURL := fmt.Sprintf("%s/%s", base, projectID)

	// ── 2. GET /api/v1/projects/{id} ─────────────────────────────────────────
	resp2, err := client.Get(projectURL)
	if err != nil {
		t.Fatalf("GET project: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("GET project: attendu 200, reçu %d", resp2.StatusCode)
	}

	var fetched contracts.Project
	if err := json.NewDecoder(resp2.Body).Decode(&fetched); err != nil {
		t.Fatalf("GET project: décodage: %v", err)
	}
	if fetched.ID != projectID {
		t.Errorf("GET project: id attendu %q, reçu %q", projectID, fetched.ID)
	}

	// ── 3. PUT /api/v1/projects/{id} ─────────────────────────────────────────
	updated := contracts.Project{Name: "integration-test-project-updated"}
	updateBody, _ := json.Marshal(updated)

	req3, _ := http.NewRequest(http.MethodPut, projectURL, bytes.NewReader(updateBody))
	req3.Header.Set("Content-Type", "application/json")
	resp3, err := client.Do(req3)
	if err != nil {
		t.Fatalf("PUT project: %v", err)
	}
	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusOK {
		t.Fatalf("PUT project: attendu 200, reçu %d", resp3.StatusCode)
	}

	var putResult contracts.Project
	if err := json.NewDecoder(resp3.Body).Decode(&putResult); err != nil {
		t.Fatalf("PUT project: décodage: %v", err)
	}
	if putResult.Name != "integration-test-project-updated" {
		t.Errorf("PUT project: name attendu 'integration-test-project-updated', reçu %q", putResult.Name)
	}

	// ── 4. GET /api/v1/projects/ (list) ──────────────────────────────────────
	resp4, err := client.Get(base + "/")
	if err != nil {
		t.Fatalf("GET projects list: %v", err)
	}
	defer resp4.Body.Close()

	if resp4.StatusCode != http.StatusOK {
		t.Fatalf("GET projects list: attendu 200, reçu %d", resp4.StatusCode)
	}

	// Le handler retourne []*contracts.Project directement
	var list []*contracts.Project
	if err := json.NewDecoder(resp4.Body).Decode(&list); err != nil {
		t.Fatalf("GET projects list: décodage: %v", err)
	}

	found := false
	for _, item := range list {
		if item.ID == projectID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GET projects list: projet %q absent de la liste (%d éléments)", projectID, len(list))
	}

	// ── 5. DELETE /api/v1/projects/{id} ──────────────────────────────────────
	req5, _ := http.NewRequest(http.MethodDelete, projectURL, nil)
	resp5, err := client.Do(req5)
	if err != nil {
		t.Fatalf("DELETE project: %v", err)
	}
	resp5.Body.Close()

	if resp5.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE project: attendu 204, reçu %d", resp5.StatusCode)
	}

	// ── 6. GET /api/v1/projects/{id} après suppression → 404 ─────────────────
	resp6, err := client.Get(projectURL)
	if err != nil {
		t.Fatalf("GET project after delete: %v", err)
	}
	defer resp6.Body.Close()

	if resp6.StatusCode != http.StatusNotFound {
		t.Fatalf("GET project after delete: attendu 404, reçu %d", resp6.StatusCode)
	}
}
