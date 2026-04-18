package e2e

import (
	"context"
	"os"
	"testing"

	csvconn "github.com/rinjold/go-etl-studio/internal/connectors/csv"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
	"github.com/rinjold/go-etl-studio/internal/etl/engine"
	"github.com/rinjold/go-etl-studio/internal/etl/transformers"
	"github.com/rs/zerolog"
)

func TestPipeline_CSV_Transform_CSV(t *testing.T) {
	// Source
	srcFile, _ := os.CreateTemp("", "src_*.csv")
	defer os.Remove(srcFile.Name())
	_, _ = srcFile.WriteString("prenom,age,statut\nAlice,30,active\nBob,17,inactive\nCarla,25,active\n")
	_ = srcFile.Close()

	// Target
	dstFile, _ := os.CreateTemp("", "dst_*.csv")
	defer os.Remove(dstFile.Name())
	_ = dstFile.Close()

	extractor := csvconn.NewExtractor(csvconn.ExtractorConfig{
		FilePath: srcFile.Name(), Delimiter: ',', HasHeader: true,
	})

	transformer := transformers.Chain{Steps: []contracts.Transformer{
		transformers.Mapper{Mapping: map[string]string{"prenom": "first_name"}},
		transformers.Filter{Rules: []transformers.FilterRule{
			{Column: "statut", Operator: transformers.OpEq, Value: "active"},
		}},
		transformers.Caster{Rules: []transformers.CastRule{
			{Column: "age", CastTo: transformers.CastInt},
		}},
	}}

	loader := csvconn.NewLoader(csvconn.LoaderConfig{
		FilePath: dstFile.Name(), Delimiter: ',', HasHeader: true,
		Columns: []string{"first_name", "age", "statut"},
	})

	ex := engine.Executor{
		Extractor:   extractor,
		Transformer: transformer,
		Loader:      loader,
		Log:         zerolog.Nop(),
	}

	result := ex.Execute(context.Background())
	if result.Err != nil {
		t.Fatalf("pipeline failed: %v", result.Err)
	}
	if result.RecordsRead != 2 {
		t.Errorf("expected 2 records after filter, got %d", result.RecordsRead)
	}

	content, _ := os.ReadFile(dstFile.Name())
	t.Logf("output CSV:\n%s", string(content))
}
