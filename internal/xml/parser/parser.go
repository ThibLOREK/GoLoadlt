package parser

import (
	"encoding/xml"
	"fmt"
	"os"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

// ParseProjectFile lit un fichier XML et retourne le Project correspondant.
func ParseProjectFile(path string) (*contracts.Project, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("parser: lecture '%s': %w", path, err)
	}
	return ParseProjectBytes(data)
}

// ParseProjectBytes parse un Project depuis des bytes XML.
func ParseProjectBytes(data []byte) (*contracts.Project, error) {
	var p contracts.Project
	if err := xml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parser: unmarshal projet: %w", err)
	}
	return &p, nil
}
