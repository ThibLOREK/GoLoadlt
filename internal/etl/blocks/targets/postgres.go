package targets

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("target.postgres", func() contracts.Block { return &PostgresTarget{} })
}

const defaultBatchSize = 500

// PostgresTarget insère les lignes reçues dans une table PostgreSQL par batch.
type PostgresTarget struct{}

func (b *PostgresTarget) Type() string { return "target.postgres" }

func (b *PostgresTarget) Run(bctx *contracts.BlockContext) error {
	dsn := bctx.Params["dsn"]
	if dsn == "" {
		return fmt.Errorf("target.postgres: paramètre 'dsn' manquant")
	}
	table := bctx.Params["table"]
	if table == "" {
		return fmt.Errorf("target.postgres: paramètre 'table' manquant")
	}
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("target.postgres: aucun port d'entrée")
	}

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("target.postgres: connexion: %w", err)
	}
	defer conn.Close(context.Background())

	in := bctx.Inputs[0]
	var batch []contracts.DataRow
	var headers []string

	flushBatch := func() error {
		if len(batch) == 0 {
			return nil
		}
		placeholders := make([]string, len(headers))
		for i := range headers {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}
		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			table,
			strings.Join(headers, ", "),
			strings.Join(placeholders, ", "),
		)
		tx, err := conn.Begin(bctx.Ctx)
		if err != nil {
			return fmt.Errorf("target.postgres: begin tx: %w", err)
		}
		for _, row := range batch {
			args := make([]any, len(headers))
			for i, h := range headers {
				args[i] = row[h]
			}
			if _, err := tx.Exec(bctx.Ctx, query, args...); err != nil {
				_ = tx.Rollback(bctx.Ctx)
				return fmt.Errorf("target.postgres: insert: %w", err)
			}
		}
		batch = batch[:0]
		return tx.Commit(bctx.Ctx)
	}

	for {
		select {
		case <-bctx.Ctx.Done():
			return bctx.Ctx.Err()
		case row, ok := <-in.Ch:
			if !ok {
				return flushBatch()
			}
			if headers == nil {
				for k := range row {
					headers = append(headers, k)
				}
			}
			batch = append(batch, row)
			if len(batch) >= defaultBatchSize {
				if err := flushBatch(); err != nil {
					return err
				}
			}
		}
	}
}
