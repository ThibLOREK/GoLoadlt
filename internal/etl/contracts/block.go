package contracts

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
)

// DataType représente le type d'une colonne.
type DataType string

const (
	TypeString  DataType = "string"
	TypeInt     DataType = "int"
	TypeFloat   DataType = "float"
	TypeBool    DataType = "bool"
	TypeDate    DataType = "date"
	TypeUnknown DataType = "unknown"
)

// ColumnDef décrit une colonne d'un schéma.
type ColumnDef struct {
	Name string
	Type DataType
}

// Schema décrit la structure d'un flux de données.
type Schema struct {
	Columns []ColumnDef
}

// DataRow est une ligne de données : map colonne → valeur.
type DataRow map[string]any

// Port est un canal de données entre deux blocs.
type Port struct {
	ID     string
	Schema Schema
	Ch     chan DataRow
	EdgeID string // Phase 11 : câblage multi-ports Fork/Merge
}

// ResolvedConnection transporte les informations de connexion résolues
// vers les blocs sources/targets après résolution des secrets.
// Ajouté Phase 6 — requis par inject_connections.go.
type ResolvedConnection struct {
	Type     string
	DSN      string
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

// BlockMetrics comptabilise les métriques d'exécution d'un bloc.
// Ajouté Phase 10 — requis par l'instrumentation OTel et les blocs Fork/Merge (Phase 11).
type BlockMetrics struct {
	RowsIn  int64
	RowsOut int64
}

// BlockContext contient le contexte d'exécution d'un bloc.
type BlockContext struct {
	Ctx           context.Context
	BlockID       string             // ID unique du bloc dans le projet (Phase 7+)
	BlockType     string             // type du bloc, ex: "transform.filter" (Phase 7+)
	RunID         string             // ID du run en cours (Phase 7+)
	Params        map[string]string
	ConnectionRef string             // référence à une connexion (remplace ConnRef)
	ActiveEnv     string             // environnement actif : dev, preprod, prod
	Preview       *PreviewStore      // capture des N premières lignes de sortie
	Inputs        []*Port            // ports d'entrée
	Outputs       []*Port            // ports de sortie
	Connection    *ResolvedConnection // connexion résolue (Phase 6+)
	Metrics       *BlockMetrics       // métriques d'exécution (Phase 10+)
	Logger        zerolog.Logger      // logger structuré par bloc (Phase 10+)
}

// Input retourne le port d'entrée par son ID, ou nil s'il est absent.
// Utilisé par tous les blocs : bctx.Input("in").
func (b *BlockContext) Input(id string) *Port {
	for _, p := range b.Inputs {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// Output retourne le port de sortie par son ID, ou nil s'il est absent.
// Utilisé par tous les blocs : bctx.Output("out").
func (b *BlockContext) Output(id string) *Port {
	for _, p := range b.Outputs {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// ErrMissingPort retourne une erreur standard pour un port manquant.
// Utilisé par Fork, Merge, PostgresCDC (Phase 11).
func ErrMissingPort(portID string) error {
	return fmt.Errorf("contracts: port requis '%s' absent du BlockContext", portID)
}

// Block est l'interface que tout nœud du graphe ETL doit implémenter.
type Block interface {
	// Run exécute le bloc : lit depuis ses Inputs, écrit vers ses Outputs.
	Run(bctx *BlockContext) error
}

// BlockFactory crée une instance de Block.
type BlockFactory func() Block
