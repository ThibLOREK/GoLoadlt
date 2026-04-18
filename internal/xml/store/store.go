package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
	"github.com/rinjold/go-etl-studio/internal/xml/parser"
	"github.com/rinjold/go-etl-studio/internal/xml/serializer"
)

// Store gère la persistance des fichiers XML de projets sur disque.
// Structure :
//   projects/{projectID}/project.xml      → version courante
//   projects/{projectID}/history/v{n}.xml → versions archivées
type Store struct {
	BaseDir string // répertoire racine (ex: "./projects")
}

// New crée un Store avec le répertoire de base donné.
func New(baseDir string) *Store {
	return &Store{BaseDir: baseDir}
}

// projectDir retourne le chemin du répertoire d'un projet.
func (s *Store) projectDir(id string) string {
	return filepath.Join(s.BaseDir, id)
}

// currentPath retourne le chemin du fichier XML courant d'un projet.
func (s *Store) currentPath(id string) string {
	return filepath.Join(s.projectDir(id), "project.xml")
}

// historyDir retourne le chemin du répertoire d'historique d'un projet.
func (s *Store) historyDir(id string) string {
	return filepath.Join(s.projectDir(id), "history")
}

// Save sérialise et sauvegarde un projet.
// L'ancienne version est archivée dans history/ avant écrasement.
// Retourne le SHA256 du fichier écrit.
func (s *Store) Save(p *contracts.Project) (sha string, err error) {
	data, err := serializer.Serialize(p)
	if err != nil {
		return "", fmt.Errorf("store.Save: %w", err)
	}

	// Créer les répertoires si nécessaire.
	if err := os.MkdirAll(s.historyDir(p.ID), 0o755); err != nil {
		return "", fmt.Errorf("store.Save: création répertoires: %w", err)
	}

	current := s.currentPath(p.ID)

	// Archiver la version courante si elle existe.
	if existing, err := os.ReadFile(current); err == nil {
		archiveName := fmt.Sprintf("v%d_%s.xml", p.Version, time.Now().Format("20060102_150405"))
		archivePath := filepath.Join(s.historyDir(p.ID), archiveName)
		if err := os.WriteFile(archivePath, existing, 0o644); err != nil {
			return "", fmt.Errorf("store.Save: archivage: %w", err)
		}
	}

	// Écrire la nouvelle version.
	if err := os.WriteFile(current, data, 0o644); err != nil {
		return "", fmt.Errorf("store.Save: écriture: %w", err)
	}

	// Calculer et retourner le SHA256.
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// Load charge et parse le XML courant d'un projet.
func (s *Store) Load(id string) (*contracts.Project, error) {
	path := s.currentPath(id)
	p, err := parser.ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("store.Load '%s': %w", id, err)
	}
	return p, nil
}

// LoadVersion charge une version archivée (numéro de version).
func (s *Store) LoadVersion(id string, version int) (*contracts.Project, error) {
	pattern := filepath.Join(s.historyDir(id), "v"+strconv.Itoa(version)+"_*.xml")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return nil, fmt.Errorf("store.LoadVersion: version %d introuvable pour projet '%s'", version, id)
	}
	return parser.ParseFile(matches[len(matches)-1])
}

// Delete supprime un projet et tout son historique.
func (s *Store) Delete(id string) error {
	if err := os.RemoveAll(s.projectDir(id)); err != nil {
		return fmt.Errorf("store.Delete '%s': %w", id, err)
	}
	return nil
}

// Exists vérifie si un projet existe sur disque.
func (s *Store) Exists(id string) bool {
	_, err := os.Stat(s.currentPath(id))
	return err == nil
}

// ListIDs retourne les IDs de tous les projets présents sur disque.
func (s *Store) ListIDs() ([]string, error) {
	entries, err := os.ReadDir(s.BaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("store.ListIDs: %w", err)
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			ids = append(ids, e.Name())
		}
	}
	return ids, nil
}
