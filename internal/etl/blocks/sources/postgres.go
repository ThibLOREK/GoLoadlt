package sources

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("source.postgres", func() contracts.Block { return &PostgresSource{} })
}

// PostgresSource exécute une requête SQL sur une connexion PostgreSQL référencée.
// Le DSN est passé via le paramètre "dsn" (résolu par le resolver avant l'exécution).
type PostgresSource struct{}

func (b *PostgresSource) Type() string { return "source.postgres" }

func (b *PostgresSource) Run(bctx *contracts.BlockContext) error {
	dsn := bctx.Params["dsn"]
	if dsn == "" {
		return fmt.Errorf("source.postgres: paramètre 'dsn' manquant")
	}
	query := bctx.Params["query"]
	if query == "" {
		return fmt.Errorf("source.postgres: paramètre 'query' manquant")
	}

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("source.postgres: connexion: %w", err)
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(bctx.Ctx, query)
	if err != nil {
		return fmt.Errorf("source.postgres: requête: %w", err)
	}
	defer rows.Close()

	fields := rows.FieldDescriptions()
	headers := make([]string, len(fields))
	for i, f := range fields {
		headers[i] = string(f.Name)
	}

	for rows.Next() {
		select {
		case <-bctx.Ctx.Done():
			return bctx.Ctx.Err()
		default:
		}
		vals, err := rows.Values()
		if err != nil {
			return fmt.Errorf("source.postgres: lecture ligne: %w", err)
		}
		row := make(contracts.DataRow, len(headers))
		for i, h := range headers {
			row[h] = vals[i]
		}
		for _, out := range bctx.Outputs {
			select {
			case out.Ch <- row:
			case <-bctx.Ctx.Done():
				return bctx.Ctx.Err()
			}
		}
	}
	for _, out := range bctx.Outputs {
		close(out.Ch)
	}
	return rows.Err()
}
