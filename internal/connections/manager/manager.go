package manager

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

// ConnEnv contient les paramètres d'un environnement de connexion.
type ConnEnv struct {
	Name      string `xml:"name,attr"`
	Host      string `xml:"host,attr"`
	Port      int    `xml:"port,attr"`
	Database  string `xml:"db,attr"`
	User      string `xml:"user,attr"`
	SecretRef string `xml:"secretRef,attr"` // ex: "${DB_PASSWORD}" ou "vault:secret/crm"
}

// Connection est une connexion réutilisable multi-environnements.
type Connection struct {
	ID   string     `xml:"id,attr"`
	Name string     `xml:"name,attr"`
	Type string     `xml:"type,attr"` // postgres, mysql, mssql, rest, csv...
	Envs []ConnEnv  `xml:"environments>env"`
}

// EnvMap retourne les environnements indexés par nom.
func (c *Connection) EnvMap() map[string]*ConnEnv {
	m := make(map[string]*ConnEnv, len(c.Envs))
	for i := range c.Envs {
		m[c.Envs[i].Name] = &c.Envs[i]
	}
	return m
}

// Manager gère les connexions persistantes en XML.
type Manager struct {
	BaseDir string // ex: "./connections"
}

// New crée un Manager.
func New(baseDir string) *Manager {
	return &Manager{BaseDir: baseDir}
}

func (m *Manager) path(id string) string {
	return filepath.Join(m.BaseDir, id+".xml")
}

// Save sauvegarde une connexion en XML.
func (m *Manager) Save(c *Connection) error {
	if err := os.MkdirAll(m.BaseDir, 0o755); err != nil {
		return fmt.Errorf("connections.Save: %w", err)
	}
	data, err := xml.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("connections.Save: marshal: %w", err)
	}
	data = append([]byte(xml.Header), data...)
	return os.WriteFile(m.path(c.ID), data, 0o644)
}

// Load charge une connexion par ID.
func (m *Manager) Load(id string) (*Connection, error) {
	data, err := os.ReadFile(m.path(id))
	if err != nil {
		return nil, fmt.Errorf("connections.Load '%s': %w", id, err)
	}
	var c Connection
	if err := xml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("connections.Load '%s': unmarshal: %w", id, err)
	}
	return &c, nil
}

// Delete supprime une connexion.
func (m *Manager) Delete(id string) error {
	if err := os.Remove(m.path(id)); err != nil {
		return fmt.Errorf("connections.Delete '%s': %w", id, err)
	}
	return nil
}

// List retourne toutes les connexions.
func (m *Manager) List() ([]*Connection, error) {
	entries, err := os.ReadDir(m.BaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("connections.List: %w", err)
	}
	var conns []*Connection
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".xml" {
			continue
		}
		id := e.Name()[:len(e.Name())-4]
		c, err := m.Load(id)
		if err != nil {
			continue
		}
		conns = append(conns, c)
	}
	return conns, nil
}
