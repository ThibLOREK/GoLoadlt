package serializer

import (
	"encoding/xml"
	"fmt"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

// SerializeProject sérialise un Project en bytes XML indentés.
func SerializeProject(p *contracts.Project) ([]byte, error) {
	data, err := xml.MarshalIndent(p, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("serializer: marshal projet: %w", err)
	}
	return append([]byte(xml.Header), data...), nil
}
