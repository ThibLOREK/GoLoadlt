package integration

import (
	"context"
	"os"
	"testing"

	csvconn "github.com/rinjold/go-etl-studio/internal/connectors/csv"
)

func TestCSVExtractor_Basic(t *testing.T) {
	tmp, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	_, _ = tmp.WriteString("name,age,city\nAlice,30,Paris\nBob,25,Lyon\n")
	_ = tmp.Close()

	extractor := csvconn.NewExtractor(csvconn.ExtractorConfig{
		FilePath:  tmp.Name(),
		Delimiter: ',',
		HasHeader: true,
	})

	records, err := extractor.Extract(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	if records[0]["name"] != "Alice" {
		t.Errorf("expected Alice, got %v", records[0]["name"])
	}
	if records[1]["city"] != "Lyon" {
		t.Errorf("expected Lyon, got %v", records[1]["city"])
	}
}
