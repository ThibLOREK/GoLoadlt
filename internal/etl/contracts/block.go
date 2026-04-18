package contracts

import "context"

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
}

// BlockContext contient le contexte d'exécution d'un bloc.
type BlockContext struct {
	Ctx       context.Context
	Params    map[string]string
	ConnRef   string   // référence à une connexion réutilisable
	ActiveEnv string   // environnement actif : dev, preprod, prod
	Inputs    []*Port  // ports d'entrée
	Outputs   []*Port  // ports de sortie
}

// Block est l'interface que tout nœud du graphe ETL doit implémenter.
type Block interface {
	// Type retourne le type unique du bloc (ex: "transform.filter").
	Type() string
	// Run exécute le bloc : lit depuis ses Inputs, écrit vers ses Outputs.
	Run(bctx *BlockContext) error
}

// BlockFactory crée une instance de Block.
type BlockFactory func() Block
