//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
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

func setupEnvServer(t *testing.T) (*httptest.Server, *manager.Manager) {
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

func TestEnvironmentGetAndSwitch(t *testing.T) {
	t.Parallel()
	srv, connManager := setupEnvServer(t)
	defer srv.Close()

	client := srv.Client()
	envURL := srv.URL + "/api/v1/environment"

	// ── 1. GET /api/v1/environment → 200 + activeEnv présent ──────────────
	resp1, err := client.Get(envURL)
	if err != nil {
		t.Fatalf("GET environment: %v", err)
	}
	defer resp1.Body.Close()

	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("GET environment: attendu 200, reçu %d", resp1.StatusCode)
	}
	var envResp map[string]string
	if err := json.NewDecoder(resp1.Body).Decode(&envResp); err != nil {
		t.Fatalf("GET environment: décodage: %v", err)
	}
	activeEnv, ok := envResp["activeEnv"]
	if !ok || activeEnv == "" {
		t.Fatalf("GET environment: champ 'activeEnv' absent ou vide: %v", envResp)
	}
	if activeEnv != "dev" {
		t.Errorf("GET environment: attendu 'dev' par défaut, reçu %q", activeEnv)
	}

	// ── 2. PUT /api/v1/environment → switch dev → preprod ──────────────────
	switchBody, _ := json.Marshal(map[string]string{"env": "preprod"})
	req2, _ := http.NewRequest(http.MethodPut, envURL, bytes.NewReader(switchBody))
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := client.Do(req2)
	if err != nil {
		t.Fatalf("PUT environment: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("PUT environment (preprod): attendu 200, reçu %d", resp2.StatusCode)
	}
	var putResp map[string]string
	if err := json.NewDecoder(resp2.Body).Decode(&putResp); err != nil {
		t.Fatalf("PUT environment: décodage: %v", err)
	}
	if putResp["activeEnv"] != "preprod" {
		t.Errorf("PUT environment: activeEnv attendu 'preprod', reçu %q", putResp["activeEnv"])
	}

	// ── 3. GET pour confirmer le switch ──────────────────────────────────
	resp3, err := client.Get(envURL)
	if err != nil {
		t.Fatalf("GET environment (après switch): %v", err)
	}
	defer resp3.Body.Close()

	var confirmResp map[string]string
	if err := json.NewDecoder(resp3.Body).Decode(&confirmResp); err != nil {
		t.Fatalf("GET environment (après switch): décodage: %v", err)
	}
	if confirmResp["activeEnv"] != "preprod" {
		t.Errorf("GET environment (après switch): attendu 'preprod', reçu %q", confirmResp["activeEnv"])
	}

	// ── 4. Vérifier que le manager inémoire reflète le switch ──────────────
	if connManager.ActiveEnv != "preprod" {
		t.Errorf("manager.ActiveEnv: attendu 'preprod', reçu %q", connManager.ActiveEnv)
	}

	// ── 5. PUT preprod → dev (retour à l'état initial) ───────────────────
	switchBack, _ := json.Marshal(map[string]string{"env": "dev"})
	req5, _ := http.NewRequest(http.MethodPut, envURL, bytes.NewReader(switchBack))
	req5.Header.Set("Content-Type", "application/json")
	resp5, err := client.Do(req5)
	if err != nil {
		t.Fatalf("PUT environment (retour dev): %v", err)
	}
	defer resp5.Body.Close()

	if resp5.StatusCode != http.StatusOK {
		t.Fatalf("PUT environment (retour dev): attendu 200, reçu %d", resp5.StatusCode)
	}
	if connManager.ActiveEnv != "dev" {
		t.Errorf("manager.ActiveEnv après retour: attendu 'dev', reçu %q", connManager.ActiveEnv)
	}
}

func TestEnvironmentInvalidSwitch(t *testing.T) {
	t.Parallel()
	srv, _ := setupEnvServer(t)
	defer srv.Close()

	client := srv.Client()
	envURL := srv.URL + "/api/v1/environment"

	// env invalide → 400
	badBody, _ := json.Marshal(map[string]string{"env": "staging"})
	req, _ := http.NewRequest(http.MethodPut, envURL, bytes.NewReader(badBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("PUT environment (invalide): %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("PUT environment (invalide): attendu 400, reçu %d", resp.StatusCode)
	}
}

func TestEnvironmentConnectionResolution(t *testing.T) {
	t.Parallel()
	srv, connManager := setupEnvServer(t)
	defer srv.Close()

	client := srv.Client()

	// Créer une connexion avec deux environnements
	conn := connections.Connection{
		Name: "env-resolution-test",
		Type: "postgres",
		Envs: map[string]connections.ConnEnv{
			"dev":    {Name: "dev", Host: "dev-host", Port: 5432, Database: "devdb", User: "devuser"},
			"preprod": {Name: "preprod", Host: "preprod-host", Port: 5432, Database: "preproddb", User: "preproduser"},
		},
	}
	body, _ := json.Marshal(conn)
	resp, err := client.Post(srv.URL+"/api/v1/connections/", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("création connexion: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("création connexion: attendu 201, reçu %d", resp.StatusCode)
	}

	// Confirmer que l'env actif est dev et que le manager a bien la connexion
	if connManager.ActiveEnv != "dev" {
		t.Errorf("env initial: attendu 'dev', reçu %q", connManager.ActiveEnv)
	}

	// Switcher vers preprod via l'API
	switchBody, _ := json.Marshal(map[string]string{"env": "preprod"})
	reqSwitch, _ := http.NewRequest(http.MethodPut, srv.URL+"/api/v1/environment", bytes.NewReader(switchBody))
	reqSwitch.Header.Set("Content-Type", "application/json")
	respSwitch, err := client.Do(reqSwitch)
	if err != nil {
		t.Fatalf("PUT environment: %v", err)
	}
	respSwitch.Body.Close()

	// Vérifier que le manager reflète bien preprod (la résolution connexion suit)
	if connManager.ActiveEnv != "preprod" {
		t.Errorf("après switch: manager.ActiveEnv attendu 'preprod', reçu %q", connManager.ActiveEnv)
	}
}
