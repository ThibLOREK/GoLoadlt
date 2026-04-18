package sources

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("source.csv", func() contracts.Block { return &CSVSource{} })
}

// CSVSource lit un fichier CSV et envoie chaque ligne vers le port de sortie.
type CSVSource struct{}

func (b *CSVSource) Type() string { return "source.csv" }

func (b *CSVSource) Run(bctx *contracts.BlockContext) error {
	path := bctx.Params["path"]
	if path == "" {
		return fmt.Errorf("source.csv: paramètre 'path' manquant")
	}

	delimiter := ','
	if d, ok := bctx.Params["delimiter"]; ok && len(d) == 1 {
		delimiter = rune(d[0])
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("source.csv: ouverture fichier '%s': %w", path, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = delimiter
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Lire l'en-tête.
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("source.csv: lecture en-tête: %w", err)
	}

	for {
		select {
		case <-bctx.Ctx.Done():
			return bctx.Ctx.Err()
		default:
		}

		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("source.csv: lecture ligne: %w", err)
		}

		row := make(contracts.DataRow, len(headers))
		for i, h := range headers {
			if i < len(record) {
				row[h] = record[i]
			}
		}

		// Envoyer la ligne à tous les ports de sortie.
		for _, out := range bctx.Outputs {
			select {
			case out.Ch <- row:
			case <-bctx.Ctx.Done():
				return bctx.Ctx.Err()
			}
		}
	}

	// Fermer tous les canaux de sortie pour signaler la fin du flux.
	for _, out := range bctx.Outputs {
		close(out.Ch)
	}
	return nil
}
