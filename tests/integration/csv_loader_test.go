package integration

import (
	"context"
	"os"
	"testing"

	csvconn "github.com/rinjold/go-etl-studio/internal/connectors/csv"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func TestCSVLoader_Basic(t *testing.T) {
	tmp, err := os.CreateTemp("", "output_*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	_ = tmp.Close()

	loader := csvconn.NewLoader(csvconn.LoaderConfig{
		FilePath:  tmp.Name(),
		Delimiter: ',',
		HasHeader: true,
		Columns:   []string{"name", "age"},
	})

	records := []contracts.Record{
		{"name": "Alice", "age": 30},
		{"name": "Bob", "age": 25},
	}

	if err := loader.Load(context.Background(), records); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}

	expected := "name,age\nAlice,30\nBob,25\n"
	if string(content) != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, string(content))
	}
}
