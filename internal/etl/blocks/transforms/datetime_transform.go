package transforms

import (
	"fmt"
	"strconv"
	"time"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.datetime", func() contracts.Block { return &DateTimeTransform{} })
}

// DateTimeTransform parse, formate ou calcule des dates.
// Paramètres :
//   column      : colonne source
//   mode        : parse | format | add | diff | extract
//   inputFormat : format d'entrée Go (ex: "2006-01-02")
//   outputFormat: format de sortie
//   addUnit     : unit (days, hours, minutes) + addValue
//   extract     : year | month | day | weekday | hour | minute | second
//   output      : colonne de sortie (défaut: column)
type DateTimeTransform struct{}

func (b *DateTimeTransform) Type() string { return "transform.datetime" }

func (b *DateTimeTransform) Run(bctx *contracts.BlockContext) error {
	col := bctx.Params["column"]
	mode := bctx.Params["mode"]
	inputFmt := bctx.Params["inputFormat"]
	outputFmt := bctx.Params["outputFormat"]
	outCol := bctx.Params["output"]
	if col == "" { return fmt.Errorf("transform.datetime: 'column' obligatoire") }
	if mode == "" { mode = "parse" }
	if inputFmt == "" { inputFmt = "2006-01-02" }
	if outputFmt == "" { outputFmt = "2006-01-02T15:04:05" }
	if outCol == "" { outCol = col }

	in := bctx.Inputs[0]
	for {
		select {
		case <-bctx.Ctx.Done():
			for _, out := range bctx.Outputs { close(out.Ch) }
			return bctx.Ctx.Err()
		case row, ok := <-in.Ch:
			if !ok {
				for _, out := range bctx.Outputs { close(out.Ch) }
				return nil
			}
			newRow := make(contracts.DataRow, len(row))
			for k, v := range row { newRow[k] = v }

			str := fmt.Sprintf("%v", row[col])
			t, err := time.Parse(inputFmt, str)
			if err != nil {
				for _, out := range bctx.Outputs { out.Ch <- newRow }
				continue
			}

			switch mode {
			case "parse", "format":
				newRow[outCol] = t.Format(outputFmt)
			case "add":
				addVal, _ := strconv.ParseFloat(bctx.Params["addValue"], 64)
				switch bctx.Params["addUnit"] {
				case "days": t = t.AddDate(0, 0, int(addVal))
				case "hours": t = t.Add(time.Duration(addVal) * time.Hour)
				case "minutes": t = t.Add(time.Duration(addVal) * time.Minute)
				}
				newRow[outCol] = t.Format(outputFmt)
			case "extract":
				switch bctx.Params["extract"] {
				case "year": newRow[outCol] = t.Year()
				case "month": newRow[outCol] = int(t.Month())
				case "day": newRow[outCol] = t.Day()
				case "weekday": newRow[outCol] = t.Weekday().String()
				case "hour": newRow[outCol] = t.Hour()
				case "minute": newRow[outCol] = t.Minute()
				case "second": newRow[outCol] = t.Second()
				}
			}
			for _, out := range bctx.Outputs { out.Ch <- newRow }
		}
	}
}