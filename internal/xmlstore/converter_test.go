package xmlstore

import (
	"encoding/xml"
	"testing"
)

func TestXMLToDAG_Basic(t *testing.T) {
	p := &XMLPipeline{
		ID:      "test",
		Name:    "Basic",
		Version: 1,
		Nodes: []XMLNode{
			{
				ID:    "n1",
				Type:  "transform.filter",
				Label: "Filtrer FR",
				Params: []XMLParam{
					{Key: "condition", Value: "country == 'FR'"},
				},
			},
			{
				ID:    "n2",
				Type:  "transform.cast",
				Label: "Cast score",
				Params: []XMLParam{
					{Key: "column", Value: "score"},
					{Key: "targetType", Value: "int"},
				},
			},
		},
		Edges: []XMLEdge{
			{From: "n1", To: "n2", FromPort: 0, ToPort: 0},
		},
	}

	nodes, edges, err := XMLToDAG(p)
	if err != nil {
		t.Fatalf("XMLToDAG: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("nodes: want 2, got %d", len(nodes))
	}
	if len(edges) != 1 {
		t.Fatalf("edges: want 1, got %d", len(edges))
	}

	// Vérifie la map de params du nœud n1.
	var n1 *DAGNode
	for i := range nodes {
		if nodes[i].ID == "n1" {
			n1 = &nodes[i]
			break
		}
	}
	if n1 == nil {
		t.Fatal("node n1 not found")
	}
	if n1.Params["condition"] != "country == 'FR'" {
		t.Errorf("param condition: want %q, got %q", "country == 'FR'", n1.Params["condition"])
	}

	// Vérifie les params du nœud n2.
	var n2 *DAGNode
	for i := range nodes {
		if nodes[i].ID == "n2" {
			n2 = &nodes[i]
			break
		}
	}
	if n2 == nil {
		t.Fatal("node n2 not found")
	}
	if n2.Params["targetType"] != "int" {
		t.Errorf("param targetType: want %q, got %q", "int", n2.Params["targetType"])
	}

	// Vérifie l'arête.
	if edges[0].From != "n1" || edges[0].To != "n2" {
		t.Errorf("edge: want n1->n2, got %s->%s", edges[0].From, edges[0].To)
	}
}

func TestDAGToXML_Basic(t *testing.T) {
	nodes := []DAGNode{
		{ID: "a", Type: "source.csv", Label: "CSV Source", Params: map[string]string{"path": "/data/in.csv"}},
		{ID: "b", Type: "sink.postgres", Label: "PG Sink", Params: map[string]string{"table": "output"}},
	}
	edges := []DAGEdge{
		{From: "a", To: "b", FromPort: 0, ToPort: 0},
	}

	p := DAGToXML("dag-001", "DAG Pipeline", nodes, edges)
	if p == nil {
		t.Fatal("DAGToXML returned nil")
	}
	if p.ID != "dag-001" {
		t.Errorf("ID: want %q, got %q", "dag-001", p.ID)
	}
	if len(p.Nodes) != 2 {
		t.Fatalf("nodes: want 2, got %d", len(p.Nodes))
	}
	if len(p.Edges) != 1 {
		t.Fatalf("edges: want 1, got %d", len(p.Edges))
	}

	// Vérifie que le marshal XML ne produit pas d'erreur.
	_, err := xml.MarshalIndent(p, "", "  ")
	if err != nil {
		t.Errorf("xml.MarshalIndent: %v", err)
	}
}

func TestXMLToDAG_EmptyNodes(t *testing.T) {
	p := &XMLPipeline{
		ID:      "empty",
		Name:    "Empty Pipeline",
		Version: 1,
	}

	nodes, edges, err := XMLToDAG(p)
	if err != nil {
		t.Fatalf("XMLToDAG on empty pipeline: %v", err)
	}
	if len(nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(nodes))
	}
	if len(edges) != 0 {
		t.Errorf("expected 0 edges, got %d", len(edges))
	}
}

func TestXMLToDAG_InvalidEdge(t *testing.T) {
	tests := []struct {
		name string
		pipeline *XMLPipeline
	}{
		{
			name: "unknown source node",
			pipeline: &XMLPipeline{
				ID: "invalid-src",
				Nodes: []XMLNode{
					{ID: "n1", Type: "transform.filter", Label: "F"},
				},
				Edges: []XMLEdge{
					{From: "ghost", To: "n1"},
				},
			},
		},
		{
			name: "unknown target node",
			pipeline: &XMLPipeline{
				ID: "invalid-dst",
				Nodes: []XMLNode{
					{ID: "n1", Type: "transform.filter", Label: "F"},
				},
				Edges: []XMLEdge{
					{From: "n1", To: "ghost"},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := XMLToDAG(tc.pipeline)
			if err == nil {
				t.Error("expected error for invalid edge reference, got nil")
			}
		})
	}
}
