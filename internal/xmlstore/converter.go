package xmlstore

import "fmt"

// DAGNode représente un nœud du DAG pipeline en mémoire.
type DAGNode struct {
	ID     string
	Type   string
	Label  string
	Params map[string]string
}

// DAGEdge représente une connexion orientée entre deux nœuds du DAG.
type DAGEdge struct {
	From     string
	To       string
	FromPort int
	ToPort   int
}

// XMLToDAG convertit un XMLPipeline en listes de nœuds et d'arêtes
// utilisables par l'orchestrateur.
// Retourne une erreur si une arête référence un nœud inexistant.
func XMLToDAG(p *XMLPipeline) (nodes []DAGNode, edges []DAGEdge, err error) {
	// Index des IDs de nœuds pour la validation des arêtes.
	nodeIDs := make(map[string]struct{}, len(p.Nodes))

	for _, xn := range p.Nodes {
		params := make(map[string]string, len(xn.Params))
		for _, xp := range xn.Params {
			params[xp.Key] = xp.Value
		}
		nodes = append(nodes, DAGNode{
			ID:     xn.ID,
			Type:   xn.Type,
			Label:  xn.Label,
			Params: params,
		})
		nodeIDs[xn.ID] = struct{}{}
	}

	for _, xe := range p.Edges {
		if _, ok := nodeIDs[xe.From]; !ok {
			return nil, nil, fmt.Errorf("xmlstore.XMLToDAG: edge references unknown source node %q", xe.From)
		}
		if _, ok := nodeIDs[xe.To]; !ok {
			return nil, nil, fmt.Errorf("xmlstore.XMLToDAG: edge references unknown target node %q", xe.To)
		}
		edges = append(edges, DAGEdge{
			From:     xe.From,
			To:       xe.To,
			FromPort: xe.FromPort,
			ToPort:   xe.ToPort,
		})
	}

	// Garantir des slices non-nil même pour un pipeline vide.
	if nodes == nil {
		nodes = []DAGNode{}
	}
	if edges == nil {
		edges = []DAGEdge{}
	}
	return nodes, edges, nil
}

// DAGToXML convertit un DAG runtime en XMLPipeline sérialisable.
func DAGToXML(id, name string, nodes []DAGNode, edges []DAGEdge) *XMLPipeline {
	p := &XMLPipeline{
		ID:      id,
		Name:    name,
		Version: 1,
	}

	for _, dn := range nodes {
		xn := XMLNode{
			ID:    dn.ID,
			Type:  dn.Type,
			Label: dn.Label,
		}
		for k, v := range dn.Params {
			xn.Params = append(xn.Params, XMLParam{Key: k, Value: v})
		}
		p.Nodes = append(p.Nodes, xn)
	}

	for _, de := range edges {
		p.Edges = append(p.Edges, XMLEdge{
			From:     de.From,
			To:       de.To,
			FromPort: de.FromPort,
			ToPort:   de.ToPort,
		})
	}
	return p
}
