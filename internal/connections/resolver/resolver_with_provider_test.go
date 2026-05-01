package resolver_test

import (
	"errors"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/connections"
	"github.com/rinjold/go-etl-studio/internal/connections/resolver"
)

// mockProvider est un Provider injecté en test pour contrôler la résolution du secret.
type mockProvider struct {
	val string
	err error
}

func (m mockProvider) Resolve(_ string) (string, error) { return m.val, m.err }

func testConn(id, connType string) *connections.Connection {
	return &connections.Connection{
		ID:   id,
		Name: "test-" + id,
		Type: connType,
		Envs: map[string]connections.ConnEnv{
			"dev": {
				Name:      "dev",
				Host:      "db.example.com",
				Port:      5432,
				Database:  "mydb",
				User:      "alice",
				SecretRef: "${DB_PASS}",
			},
		},
	}
}

// TestResolveWithProvider_Success vérifie la résolution nominale avec un provider mocké.
func TestResolveWithProvider_Success(t *testing.T) {
	conn := testConn("c1", "postgres")
	rc, err := resolver.ResolveWithProvider(conn, "dev", mockProvider{val: "s3cr3t"})
	if err != nil {
		t.Fatalf("ResolveWithProvider() erreur inattendue: %v", err)
	}
	if rc.Password != "s3cr3t" {
		t.Errorf("Password = %q, attendu %q", rc.Password, "s3cr3t")
	}
	if rc.Host != "db.example.com" {
		t.Errorf("Host = %q, attendu %q", rc.Host, "db.example.com")
	}
	if rc.DSN == "" {
		t.Error("DSN vide — attendu non-vide")
	}
}

// TestResolveWithProvider_ProviderError vérifie que l'erreur du provider est propagée.
func TestResolveWithProvider_ProviderError(t *testing.T) {
	conn := testConn("c2", "postgres")
	providerErr := errors.New("vault: connexion refusée")
	_, err := resolver.ResolveWithProvider(conn, "dev", mockProvider{err: providerErr})
	if err == nil {
		t.Fatal("attendu une erreur du provider, got nil")
	}
	if !errors.Is(err, providerErr) {
		t.Errorf("erreur = %v, attendu wrapping de %v", err, providerErr)
	}
}

// TestResolveWithProvider_MissingProfile vérifie l'erreur si le profil d'env est absent.
func TestResolveWithProvider_MissingProfile(t *testing.T) {
	conn := testConn("c3", "postgres")
	_, err := resolver.ResolveWithProvider(conn, "prod", mockProvider{val: "x"})
	if err == nil {
		t.Fatal("attendu une erreur pour profil 'prod' absent, got nil")
	}
}

// TestResolveWithEnv_BackwardCompat vérifie que ResolveWithEnv (ancienne API) fonctionne toujours.
func TestResolveWithEnv_BackwardCompat(t *testing.T) {
	conn := &connections.Connection{
		ID:   "compat",
		Name: "compat",
		Type: "postgres",
		Envs: map[string]connections.ConnEnv{
			"dev": {
				Name: "dev", Host: "h", Port: 5432,
				Database: "d", User: "u", SecretRef: "plainpass",
			},
		},
	}
	// ResolveWithEnv doit continuer de fonctionner sans provider explicite
	rc, err := resolver.ResolveWithEnv(conn, "dev")
	if err != nil {
		t.Fatalf("ResolveWithEnv() erreur inattendue: %v", err)
	}
	if rc.Password != "plainpass" {
		t.Errorf("Password = %q, attendu %q", rc.Password, "plainpass")
	}
}
