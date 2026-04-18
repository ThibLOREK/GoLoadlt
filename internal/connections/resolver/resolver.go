package resolver

import (
	"fmt"

	"github.com/rinjold/go-etl-studio/internal/connections/manager"
	"github.com/rinjold/go-etl-studio/internal/connections/secrets"
)

// ResolvedConn contient les paramètres résolus d'une connexion pour un environnement donné.
type ResolvedConn struct {
	ID       string
	Name     string
	Type     string
	Host     string
	Port     int
	Database string
	User     string
	Password string // résolu depuis SecretRef (jamais en clair dans le XML)
}

// DSN retourne la chaîne de connexion PostgreSQL.
func (r *ResolvedConn) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		r.Host, r.Port, r.Database, r.User, r.Password,
	)
}

// Resolver résout une connexion pour l'environnement actif.
type Resolver struct {
	Manager   *manager.Manager
	ActiveEnv string // "dev", "preprod", "prod"
}

// New crée un Resolver.
func New(mgr *manager.Manager, activeEnv string) *Resolver {
	return &Resolver{Manager: mgr, ActiveEnv: activeEnv}
}

// Resolve retourne les paramètres de connexion résolus pour l'environnement actif.
func (r *Resolver) Resolve(connID string) (*ResolvedConn, error) {
	conn, err := r.Manager.Load(connID)
	if err != nil {
		return nil, fmt.Errorf("resolver: connexion '%s' introuvable: %w", connID, err)
	}

	envMap := conn.EnvMap()
	env, ok := envMap[r.ActiveEnv]
	if !ok {
		return nil, fmt.Errorf("resolver: connexion '%s' n'a pas de profil pour l'env '%s'", connID, r.ActiveEnv)
	}

	password, err := secrets.Resolve(env.SecretRef)
	if err != nil {
		return nil, fmt.Errorf("resolver: secret '%s': %w", env.SecretRef, err)
	}

	return &ResolvedConn{
		ID:       conn.ID,
		Name:     conn.Name,
		Type:     conn.Type,
		Host:     env.Host,
		Port:     env.Port,
		Database: env.Database,
		User:     env.User,
		Password: password,
	}, nil
}
