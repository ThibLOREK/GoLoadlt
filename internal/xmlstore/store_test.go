package xmlstore

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestPipeline(id string) *XMLPipeline {
	return &XMLPipeline{
		ID:      id,
		Name:    "Test Pipeline",
		Version: 1,
		Nodes: []XMLNode{
			{
				ID:    "n1",
				Type:  "transform.filter",
				Label: "Filter FR",
				Params: []XMLParam{
					{Key: "condition", Value: "country == 'FR'"},
				},
			},
		},
		Edges: []XMLEdge{
			{From: "n1", To: "n2", FromPort: 0, ToPort: 0},
		},
	}
}

func newStore(t *testing.T) *FileStore {
	t.Helper()
	fs, err := NewFileStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	return fs
}

func TestSave_Load_RoundTrip(t *testing.T) {
	ctx := context.Background()
	store := newStore(t)
	orig := newTestPipeline("pipeline-rt")

	if err := store.Save(ctx, orig); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := store.Load(ctx, orig.ID)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.ID != orig.ID {
		t.Errorf("ID: want %q, got %q", orig.ID, got.ID)
	}
	if got.Name != orig.Name {
		t.Errorf("Name: want %q, got %q", orig.Name, got.Name)
	}
	if got.Version != orig.Version {
		t.Errorf("Version: want %d, got %d", orig.Version, got.Version)
	}
	if len(got.Nodes) != len(orig.Nodes) {
		t.Fatalf("Nodes len: want %d, got %d", len(orig.Nodes), len(got.Nodes))
	}
	if got.Nodes[0].ID != orig.Nodes[0].ID {
		t.Errorf("Node ID: want %q, got %q", orig.Nodes[0].ID, got.Nodes[0].ID)
	}
	if got.Nodes[0].Params[0].Value != orig.Nodes[0].Params[0].Value {
		t.Errorf("Param value: want %q, got %q",
			orig.Nodes[0].Params[0].Value, got.Nodes[0].Params[0].Value)
	}
	if len(got.Edges) != len(orig.Edges) {
		t.Fatalf("Edges len: want %d, got %d", len(orig.Edges), len(got.Edges))
	}
	if got.Edges[0].From != orig.Edges[0].From {
		t.Errorf("Edge From: want %q, got %q", orig.Edges[0].From, got.Edges[0].From)
	}
}

func TestLoad_NotFound(t *testing.T) {
	ctx := context.Background()
	store := newStore(t)

	_, err := store.Load(ctx, "does-not-exist")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected os.ErrNotExist in error chain, got: %v", err)
	}
}

func TestList_Empty(t *testing.T) {
	ctx := context.Background()
	store := newStore(t)

	list, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d items", len(list))
	}
}

func TestList_MultipleFiles(t *testing.T) {
	ctx := context.Background()
	store := newStore(t)

	for _, id := range []string{"p1", "p2", "p3"} {
		if err := store.Save(ctx, newTestPipeline(id)); err != nil {
			t.Fatalf("Save %s: %v", id, err)
		}
	}
	list, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 pipelines, got %d", len(list))
	}
}

func TestDelete_RemovesFile(t *testing.T) {
	ctx := context.Background()
	store := newStore(t)
	p := newTestPipeline("to-delete")

	if err := store.Save(ctx, p); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := store.Delete(ctx, p.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := store.Load(ctx, p.ID)
	if err == nil {
		t.Fatal("expected error after Delete, got nil")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected os.ErrNotExist, got: %v", err)
	}
}

func TestSave_Atomic(t *testing.T) {
	ctx := context.Background()
	store := newStore(t)
	p := newTestPipeline("atomic-test")

	if err := store.Save(ctx, p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Aucun fichier temporaire ne doit subsister après Save.
	entries, err := os.ReadDir(store.baseDir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp-") {
			t.Errorf("tmp file leftover: %s", e.Name())
		}
	}
	// Le fichier cible doit exister.
	expected := filepath.Join(store.baseDir, p.ID+".xml")
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Errorf("destination file missing: %s", expected)
	}
}

func TestXML_Roundtrip_Params(t *testing.T) {
	ctx := context.Background()
	store := newStore(t)

	specialValues := []struct {
		key   string
		value string
	}{
		{"lt", "a < b"},
		{"gt", "a > b"},
		{"amp", "a & b"},
		{"quot", `a " b`},
	}

	p := &XMLPipeline{
		ID:      "special-chars",
		Name:    "Special Chars Pipeline",
		Version: 1,
		Nodes: []XMLNode{
			{
				ID:    "n1",
				Type:  "transform.filter",
				Label: "Special",
				Params: func() []XMLParam {
					var ps []XMLParam
					for _, sv := range specialValues {
						ps = append(ps, XMLParam{Key: sv.key, Value: sv.value})
					}
					return ps
				}(),
			},
		},
	}

	if err := store.Save(ctx, p); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := store.Load(ctx, p.ID)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got.Nodes[0].Params) != len(specialValues) {
		t.Fatalf("params len: want %d, got %d", len(specialValues), len(got.Nodes[0].Params))
	}
	for i, sv := range specialValues {
		if got.Nodes[0].Params[i].Value != sv.value {
			t.Errorf("param[%d] %q: want %q, got %q",
				i, sv.key, sv.value, got.Nodes[0].Params[i].Value)
		}
	}
}
