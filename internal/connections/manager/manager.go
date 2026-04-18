package manager

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/rinjold/go-etl-studio/internal/connections"
)

// Manager gère le CRUD des connexions et le switch d'environnement global.
type Manager struct {
	mu          sync.RWMutex
	connsDir    string
	connections map[string]*connections.Connection
	ActiveEnv   string // "dev" | "preprod" | "prod"
}

// New crée un Manager et charge les connexions existantes depuis le répertoire.
func New(connsDir string, activeEnv string) (*Manager, error) {
	if err := os.MkdirAll(connsDir, 0o755); err != nil {
		return nil, fmt.Errorf("manager: création répertoire connexions: %w", err)
	}
	m := &Manager{
		connsDir:    connsDir,
		connections: make(map[string]*connections.Connection),
		ActiveEnv:   activeEnv,
	}
	return m, m.loadAll()
}

// loadAll charge tous les fichiers XML du répertoire de connexions.
func (m *Manager) loadAll() error {
	entries, err := os.ReadDir(m.connsDir)
	if err != nil {
		return fmt.Errorf("manager: lecture répertoire connexions: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".xml" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(m.connsDir, e.Name()))
		if err != nil {
			return fmt.Errorf("manager: lecture connexion '%s': %w", e.Name(), err)
		}
		var conn connections.Connection
		if err := xml.Unmarshal(data, &conn); err != nil {
			return fmt.Errorf("manager: parse connexion '%s': %w", e.Name(), err)
		}
		conn.Envs = make(map[string]connections.ConnEnv, len(conn.EnvList))
		for _, env := range conn.EnvList {
			conn.Envs[env.Name] = env
		}
		m.connections[conn.ID] = &conn
	}
	return nil
}

// Get retourne une connexion par ID.
func (m *Manager) Get(id string) (*connections.Connection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, ok := m.connections[id]
	if !ok {
		return nil, fmt.Errorf("manager: connexion '%s' introuvable", id)
	}
	return conn, nil
}

// List retourne toutes les connexions.
func (m *Manager) List() []*connections.Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*connections.Connection, 0, len(m.connections))
	for _, c := range m.connections {
		list = append(list, c)
	}
	return list
}

// Save crée ou met à jour une connexion (persistance XML).
func (m *Manager) Save(conn *connections.Connection) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	conn.EnvList = make([]connections.ConnEnv, 0, len(conn.Envs))
	for _, env := range conn.Envs {
		conn.EnvList = append(conn.EnvList, env)
	}
	data, err := xml.MarshalIndent(conn, "", "  ")
	if err != nil {
		return fmt.Errorf("manager: sérialisation connexion '%s': %w", conn.ID, err)
	}
	path := filepath.Join(m.connsDir, conn.ID+".xml")
	if err := os.WriteFile(path, append([]byte(xml.Header), data...), 0o644); err != nil {
		return fmt.Errorf("manager: écriture connexion '%s': %w", conn.ID, err)
	}
	m.connections[conn.ID] = conn
	return nil
}

// Delete supprime une connexion.
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.connections, id)
	return os.Remove(filepath.Join(m.connsDir, id+".xml"))
}

// SwitchEnv bascule l'environnement actif globalement.
func (m *Manager) SwitchEnv(env string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveEnv = env
}
