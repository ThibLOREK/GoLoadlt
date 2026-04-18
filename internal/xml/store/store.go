package store

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
	"github.com/rinjold/go-etl-studio/internal/xml/parser"
	"github.com/rinjold/go-etl-studio/internal/xml/serializer"
)

// ProjectStore gère la persistance des fichiers XML de projets.
// Structure : projectsDir/{projectID}/project.xml
//             projectsDir/{projectID}/history/v{n}.xml
type ProjectStore struct {
	projectsDir string
}

// NewProjectStore crée un ProjectStore. Le répertoire est créé si nécessaire.
func NewProjectStore(projectsDir string) (*ProjectStore, error) {
	if err := os.MkdirAll(projectsDir, 0o755); err != nil {
		return nil, fmt.Errorf("store: création répertoire projets: %w", err)
	}
	return &ProjectStore{projectsDir: projectsDir}, nil
}

// Save sauvegarde un projet : archive la version précédente dans history/, écrit la nouvelle.
func (s *ProjectStore) Save(p *contracts.Project) error {
	dir := filepath.Join(s.projectsDir, p.ID)
	if err := os.MkdirAll(filepath.Join(dir, "history"), 0o755); err != nil {
		return fmt.Errorf("store: création répertoire projet '%s': %w", p.ID, err)
	}

	projectFile := filepath.Join(dir, "project.xml")

	// Archiver la version précédente si elle existe.
	if _, err := os.Stat(projectFile); err == nil {
		archivePath := filepath.Join(dir, "history", fmt.Sprintf("v%d.xml", p.Version-1))
		data, _ := os.ReadFile(projectFile)
		_ = os.WriteFile(archivePath, data, 0o644)
	}

	// Sérialiser et écrire le nouveau fichier.
	data, err := serializer.SerializeProject(p)
	if err != nil {
		return err
	}
	if err := os.WriteFile(projectFile, data, 0o644); err != nil {
		return fmt.Errorf("store: écriture projet '%s': %w", p.ID, err)
	}
	return nil
}

// Load charge un projet depuis son fichier XML.
func (s *ProjectStore) Load(projectID string) (*contracts.Project, error) {
	path := filepath.Join(s.projectsDir, projectID, "project.xml")
	return parser.ParseProjectFile(path)
}

// Delete supprime le répertoire d'un projet.
func (s *ProjectStore) Delete(projectID string) error {
	return os.RemoveAll(filepath.Join(s.projectsDir, projectID))
}

// SHA256 retourne le hash SHA256 du fichier XML courant d'un projet.
func (s *ProjectStore) SHA256(projectID string) (string, error) {
	data, err := os.ReadFile(filepath.Join(s.projectsDir, projectID, "project.xml"))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}
