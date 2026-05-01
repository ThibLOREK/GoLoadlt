package resolver_test

import (
	"os"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/connections"
	"github.com/rinjold/go-etl-studio/internal/connections/manager"
	"github.com/rinjold/go-etl-studio/internal/connections/resolver"
)

// newManagerWithConn crée un Manager temporaire avec une connexion pré-chargée en mémoire.
func newManagerWithConn(t *testing.T, conn *connections.Connection) *manager.Manager {
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

func makeConn(id, connType string) *connections.Connection {
	return &connections.Connection{
		ID:   id,
		Name: "test-" + id,
		Type: connType,
		Envs: map[string]connections.ConnEnv{
			"dev": {
				Name:      "dev",
				Host:      "localhost",
				Port:      5432,
				Database:  "testdb",
				User:      "testuser",
				SecretRef: "plaintext", // valeur brute autorisée en dev
			},
		},
	}
}

// TestResolveWithEnv_Success vérifie qu'un profil valide est résolu correctement.
func TestResolveWithEnv_Success(t *testing.T) {
	conn := makeConn("conn-1", "postgres")
	rc, err := resolver.ResolveWithEnv(conn, "dev")
	if err != nil {
		t.Fatalf("ResolveWithEnv() erreur inattendue: %v", err)
	}
	if rc.Type != "postgres" {
		t.Errorf("Type = %q, attendu \"postgres\"", rc.Type)
	}
	if rc.Host != "localhost" {
		t.Errorf("Host = %q, attendu \"localhost\"", rc.Host)
	}
	if rc.Database != "testdb" {
		t.Errorf("Database = %q, attendu \"testdb\"", rc.Database)
	}
	if rc.Password != "plaintext" {
		t.Errorf("Password = %q, attendu \"plaintext\"", rc.Password)
	}
	if rc.DSN == "" {
		t.Error("DSN vide — attendu non-vide")
	}
}

// TestResolveWithEnv_MissingEnvProfile vérifie l'erreur quand le profil d'env est absent.
func TestResolveWithEnv_MissingEnvProfile(t *testing.T) {
	conn := makeConn("conn-2", "postgres")
	_, err := resolver.ResolveWithEnv(conn, "prod") // "prod" absent
	if err == nil {
		t.Fatal("attendu une erreur pour profil 'prod' absent, got nil")
	}
}

// TestResolveWithEnv_SecretEnvVarMissing vérifie l'erreur si la variable d'env est absente.
func TestResolveWithEnv_SecretEnvVarMissing(t *testing.T) {
	conn := &connections.Connection{
		ID:   "conn-3",
		Name: "test",
		Type: "postgres",
		Envs: map[string]connections.ConnEnv{
			"dev": {
				Name:      "dev",
				Host:      "localhost",
				Port:      5432,
				Database:  "db",
				User:      "u",
				SecretRef: "${UNDEFINED_SECRET_XYZ_12345}",
			},
		},
	}
	// S'assurer que la variable n'est pas définie
	os.Unsetenv("UNDEFINED_SECRET_XYZ_12345")
	_, err := resolver.ResolveWithEnv(conn, "dev")
	if err == nil {
		t.Fatal("attendu une erreur pour secret non défini, got nil")
	}
}

// TestResolve_ConnNotFound vérifie que Resolve() retourne une erreur si l'ID est absent.
func TestResolve_ConnNotFound(t *testing.T) {
	dir := t.TempDir()
	m, err := manager.New(dir, "dev")
	if err != nil {
		t.Fatalf("manager.New: %v", err)
	}
	_, err = resolver.Resolve(m, "inexistant")
	if err == nil {
		t.Fatal("attendu une erreur pour connexion inexistante, got nil")
	}
}

// TestResolve_ActiveEnv vérifie que Resolve() utilise bien l'environnement actif du manager.
func TestResolve_ActiveEnv(t *testing.T) {
	conn := &connections.Connection{
		ID:   "conn-env",
		Name: "test-env",
		Type: "postgres",
		Envs: map[string]connections.ConnEnv{
			"dev": {
				Name: "dev", Host: "dev-host", Port: 5432,
				Database: "devdb", User: "devuser", SecretRef: "devpass",
			},
			"prod": {
				Name: "prod", Host: "prod-host", Port: 5432,
				Database: "proddb", User: "produser", SecretRef: "prodpass",
			},
		},
	}
	m := newManagerWithConn(t, conn)

	// Env actif = dev (défaut)
	rc, err := resolver.Resolve(m, "conn-env")
	if err != nil {
		t.Fatalf("Resolve() erreur inattendue: %v", err)
	}
	if rc.Host != "dev-host" {
		t.Errorf("Host = %q, attendu \"dev-host\" pour env dev", rc.Host)
	}

	// Switch vers prod
	if err := m.SwitchEnv("prod"); err != nil {
		t.Fatalf("SwitchEnv: %v", err)
	}
	rc, err = resolver.Resolve(m, "conn-env")
	if err != nil {
		t.Fatalf("Resolve() après switch prod: %v", err)
	}
	if rc.Host != "prod-host" {
		t.Errorf("Host = %q, attendu \"prod-host\" pour env prod", rc.Host)
	}
}
