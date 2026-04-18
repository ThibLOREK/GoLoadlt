package serializer

import (
	"encoding/xml"
	"fmt"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

// Serialize convertit un Project en bytes XML (avec indentation).
func Serialize(p *contracts.Project) ([]byte, error) {
	data, err := xml.MarshalIndent(p, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("serializer: marshal XML: %w", err)
	}
	// Préfixer avec la déclaration XML standard.
	header := []byte(xml.Header)
	return append(header, data...), nil
}
