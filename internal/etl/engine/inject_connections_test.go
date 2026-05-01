package engine

import (
	"errors"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/connections/resolver"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

// mockResolve retourne un ResolvedConn fixe ou une erreur selon connID.
func mockResolve(connID string) (*resolver.ResolvedConn, error) {
	switch connID {
	case "conn-postgres":
		return &resolver.ResolvedConn{
			Type:     "postgres",
			Host:     "db.example.com",
			Port:     5432,
			Database: "mydb",
			User:     "app",
			Password: "s3cr3t",
			DSN:      "host=db.example.com port=5432 dbname=mydb user=app password=s3cr3t sslmode=disable",
		}, nil
	case "conn-broken":
		return nil, errors.New("resolver: profil 'dev' introuvable pour la connexion 'conn-broken'")
	default:
		return nil, errors.New("manager: connexion '" + connID + "' introuvable")
	}
}

// TestInjectResolvedConnections_Nominal vérifie que les params dsn/db_type/host/database
// sont correctement injectés dans le node qui porte un ConnRef.
func TestInjectResolvedConnections_Nominal(t *testing.T) {
	project := &contracts.Project{
		ID: "proj-1",
		Nodes: []contracts.Node{
			{
				ID:      "src-1",
				Type:    "source.postgres",
				ConnRef: "conn-postgres",
				Params: []contracts.Param{
					{Name: "query", Value: "SELECT 1"},
				},
			},
		},
	}

	if err := InjectResolvedConnections(project, mockResolve); err != nil {
		t.Fatalf("injection inattendue d'erreur: %v", err)
	}

	params := project.Nodes[0].ParamMap()

	cases := []struct{ key, want string }{
		{"dsn", "host=db.example.com port=5432 dbname=mydb user=app password=s3cr3t sslmode=disable"},
		{"db_type", "postgres"},
		{"host", "db.example.com"},
		{"database", "mydb"},
		{"query", "SELECT 1"}, // param préexistant conservé
	}
	for _, c := range cases {
		if got := params[c.key]; got != c.want {
			t.Errorf("param[%q] = %q, want %q", c.key, got, c.want)
		}
	}
}

// TestInjectResolvedConnections_NoConnRef vérifie que les nodes sans ConnRef
// sont ignorés sans erreur.
func TestInjectResolvedConnections_NoConnRef(t *testing.T) {
	project := &contracts.Project{
		ID: "proj-2",
		Nodes: []contracts.Node{
			{ID: "transform-1", Type: "transform.filter", ConnRef: ""},
		},
	}

	if err := InjectResolvedConnections(project, mockResolve); err != nil {
		t.Fatalf("erreur inattendue sur node sans ConnRef: %v", err)
	}

	if len(project.Nodes[0].Params) != 0 {
		t.Errorf("aucun param ne doit être injecté sur un node sans ConnRef, got %v", project.Nodes[0].Params)
	}
}

// TestInjectResolvedConnections_ResolverError vérifie que l'erreur du resolver
// est remontée correctement avec le contexte du node.
func TestInjectResolvedConnections_ResolverError(t *testing.T) {
	project := &contracts.Project{
		ID: "proj-3",
		Nodes: []contracts.Node{
			{ID: "src-broken", Type: "source.postgres", ConnRef: "conn-broken"},
		},
	}

	err := InjectResolvedConnections(project, mockResolve)
	if err == nil {
		t.Fatal("une erreur était attendue, got nil")
	}
	const wantSubstr = "src-broken"
	if !contains(err.Error(), wantSubstr) {
		t.Errorf("erreur %q ne contient pas %q", err.Error(), wantSubstr)
	}
}

// TestInjectResolvedConnections_MultiNode vérifie que plusieurs nodes avec ConnRef
// différents sont tous résolus, et que l'erreur arrête le traitement au premier
// node en échec.
func TestInjectResolvedConnections_MultiNode(t *testing.T) {
	project := &contracts.Project{
		ID: "proj-4",
		Nodes: []contracts.Node{
			{ID: "src-1", Type: "source.postgres", ConnRef: "conn-postgres"},
			{ID: "src-2", Type: "source.postgres", ConnRef: "conn-broken"},
			{ID: "src-3", Type: "source.postgres", ConnRef: "conn-postgres"},
		},
	}

	err := InjectResolvedConnections(project, mockResolve)
	if err == nil {
		t.Fatal("une erreur était attendue sur src-2")
	}

	// src-1 doit avoir été injecté avant l'erreur
	params1 := project.Nodes[0].ParamMap()
	if params1["db_type"] != "postgres" {
		t.Errorf("src-1 aurait dû être injecté avant l'erreur")
	}

	// src-3 ne doit PAS avoir été injecté (arrêt au premier échec)
	params3 := project.Nodes[2].ParamMap()
	if params3["db_type"] != "" {
		t.Errorf("src-3 ne devrait pas avoir été injecté après l'erreur")
	}
}

// TestInjectResolvedConnections_EnsureParamUpdates vérifie que si un param
// existe déjà (ex: dsn hard-codé), il est bien écrasé par la valeur résolue.
func TestInjectResolvedConnections_EnsureParamUpdates(t *testing.T) {
	project := &contracts.Project{
		ID: "proj-5",
		Nodes: []contracts.Node{
			{
				ID:      "src-1",
				Type:    "source.postgres",
				ConnRef: "conn-postgres",
				Params: []contracts.Param{
					{Name: "dsn", Value: "host=old-value"},
					{Name: "host", Value: "old-host"},
				},
			},
		},
	}

	if err := InjectResolvedConnections(project, mockResolve); err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}

	params := project.Nodes[0].ParamMap()
	if params["dsn"] == "host=old-value" {
		t.Error("le param dsn aurait dû être écrasé par la valeur résolue")
	}
	if params["host"] == "old-host" {
		t.Error("le param host aurait dû être écrasé par la valeur résolue")
	}
	if params["host"] != "db.example.com" {
		t.Errorf("host = %q, want %q", params["host"], "db.example.com")
	}
}

// contains est un helper minimal pour éviter d'importer strings dans les tests.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
