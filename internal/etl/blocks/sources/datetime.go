package sources

import (
	"time"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("source.datetime", func() contracts.Block { return &CurrentDateTime{} })
}

// CurrentDateTime émet une ligne unique avec la date/heure actuelle.
type CurrentDateTime struct{}

func (b *CurrentDateTime) Type() string { return "source.datetime" }

func (b *CurrentDateTime) Run(bctx *contracts.BlockContext) error {
	now := time.Now()
	row := contracts.DataRow{
		"datetime":    now.Format("2006-01-02T15:04:05"),
		"date":        now.Format("2006-01-02"),
		"time":        now.Format("15:04:05"),
		"year":        now.Year(),
		"month":       int(now.Month()),
		"day":         now.Day(),
		"hour":        now.Hour(),
		"minute":      now.Minute(),
		"second":      now.Second(),
		"unix":        now.Unix(),
		"weekday":     now.Weekday().String(),
	}
	for _, out := range bctx.Outputs {
		select {
		case out.Ch <- row:
		case <-bctx.Ctx.Done():
			for _, o := range bctx.Outputs { close(o.Ch) }
			return bctx.Ctx.Err()
		}
	}
	for _, out := range bctx.Outputs { close(out.Ch) }
	return nil
}