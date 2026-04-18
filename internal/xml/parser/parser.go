package parser

import (
	"encoding/xml"
	"fmt"
	"os"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

// ParseFile lit un fichier XML et retourne un Project.
func ParseFile(path string) (*contracts.Project, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("parser: lecture fichier %s: %w", path, err)
	}
	return ParseBytes(data)
}

// ParseBytes désérialise un slice de bytes XML en Project.
func ParseBytes(data []byte) (*contracts.Project, error) {
	var project contracts.Project
	if err := xml.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("parser: unmarshal XML: %w", err)
	}
	if err := validate(&project); err != nil {
		return nil, err
	}
	return &project, nil
}

// validate vérifie la cohérence minimale du projet.
func validate(p *contracts.Project) error {
	if p.ID == "" {
		return fmt.Errorf("parser: le projet n'a pas d'ID")
	}
	if len(p.Nodes) == 0 {
		return fmt.Errorf("parser: le projet '%s' ne contient aucun bloc", p.ID)
	}
	// Vérifier que chaque edge référence des nodes existants.
	nodeIDs := make(map[string]bool, len(p.Nodes))
	for _, n := range p.Nodes {
		if n.ID == "" {
			return fmt.Errorf("parser: un bloc n'a pas d'ID dans le projet '%s'", p.ID)
		}
		if n.Type == "" {
			return fmt.Errorf("parser: le bloc '%s' n'a pas de type", n.ID)
		}
		nodeIDs[n.ID] = true
	}
	for _, e := range p.Edges {
		if !nodeIDs[e.From] {
			return fmt.Errorf("parser: edge référence un bloc source inconnu '%s'", e.From)
		}
		if !nodeIDs[e.To] {
			return fmt.Errorf("parser: edge référence un bloc cible inconnu '%s'", e.To)
		}
	}
	return nil
}
