package secrets_test

import (
	"testing"

	"github.com/rinjold/go-etl-studio/internal/connections/secrets"
)

// --- EnvProvider ---

func TestEnvProvider_Resolve_EnvVar(t *testing.T) {
	t.Setenv("TEST_PROV_SECRET", "mysecret")
	p := secrets.EnvProvider{}
	got, err := p.Resolve("${TEST_PROV_SECRET}")
	if err != nil {
		t.Fatalf("Resolve() erreur inattendue: %v", err)
	}
	if got != "mysecret" {
		t.Errorf("Resolve() = %q, attendu %q", got, "mysecret")
	}
}

func TestEnvProvider_Resolve_PlainText(t *testing.T) {
	p := secrets.EnvProvider{}
	got, err := p.Resolve("plaintextpass")
	if err != nil {
		t.Fatalf("Resolve() erreur inattendue: %v", err)
	}
	if got != "plaintextpass" {
		t.Errorf("Resolve() = %q, attendu %q", got, "plaintextpass")
	}
}

func TestEnvProvider_Resolve_MissingVar(t *testing.T) {
	p := secrets.EnvProvider{}
	_, err := p.Resolve("${UNDEFINED_PROV_VAR_XYZ}")
	if err == nil {
		t.Fatal("attendu une erreur pour variable non définie, got nil")
	}
}

// --- VaultProvider ---

func TestVaultProvider_Resolve_ReturnsError(t *testing.T) {
	p := secrets.VaultProvider{
		Address: "https://vault.example.com",
		Token:   "s.sometoken",
	}
	_, err := p.Resolve("vault:secret/data/myapp#password")
	if err == nil {
		t.Fatal("VaultProvider.Resolve() doit retourner une erreur (non implémenté)")
	}
}

// --- Interface compliance ---

// Vérifie à la compilation que EnvProvider et VaultProvider satisfont Provider.
var _ secrets.Provider = secrets.EnvProvider{}
var _ secrets.Provider = secrets.VaultProvider{}

// --- MockProvider pour valider l'interface ---

type mockProvider struct{ val string }

func (m mockProvider) Resolve(_ string) (string, error) { return m.val, nil }

var _ secrets.Provider = mockProvider{}

func TestMockProvider_ImplementsProvider(t *testing.T) {
	p := mockProvider{val: "injected"}
	got, err := p.Resolve("anything")
	if err != nil {
		t.Fatalf("mock Resolve() erreur: %v", err)
	}
	if got != "injected" {
		t.Errorf("mock Resolve() = %q, attendu %q", got, "injected")
	}
}
