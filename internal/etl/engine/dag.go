package engine

import (
	"fmt"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

// DAG représente le graphe orienté acyclique d'un projet ETL.
type DAG struct {
	Nodes map[string]*contracts.Node
	// adjacency : nodeID → liste des nodeIDs successeurs
	adjacency map[string][]string
	// portMap : "nodeID:portID" → édge (pour blocs multi-sorties)
	portEdges map[string][]contracts.Edge
}

// BuildDAG construit le DAG à partir d'un Project.
func BuildDAG(p *contracts.Project) (*DAG, error) {
	d := &DAG{
		Nodes:     make(map[string]*contracts.Node, len(p.Nodes)),
		adjacency: make(map[string][]string),
		portEdges: make(map[string][]contracts.Edge),
	}

	for i := range p.Nodes {
		n := &p.Nodes[i]
		d.Nodes[n.ID] = n
		d.adjacency[n.ID] = []string{}
	}

	for _, e := range p.Edges {
		if _, ok := d.Nodes[e.From]; !ok {
			return nil, fmt.Errorf("dag: nœud source '%s' introuvable", e.From)
		}
		if _, ok := d.Nodes[e.To]; !ok {
			return nil, fmt.Errorf("dag: nœud cible '%s' introuvable", e.To)
		}
		d.adjacency[e.From] = append(d.adjacency[e.From], e.To)

		// Index par port de sortie (pour les blocs split).
		portKey := e.From + ":" + e.FromPort
		d.portEdges[portKey] = append(d.portEdges[portKey], e)
	}

	return d, nil
}

// TopologicalSort retourne les nodes dans l'ordre d'exécution (Kahn's algorithm).
func (d *DAG) TopologicalSort() ([]*contracts.Node, error) {
	// Calculer le degré d'entrée de chaque nœud.
	inDegree := make(map[string]int, len(d.Nodes))
	for id := range d.Nodes {
		inDegree[id] = 0
	}
	for _, successors := range d.adjacency {
		for _, s := range successors {
			inDegree[s]++
		}
	}

	// File des nœuds sans dépendance entrante.
	queue := make([]string, 0)
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	var sorted []*contracts.Node
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		sorted = append(sorted, d.Nodes[current])

		for _, successor := range d.adjacency[current] {
			inDegree[successor]--
			if inDegree[successor] == 0 {
				queue = append(queue, successor)
			}
		}
	}

	if len(sorted) != len(d.Nodes) {
		return nil, fmt.Errorf("dag: cycle détecté dans le graphe du projet")
	}
	return sorted, nil
}

// Successors retourne les nœuds successeurs d'un nœud donné.
func (d *DAG) Successors(nodeID string) []*contracts.Node {
	ids := d.adjacency[nodeID]
	nodes := make([]*contracts.Node, 0, len(ids))
	for _, id := range ids {
		if n, ok := d.Nodes[id]; ok {
			nodes = append(nodes, n)
		}
	}
	return nodes
}

// PortEdges retourne les edges associés à un port de sortie spécifique.
func (d *DAG) PortEdges(nodeID, portID string) []contracts.Edge {
	return d.portEdges[nodeID+":"+portID]
}
