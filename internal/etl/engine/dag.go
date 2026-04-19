package engine

import (
	"fmt"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

// DAG représente le graphe orienté acyclique d'un projet ETL.
type DAG struct {
	Nodes map[string]*contracts.Node
	// adjacency : nodeID → liste des nodeIDs successeurs (edges actifs uniquement)
	adjacency map[string][]string
	// portMap : "nodeID:portID" → edges (pour blocs multi-sorties, actifs uniquement)
	portEdges map[string][]contracts.Edge
}

// BuildDAG construit le DAG à partir d'un Project.
// Règles :
//   - Les edges disabled sont ignorés (le nœud cible est traité comme isolé).
//   - Les edges dont le nœud source ou cible est introuvable sont ignorés
//     avec un avertissement (évite le crash lors de la suppression d'un bloc
//     alors que les edges n'ont pas encore été sauvegardés).
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
		// Ignorer les edges désactivés : les blocs reliés deviennent des sources/targets isolés.
		if e.Disabled {
			continue
		}

		// Ignorer silencieusement les edges dont le nœud source est introuvable.
		// Cas typique : suppression d'un bloc avant que le save ne nettoie les edges.
		if _, ok := d.Nodes[e.From]; !ok {
			continue
		}
		// Idem pour le nœud cible.
		if _, ok := d.Nodes[e.To]; !ok {
			continue
		}

		d.adjacency[e.From] = append(d.adjacency[e.From], e.To)

		portKey := e.From + ":" + e.FromPort
		d.portEdges[portKey] = append(d.portEdges[portKey], e)
	}

	return d, nil
}

// TopologicalSort retourne les nodes dans l'ordre d'exécution (Kahn's algorithm).
func (d *DAG) TopologicalSort() ([]*contracts.Node, error) {
	inDegree := make(map[string]int, len(d.Nodes))
	for id := range d.Nodes {
		inDegree[id] = 0
	}
	for _, successors := range d.adjacency {
		for _, s := range successors {
			inDegree[s]++
		}
	}

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

// Successors retourne les nœuds successeurs actifs d'un nœud donné.
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

// PortEdges retourne les edges actifs associés à un port de sortie spécifique.
func (d *DAG) PortEdges(nodeID, portID string) []contracts.Edge {
	return d.portEdges[nodeID+":"+portID]
}
