package manager

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// newTestManager crée un Manager dans un répertoire temporaire sans connexions XML.
func newTestManager(t *testing.T, defaultEnv string) *Manager {
	t.Helper()
	dir := t.TempDir()
	m, err := New(dir, defaultEnv)
	if err != nil {
		t.Fatalf("New() erreur inattendue: %v", err)
	}
	return m
}

// TestNew_DefaultEnv vérifie que le manager démarre avec la valeur par défaut
// quand aucun fichier de persistance n'existe encore.
func TestNew_DefaultEnv(t *testing.T) {
	m := newTestManager(t, "dev")
	if m.ActiveEnv != "dev" {
		t.Errorf("ActiveEnv = %q, attendu %q", m.ActiveEnv, "dev")
	}
}

// TestSwitchEnv_PersistsOnDisk vérifie que SwitchEnv écrit bien le fichier .env-state.json.
func TestSwitchEnv_PersistsOnDisk(t *testing.T) {
	m := newTestManager(t, "dev")

	if err := m.SwitchEnv("prod"); err != nil {
		t.Fatalf("SwitchEnv() erreur inattendue: %v", err)
	}
	if m.ActiveEnv != "prod" {
		t.Errorf("ActiveEnv en mémoire = %q, attendu %q", m.ActiveEnv, "prod")
	}

	// Vérifie le contenu du fichier persisté
	data, err := os.ReadFile(filepath.Join(m.connsDir, envStateFile))
	if err != nil {
		t.Fatalf("fichier %s absent après SwitchEnv: %v", envStateFile, err)
	}
	var s struct {
		ActiveEnv string `json:"activeEnv"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("impossible de parser %s: %v", envStateFile, err)
	}
	if s.ActiveEnv != "prod" {
		t.Errorf("activeEnv persisté = %q, attendu %q", s.ActiveEnv, "prod")
	}
}

// TestNew_LoadsPersistedEnv simule un redémarrage : crée un manager, switch l'env,
// recrée un manager dans le même répertoire et vérifie que l'env est restauré.
func TestNew_LoadsPersistedEnv(t *testing.T) {
	dir := t.TempDir()

	// Première exécution : switch vers "preprod"
	m1, err := New(dir, "dev")
	if err != nil {
		t.Fatalf("New() (1ère instance) erreur: %v", err)
	}
	if err := m1.SwitchEnv("preprod"); err != nil {
		t.Fatalf("SwitchEnv() erreur: %v", err)
	}

	// Redémarrage simulé : nouveau Manager dans le même répertoire
	m2, err := New(dir, "dev") // default "dev" doit être écrasé par la valeur persistée
	if err != nil {
		t.Fatalf("New() (2ème instance) erreur: %v", err)
	}
	if m2.ActiveEnv != "preprod" {
		t.Errorf("après redémarrage, ActiveEnv = %q, attendu %q", m2.ActiveEnv, "preprod")
	}
}

// TestNew_NoStateFile vérifie le comportement quand le fichier de persistance est absent.
func TestNew_NoStateFile(t *testing.T) {
	dir := t.TempDir()
	// Aucun fichier .env-state.json — doit utiliser le défaut sans erreur
	m, err := New(dir, "preprod")
	if err != nil {
		t.Fatalf("New() erreur inattendue sans fichier d'état: %v", err)
	}
	if m.ActiveEnv != "preprod" {
		t.Errorf("ActiveEnv = %q, attendu %q (défaut)", m.ActiveEnv, "preprod")
	}
}

// TestNew_InvalidStateFile vérifie que le manager démarre proprement (valeur par défaut)
// si le fichier .env-state.json est corrompu ou invalide.
func TestNew_InvalidStateFile(t *testing.T) {
	dir := t.TempDir()
	// Écriture d'un fichier invalide
	if err := os.WriteFile(filepath.Join(dir, envStateFile), []byte("{invalid json}"), 0o644); err != nil {
		t.Fatalf("écriture fichier corrompu: %v", err)
	}

	m, err := New(dir, "dev")
	if err != nil {
		t.Fatalf("New() ne doit pas échouer sur un fichier corrompu: %v", err)
	}
	// Doit conserver la valeur par défaut
	if m.ActiveEnv != "dev" {
		t.Errorf("ActiveEnv = %q, attendu %q (défaut car fichier corrompu)", m.ActiveEnv, "dev")
	}
}

// TestNew_EmptyActiveEnvInStateFile vérifie qu'un activeEnv vide dans le fichier
// est ignoré et que la valeur par défaut est conservée.
func TestNew_EmptyActiveEnvInStateFile(t *testing.T) {
	dir := t.TempDir()
	data, _ := json.Marshal(map[string]string{"activeEnv": ""})
	if err := os.WriteFile(filepath.Join(dir, envStateFile), data, 0o644); err != nil {
		t.Fatalf("écriture fichier état vide: %v", err)
	}

	m, err := New(dir, "prod")
	if err != nil {
		t.Fatalf("New() erreur inattendue: %v", err)
	}
	if m.ActiveEnv != "prod" {
		t.Errorf("ActiveEnv = %q, attendu %q (activeEnv vide ignoré)", m.ActiveEnv, "prod")
	}
}

// TestSwitchEnv_MultipleSwitch vérifie que plusieurs switch successifs
// persistent correctement le dernier état.
func TestSwitchEnv_MultipleSwitch(t *testing.T) {
	dir := t.TempDir()

	m, err := New(dir, "dev")
	if err != nil {
		t.Fatalf("New() erreur: %v", err)
	}

	for _, env := range []string{"preprod", "prod", "dev", "prod"} {
		if err := m.SwitchEnv(env); err != nil {
			t.Fatalf("SwitchEnv(%q) erreur: %v", env, err)
		}
	}

	// Redémarrage : doit restaurer "prod" (le dernier)
	m2, err := New(dir, "dev")
	if err != nil {
		t.Fatalf("New() (redémarrage) erreur: %v", err)
	}
	if m2.ActiveEnv != "prod" {
		t.Errorf("après redémarrage, ActiveEnv = %q, attendu %q", m2.ActiveEnv, "prod")
	}
}
