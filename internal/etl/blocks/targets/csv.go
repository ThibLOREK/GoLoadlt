package targets

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("target.csv", func() contracts.Block { return &CSVTarget{} })
}

// CSVTarget écrit les lignes reçues dans un fichier CSV.
type CSVTarget struct{}

func (b *CSVTarget) Type() string { return "target.csv" }

func (b *CSVTarget) Run(bctx *contracts.BlockContext) error {
	path := bctx.Params["path"]
	if path == "" {
		return fmt.Errorf("target.csv: paramètre 'path' manquant")
	}
	delimiter := ','
	if d := bctx.Params["delimiter"]; len(d) == 1 {
		delimiter = rune(d[0])
	}
	appendMode := bctx.Params["append"] == "true"

	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if appendMode {
		flag = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	}
	f, err := os.OpenFile(path, flag, 0o644)
	if err != nil {
		return fmt.Errorf("target.csv: ouverture fichier '%s': %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.Comma = delimiter
	defer w.Flush()

	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("target.csv: aucun port d'entrée")
	}
	in := bctx.Inputs[0]

	var headers []string
	first := true
	for {
		select {
		case <-bctx.Ctx.Done():
			return bctx.Ctx.Err()
		case row, ok := <-in.Ch:
			if !ok {
				return nil
			}
			if first {
				headers = make([]string, 0, len(row))
				for k := range row {
					headers = append(headers, k)
				}
				sort.Strings(headers)
				if !appendMode {
					_ = w.Write(headers)
				}
				first = false
			}
			record := make([]string, len(headers))
			for i, h := range headers {
				record[i] = fmt.Sprintf("%v", row[h])
			}
			if err := w.Write(record); err != nil {
				return fmt.Errorf("target.csv: écriture ligne: %w", err)
			}
		}
	}
}
