package sources

import (
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("source.text", func() contracts.Block { return &TextInput{} })
}

// TextInput permet de saisir des données en ligne dans l'UI (comme Alteryx Text Input).
// Paramètre 'data' : CSV inline, ex: "name,age\nAlice,30\nBob,25"
type TextInput struct{}

func (b *TextInput) Type() string { return "source.text" }

func (b *TextInput) Run(bctx *contracts.BlockContext) error {
	data := bctx.Params["data"]
	if data == "" {
		return fmt.Errorf("source.text: paramètre 'data' manquant")
	}
	data = strings.ReplaceAll(data, `\n`, "\n")

	reader := csv.NewReader(strings.NewReader(data))
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("source.text: lecture en-tête: %w", err)
	}

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("source.text: lecture données: %w", err)
	}

	for _, record := range records {
		select {
		case <-bctx.Ctx.Done():
			for _, out := range bctx.Outputs { close(out.Ch) }
			return bctx.Ctx.Err()
		default:
		}
		row := make(contracts.DataRow, len(headers))
		for i, h := range headers {
			if i < len(record) {
				row[h] = record[i]
			}
		}
		for _, out := range bctx.Outputs { out.Ch <- row }
	}
	for _, out := range bctx.Outputs { close(out.Ch) }
	return nil
}