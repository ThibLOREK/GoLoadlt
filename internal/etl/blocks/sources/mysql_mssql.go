package sources

import (
	"fmt"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("source.mysql", func() contracts.Block { return &MySQLSource{} })
	blocks.Register("source.mssql", func() contracts.Block { return &MSSQLSource{} })
}

// MySQLSource et MSSQLSource sont des wrappers MVP.
// Pour l'instant ils valident les paramètres et renvoient une erreur explicite
// tant que les drivers / implémentations natives ne sont pas branchés.

type MySQLSource struct{}

type MSSQLSource struct{}

func (b *MySQLSource) Type() string { return "source.mysql" }
func (b *MSSQLSource) Type() string { return "source.mssql" }

func (b *MySQLSource) Run(bctx *contracts.BlockContext) error {
	if bctx.Params["dsn"] == "" {
		return fmt.Errorf("source.mysql: paramètre 'dsn' manquant")
	}
	if bctx.Params["query"] == "" {
		return fmt.Errorf("source.mysql: paramètre 'query' manquant")
	}
	return fmt.Errorf("source.mysql: implémentation native non encore branchée")
}

func (b *MSSQLSource) Run(bctx *contracts.BlockContext) error {
	if bctx.Params["dsn"] == "" {
		return fmt.Errorf("source.mssql: paramètre 'dsn' manquant")
	}
	if bctx.Params["query"] == "" {
		return fmt.Errorf("source.mssql: paramètre 'query' manquant")
	}
	return fmt.Errorf("source.mssql: implémentation native non encore branchée")
}
