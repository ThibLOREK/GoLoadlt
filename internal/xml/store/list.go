package store

import (
	"os"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

// ProjectsDir retourne le répertoire des projets (utilisé par le worker).
func (s *ProjectStore) ProjectsDir() string {
	return s.projectsDir
}

// ListAll charge et retourne tous les projets du store.
func (s *ProjectStore) ListAll() ([]*contracts.Project, error) {
	entries, err := os.ReadDir(s.projectsDir)
	if err != nil {
		return nil, err
	}
	var projects []*contracts.Project
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p, err := s.Load(e.Name())
		if err != nil {
			continue // projet sans fichier XML valide, on l'ignore
		}
		projects = append(projects, p)
	}
	return projects, nil
}
