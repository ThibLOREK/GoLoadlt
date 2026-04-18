package targets

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("target.csv", func() contracts.Block { return &CSVTarget{} })
}

// CSVTarget reçoit des lignes et les écrit dans un fichier CSV.
type CSVTarget struct{}

func (b *CSVTarget) Type() string { return "target.csv" }

func (b *CSVTarget) Run(bctx *contracts.BlockContext) error {
	path := bctx.Params["path"]
	if path == "" {
		return fmt.Errorf("target.csv: paramètre 'path' manquant")
	}

	appendMode := bctx.Params["append"] == "true"
	delimiter := ','
	if d, ok := bctx.Params["delimiter"]; ok && len(d) == 1 {
		delimiter = rune(d[0])
	}

	flags := os.O_CREATE | os.O_WRONLY
	if appendMode {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	f, err := os.OpenFile(path, flags, 0o644)
	if err != nil {
		return fmt.Errorf("target.csv: ouverture fichier '%s': %w", path, err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	writer.Comma = delimiter
	defer writer.Flush()

	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("target.csv: aucun port d'entrée")
	}
	in := bctx.Inputs[0]

	headerWritten := false
	var headers []string

	for {
		select {
		case <-bctx.Ctx.Done():
			return bctx.Ctx.Err()
		case row, ok := <-in.Ch:
			if !ok {
				return writer.Error()
			}
			if !headerWritten {
				for k := range row {
					headers = append(headers, k)
				}
				if err := writer.Write(headers); err != nil {
					return fmt.Errorf("target.csv: écriture en-tête: %w", err)
				}
				headerWritten = true
			}
			record := make([]string, len(headers))
			for i, h := range headers {
				if v, ok := row[h]; ok {
					record[i] = fmt.Sprintf("%v", v)
				}
			}
			if err := writer.Write(record); err != nil {
				return fmt.Errorf("target.csv: écriture ligne: %w", err)
			}
		}
	}
}
