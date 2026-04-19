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

// CSVSource lit un fichier CSV et envoie chaque ligne vers le(s) port(s) de sortie.
type CSVSource struct{}

func (b *CSVSource) Type() string { return "source.csv" }

func (b *CSVSource) Run(bctx *contracts.BlockContext) error {
	path := bctx.Params["path"]
	if path == "" {
		return fmt.Errorf("source.csv: paramètre 'path' manquant")
	}

	if len(bctx.Outputs) == 0 {
		return fmt.Errorf("source.csv: aucun port de sortie connecté")
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

	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("source.csv: lecture en-tête: %w", err)
	}

	for {
		select {
		case <-bctx.Ctx.Done():
			// Fermer les sorties avant de retourner l'annulation.
			for _, out := range bctx.Outputs {
				close(out.Ch)
			}
			return bctx.Ctx.Err()
		default:
		}

		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Fermer les sorties en cas d'erreur pour ne pas bloquer les consommateurs.
			for _, out := range bctx.Outputs {
				close(out.Ch)
			}
			return fmt.Errorf("source.csv: lecture ligne: %w", err)
		}

		row := make(contracts.DataRow, len(headers))
		for i, h := range headers {
			if i < len(record) {
				row[h] = record[i]
			}
		}

		for _, out := range bctx.Outputs {
			select {
			case out.Ch <- row:
			case <-bctx.Ctx.Done():
				for _, o := range bctx.Outputs {
					close(o.Ch)
				}
				return bctx.Ctx.Err()
			}
		}
	}

	// Fin normale : fermer tous les canaux de sortie.
	for _, out := range bctx.Outputs {
		close(out.Ch)
	}
	return nil
}
