package manager

import (
	"encoding/json"
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

// envStateFile est le nom du fichier de persistance de l'environnement actif.
const envStateFile = ".env-state.json"

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
	if err := m.loadAll(); err != nil {
		return nil, err
	}
	// Restaure l'env actif persisté sur disque (survit aux redémarrages)
	m.loadEnvState()
	return m, nil
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
		// Hydrate la map Envs depuis EnvList (XML → mémoire)
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
	// Aplatit la map Envs vers EnvList pour la sérialisation XML
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

// SwitchEnv bascule l'environnement actif globalement et persiste le choix sur disque.
func (m *Manager) SwitchEnv(env string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveEnv = env
	return m.persistEnvState()
}

// persistEnvState écrit l'env actif dans .env-state.json dans le répertoire des connexions.
func (m *Manager) persistEnvState() error {
	type state struct {
		ActiveEnv string `json:"activeEnv"`
	}
	data, err := json.Marshal(state{ActiveEnv: m.ActiveEnv})
	if err != nil {
		return fmt.Errorf("manager: marshal env state: %w", err)
	}
	path := filepath.Join(m.connsDir, envStateFile)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("manager: écriture env state: %w", err)
	}
	return nil
}

// loadEnvState restaure l'env actif depuis .env-state.json si présent.
func (m *Manager) loadEnvState() {
	type state struct {
		ActiveEnv string `json:"activeEnv"`
	}
	data, err := os.ReadFile(filepath.Join(m.connsDir, envStateFile))
	if err != nil {
		return // fichier absent = première exécution, on garde la valeur par défaut
	}
	var s state
	if json.Unmarshal(data, &s) == nil && s.ActiveEnv != "" {
		m.ActiveEnv = s.ActiveEnv
	}
}
