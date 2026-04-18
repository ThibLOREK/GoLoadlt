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
// Retourne toujours un slice non-nil pour que JSON encode [] et non null.
func (s *ProjectStore) ListAll() ([]*contracts.Project, error) {
	entries, err := os.ReadDir(s.projectsDir)
	if err != nil {
		return []*contracts.Project{}, err
	}
	projects := make([]*contracts.Project, 0)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p, err := s.Load(e.Name())
		if err != nil {
			continue
		}
		projects = append(projects, p)
	}
	return projects, nil
}
