package resolver

import (
	"fmt"

	"github.com/rinjold/go-etl-studio/internal/connections"
	"github.com/rinjold/go-etl-studio/internal/connections/manager"
	"github.com/rinjold/go-etl-studio/internal/connections/secrets"
)

// ResolvedConn contient les paramètres de connexion résolus pour l'env actif.
type ResolvedConn struct {
	Type     string
	Host     string
	Port     int
	Database string
	User     string
	Password string
	DSN      string
}

// Resolve retourne les paramètres résolus d'une connexion pour l'environnement actif.
func Resolve(mgr *manager.Manager, connID string) (*ResolvedConn, error) {
	conn, err := mgr.Get(connID)
	if err != nil {
		return nil, err
	}
	return ResolveWithEnv(conn, mgr.ActiveEnv)
}

// ResolveWithEnv résout une connexion pour un environnement donné.
func ResolveWithEnv(conn *connections.Connection, env string) (*ResolvedConn, error) {
	envProfile, ok := conn.Envs[env]
	if !ok {
		return nil, fmt.Errorf("resolver: profil '%s' introuvable pour la connexion '%s'", env, conn.ID)
	}
	password, err := secrets.Resolve(envProfile.SecretRef)
	if err != nil {
		return nil, fmt.Errorf("resolver: %w", err)
	}
	rc := &ResolvedConn{
		Type:     conn.Type,
		Host:     envProfile.Host,
		Port:     envProfile.Port,
		Database: envProfile.Database,
		User:     envProfile.User,
		Password: password,
	}
	rc.DSN = envProfile.DSN(password)
	return rc, nil
}
