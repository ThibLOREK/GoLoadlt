package contracts

// Project représente un projet ETL complet (graphe de blocs).
type Project struct {
	ID          string  `xml:"id,attr"          json:"id"`
	Name        string  `xml:"name,attr"         json:"name"`
	Description string  `xml:"description,attr,omitempty" json:"description,omitempty"`
	Version     int     `xml:"version,attr"      json:"version"`
	ActiveEnv   string  `xml:"activeEnv,attr,omitempty" json:"activeEnv,omitempty"`
	Nodes       []Node  `xml:"nodes>node"        json:"nodes"`
	Edges       []Edge  `xml:"edges>edge"        json:"edges"`
}

// Node représente un bloc dans le graphe ETL.
type Node struct {
	ID      string  `xml:"id,attr"                    json:"id"`
	Type    string  `xml:"type,attr"                   json:"type"`
	Label   string  `xml:"label,attr,omitempty"        json:"label,omitempty"`
	ConnRef string  `xml:"connectionRef,attr,omitempty" json:"connectionRef,omitempty"`
	PosX    float64 `xml:"posX,attr,omitempty"         json:"posX,omitempty"`
	PosY    float64 `xml:"posY,attr,omitempty"         json:"posY,omitempty"`
	Params  []Param `xml:"params>param,omitempty"      json:"params,omitempty"`
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
	Name  string `xml:"name,attr"  json:"name"`
	Value string `xml:"value,attr" json:"value"`
}

// Edge représente un lien dirigé entre deux blocs.
type Edge struct {
	From     string `xml:"from,attr"             json:"from"`
	To       string `xml:"to,attr"               json:"to"`
	FromPort string `xml:"fromPort,attr,omitempty" json:"fromPort,omitempty"`
	ToPort   string `xml:"toPort,attr,omitempty"   json:"toPort,omitempty"`
}
