package contracts

// Project représente un projet ETL complet (graphe de blocs).
type Project struct {
	ID          string  `xml:"id,attr"`
	Name        string  `xml:"name,attr"`
	Description string  `xml:"description,attr,omitempty"`
	Version     int     `xml:"version,attr"`
	ActiveEnv   string  `xml:"activeEnv,attr,omitempty"`
	Nodes       []Node  `xml:"nodes>node"`
	Edges       []Edge  `xml:"edges>edge"`
}

// Node représente un bloc dans le graphe ETL.
type Node struct {
	ID          string            `xml:"id,attr"`
	Type        string            `xml:"type,attr"`
	Label       string            `xml:"label,attr,omitempty"`
	ConnRef     string            `xml:"connectionRef,attr,omitempty"`
	// PosX et PosY stockent la position visuelle dans l'UI.
	PosX        float64           `xml:"posX,attr,omitempty"`
	PosY        float64           `xml:"posY,attr,omitempty"`
	Params      []Param           `xml:"params>param,omitempty"`
}

// ParamMap retourne les paramètres du node sous forme de map.
func (n *Node) ParamMap() map[string]string {
	m := make(map[string]string, len(n.Params))
	for _, p := range n.Params {
		m[p.Name] = p.Value
	}
	return m
}

// Param est un paramètre clé/valeur d'un bloc.
type Param struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// Edge représente un lien dirigé entre deux blocs.
// FromPort et ToPort permettent de gérer les blocs multi-sorties (ex: split).
type Edge struct {
	From     string `xml:"from,attr"`
	To       string `xml:"to,attr"`
	FromPort string `xml:"fromPort,attr,omitempty"` // ex: "out0", "out1" pour split
	ToPort   string `xml:"toPort,attr,omitempty"`
}
