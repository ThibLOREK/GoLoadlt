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
	"github.com/rinjold/go-etl-studio/internal/connections"
	"github.com/rinjold/go-etl-studio/internal/connections/manager"
	"github.com/rinjold/go-etl-studio/internal/etl/engine"
	"github.com/rinjold/go-etl-studio/internal/orchestrator"
	"github.com/rinjold/go-etl-studio/internal/xml/store"
)

func setupConnectionServer(t *testing.T) (*httptest.Server, *manager.Manager) {
	t.Helper()

	projectStore, err := store.NewProjectStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewProjectStore: %v", err)
	}
	connsDir := t.TempDir()
	connManager, err := manager.New(connsDir, "dev")
	if err != nil {
		t.Fatalf("manager.New: %v", err)
	}

	log := zerolog.Nop()
	exec := engine.NewExecutor(log, "dev")
	orch := orchestrator.NewService(exec, projectStore, nil)
	rh := handlers.NewRunHandler(orch, nil, log)

	router := handlers.NewRouter(log, projectStore, connManager, rh)
	return httptest.NewServer(router), connManager
}

func TestConnectionCRUD(t *testing.T) {
	t.Parallel()
	srv, _ := setupConnectionServer(t)
	defer srv.Close()

	client := srv.Client()
	base := srv.URL + "/api/v1/connections"

	// ── 1. POST /api/v1/connections → 201 ─────────────────────────────────
	conn := connections.Connection{
		Name: "test-pg",
		Type: "postgres",
		Envs: map[string]connections.ConnEnv{
			"dev": {
				Name:     "dev",
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				User:     "testuser",
			},
		},
	}
	body, _ := json.Marshal(conn)

	resp1, err := client.Post(base+"/", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST connections: %v", err)
	}
	defer resp1.Body.Close()

	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("POST connections: attendu 201, reçu %d", resp1.StatusCode)
	}

	var created connections.Connection
	if err := json.NewDecoder(resp1.Body).Decode(&created); err != nil {
		t.Fatalf("POST connections: décodage: %v", err)
	}
	if created.ID == "" {
		t.Fatal("POST connections: id absent")
	}
	connID := created.ID
	connURL := fmt.Sprintf("%s/%s", base, connID)

	// ── 2. GET /api/v1/connections/{connID} → 200 ─────────────────────────
	resp2, err := client.Get(connURL)
	if err != nil {
		t.Fatalf("GET connection: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("GET connection: attendu 200, reçu %d", resp2.StatusCode)
	}
	var fetched connections.Connection
	if err := json.NewDecoder(resp2.Body).Decode(&fetched); err != nil {
		t.Fatalf("GET connection: décodage: %v", err)
	}
	if fetched.ID != connID {
		t.Errorf("GET connection: id attendu %q, reçu %q", connID, fetched.ID)
	}

	// ── 3. PUT /api/v1/connections/{connID} → 200 ─────────────────────────
	conn.Name = "test-pg-updated"
	updateBody, _ := json.Marshal(conn)
	req3, _ := http.NewRequest(http.MethodPut, connURL, bytes.NewReader(updateBody))
	req3.Header.Set("Content-Type", "application/json")
	resp3, err := client.Do(req3)
	if err != nil {
		t.Fatalf("PUT connection: %v", err)
	}
	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusOK {
		t.Fatalf("PUT connection: attendu 200, reçu %d", resp3.StatusCode)
	}
	var putResult connections.Connection
	if err := json.NewDecoder(resp3.Body).Decode(&putResult); err != nil {
		t.Fatalf("PUT connection: décodage: %v", err)
	}
	if putResult.Name != "test-pg-updated" {
		t.Errorf("PUT connection: name attendu 'test-pg-updated', reçu %q", putResult.Name)
	}

	// ── 4. POST /api/v1/connections/{connID}/test → 200 ou 4xx ────────────
	// Le ping va échouer (pas de vrai postgres) — on accepte 200 ou 4xx mais
	// jamais 5xx non géré ni panic.
	resp4, err := client.Post(fmt.Sprintf("%s/test", connURL), "application/json", nil)
	if err != nil {
		t.Fatalf("POST connection/test: %v", err)
	}
	defer resp4.Body.Close()

	if resp4.StatusCode == http.StatusInternalServerError {
		t.Errorf("POST connection/test: 500 inattendu (doit être 200, 400, 422 ou 502)")
	}
	// Le handler retourne 502 Bad Gateway si le ping échoue — c'est acceptable.
	acceptable := map[int]bool{
		http.StatusOK:                  true,
		http.StatusBadRequest:          true,
		http.StatusUnprocessableEntity: true,
		http.StatusBadGateway:          true,
	}
	if !acceptable[resp4.StatusCode] {
		t.Errorf("POST connection/test: statut inattendu %d", resp4.StatusCode)
	}

	// ── 5. DELETE /api/v1/connections/{connID} → 204 ─────────────────────
	req5, _ := http.NewRequest(http.MethodDelete, connURL, nil)
	resp5, err := client.Do(req5)
	if err != nil {
		t.Fatalf("DELETE connection: %v", err)
	}
	resp5.Body.Close()

	if resp5.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE connection: attendu 204, reçu %d", resp5.StatusCode)
	}

	// ── 6. GET après suppression → 404 ──────────────────────────────────
	resp6, err := client.Get(connURL)
	if err != nil {
		t.Fatalf("GET connection after delete: %v", err)
	}
	defer resp6.Body.Close()

	if resp6.StatusCode != http.StatusNotFound {
		t.Fatalf("GET connection after delete: attendu 404, reçu %d", resp6.StatusCode)
	}
}
