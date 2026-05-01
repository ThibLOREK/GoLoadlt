package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/rinjold/go-etl-studio/internal/connections"
	"github.com/rinjold/go-etl-studio/internal/connections/manager"
	"github.com/rinjold/go-etl-studio/internal/connections/resolver"
)

// ---------------------------------------------------------------------------
// Tests unitaires de pingConnection
// ---------------------------------------------------------------------------

// TestPingConnection_PostgresUnreachable vérifie que postgres retourne une erreur
// si la cible n'est pas joignable (pas besoin d'infra réelle — timeout immédiat).
func TestPingConnection_PostgresUnreachable(t *testing.T) {
	rc := &resolver.ResolvedConn{
		Type:     "postgres",
		Host:     "127.0.0.1",
		Port:     9999, // port fermé
		Database: "testdb",
		User:     "user",
		Password: "pass",
		DSN:      "host=127.0.0.1 port=9999 dbname=testdb user=user password=pass sslmode=disable",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*1000*1000*1000) // 2s
	defer cancel()
	err := pingConnection(ctx, rc)
	if err == nil {
		t.Error("attendu une erreur pour postgres non joignable, got nil")
	}
}

// TestPingConnection_MySQL_Unsupported vérifie le message d'erreur explicite pour mysql.
func TestPingConnection_MySQL_Unsupported(t *testing.T) {
	rc := &resolver.ResolvedConn{Type: "mysql", Host: "localhost", Port: 3306, Database: "db", User: "u", Password: "p"}
	err := pingConnection(context.Background(), rc)
	if err == nil {
		t.Fatal("attendu une erreur pour mysql non supporté, got nil")
	}
	if !strings.Contains(err.Error(), "driver non disponible") {
		t.Errorf("message inattendu: %v", err)
	}
}

// TestPingConnection_MSSQL_Unsupported vérifie le message d'erreur explicite pour mssql.
func TestPingConnection_MSSQL_Unsupported(t *testing.T) {
	rc := &resolver.ResolvedConn{Type: "mssql", Host: "localhost", Port: 1433, Database: "db", User: "u", Password: "p"}
	err := pingConnection(context.Background(), rc)
	if err == nil {
		t.Fatal("attendu une erreur pour mssql non supporté, got nil")
	}
	if !strings.Contains(err.Error(), "driver non disponible") {
		t.Errorf("message inattendu: %v", err)
	}
}

// TestPingConnection_UnknownType vérifie l'erreur pour un type inconnu.
func TestPingConnection_UnknownType(t *testing.T) {
	rc := &resolver.ResolvedConn{Type: "oracle"}
	err := pingConnection(context.Background(), rc)
	if err == nil {
		t.Fatal("attendu une erreur pour type inconnu, got nil")
	}
	if !strings.Contains(err.Error(), "non supporté") {
		t.Errorf("message inattendu: %v", err)
	}
}

// TestPingConnection_REST_InvalidURL vérifie l'erreur pour une URL REST invalide.
func TestPingConnection_REST_InvalidURL(t *testing.T) {
	rc := &resolver.ResolvedConn{Type: "rest", Host: "not-a-url"}
	err := pingConnection(context.Background(), rc)
	if err == nil {
		t.Fatal("attendu une erreur pour URL REST invalide, got nil")
	}
}

// TestPingConnection_REST_EmptyHost vérifie l'erreur pour host REST vide.
func TestPingConnection_REST_EmptyHost(t *testing.T) {
	rc := &resolver.ResolvedConn{Type: "rest", Host: ""}
	err := pingConnection(context.Background(), rc)
	if err == nil {
		t.Fatal("attendu une erreur pour host REST vide, got nil")
	}
}

// TestPingConnection_REST_Success teste un vrai appel HTTP contre un serveur local.
func TestPingConnection_REST_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	rc := &resolver.ResolvedConn{Type: "rest", Host: ts.URL}
	ctx, cancel := context.WithTimeout(context.Background(), 3*1000*1000*1000)
	defer cancel()
	if err := pingConnection(ctx, rc); err != nil {
		t.Fatalf("pingConnection REST succès attendu, got: %v", err)
	}
}

// TestPingConnection_REST_Unreachable vérifie l'erreur quand le serveur REST est fermé.
func TestPingConnection_REST_Unreachable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := ts.URL
	ts.Close() // fermer immédiatement

	rc := &resolver.ResolvedConn{Type: "rest", Host: url}
	ctx, cancel := context.WithTimeout(context.Background(), 2*1000*1000*1000)
	defer cancel()
	if err := pingConnection(ctx, rc); err == nil {
		t.Error("attendu une erreur pour REST non joignable, got nil")
	}
}

// ---------------------------------------------------------------------------
// Tests d'intégration HTTP du handler Test
// ---------------------------------------------------------------------------

func newTestManagerWithConn(t *testing.T, conn *connections.Connection) *manager.Manager {
	t.Helper()
	dir := t.TempDir()
	m, err := manager.New(dir, "dev")
	if err != nil {
		t.Fatalf("manager.New: %v", err)
	}
	if err := m.Save(conn); err != nil {
		t.Fatalf("manager.Save: %v", err)
	}
	return m
}

func routerWithHandler(m *manager.Manager) http.Handler {
	r := chi.NewRouter()
	h := NewConnectionHandler(m, zerolog.Nop())
	r.Post("/api/v1/connections/{connID}/test", h.Test)
	r.Get("/api/v1/connections", h.List)
	r.Put("/api/v1/environment", h.SwitchEnv)
	r.Get("/api/v1/environment", h.GetEnv)
	return r
}

// TestHandlerTest_ConnNotFound vérifie que le handler retourne 400 pour un ID inconnu.
func TestHandlerTest_ConnNotFound(t *testing.T) {
	dir := t.TempDir()
	m, _ := manager.New(dir, "dev")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/connections/unknown-id/test", nil)
	routerWithHandler(m).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, attendu 400", rr.Code)
	}
}

// TestHandlerTest_ProfileMissing vérifie 400 quand le profil env est absent.
func TestHandlerTest_ProfileMissing(t *testing.T) {
	conn := &connections.Connection{
		ID: "no-prod", Name: "test", Type: "postgres",
		Envs: map[string]connections.ConnEnv{
			"dev": {Name: "dev", Host: "localhost", Port: 5432, Database: "db", User: "u", SecretRef: "p"},
		},
	}
	m := newTestManagerWithConn(t, conn)
	if err := m.SwitchEnv("prod"); err != nil {
		t.Fatalf("SwitchEnv: %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/connections/no-prod/test", nil)
	routerWithHandler(m).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, attendu 400", rr.Code)
	}
}

// TestHandlerTest_MySQL_Returns502 vérifie que mysql retourne 502 avec message explicite.
func TestHandlerTest_MySQL_Returns502(t *testing.T) {
	conn := &connections.Connection{
		ID: "mysql-conn", Name: "mysql", Type: "mysql",
		Envs: map[string]connections.ConnEnv{
			"dev": {Name: "dev", Host: "localhost", Port: 3306, Database: "db", User: "u", SecretRef: "p"},
		},
	}
	m := newTestManagerWithConn(t, conn)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/connections/mysql-conn/test", nil)
	routerWithHandler(m).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Errorf("status = %d, attendu 502", rr.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !strings.Contains(body["error"], "driver non disponible") {
		t.Errorf("message inattendu: %q", body["error"])
	}
}

// TestHandlerTest_REST_Success vérifie que rest retourne 200 sur un serveur local joignable.
func TestHandlerTest_REST_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	conn := &connections.Connection{
		ID: "rest-conn", Name: "rest", Type: "rest",
		Envs: map[string]connections.ConnEnv{
			"dev": {Name: "dev", Host: ts.URL, SecretRef: ""},
		},
	}
	m := newTestManagerWithConn(t, conn)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/connections/rest-conn/test"), nil)
	routerWithHandler(m).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, attendu 200, body: %s", rr.Code, rr.Body.String())
	}
	var body map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %q, attendu \"ok\"", body["status"])
	}
	if body["env"] != "dev" {
		t.Errorf("env = %q, attendu \"dev\"", body["env"])
	}
}

// TestHandlerSwitchEnv_InvalidEnv vérifie que SwitchEnv rejette les valeurs invalides.
func TestHandlerSwitchEnv_InvalidEnv(t *testing.T) {
	dir := t.TempDir()
	m, _ := manager.New(dir, "dev")

	rr := httptest.NewRecorder()
	body := strings.NewReader(`{"env":"staging"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/environment", body)
	routerWithHandler(m).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, attendu 400", rr.Code)
	}
}

// TestHandlerSwitchEnv_ValidEnv vérifie que SwitchEnv accepte et persiste les valeurs valides.
func TestHandlerSwitchEnv_ValidEnv(t *testing.T) {
	dir := t.TempDir()
	m, _ := manager.New(dir, "dev")

	rr := httptest.NewRecorder()
	body := strings.NewReader(`{"env":"prod"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/environment", body)
	routerWithHandler(m).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, attendu 200", rr.Code)
	}
	if m.ActiveEnv != "prod" {
		t.Errorf("ActiveEnv = %q, attendu \"prod\"", m.ActiveEnv)
	}
}
