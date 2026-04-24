package contracts

// Project représente un projet ETL complet (graphe de blocs).
type Project struct {
	ID          string `xml:"id,attr"                       json:"id"`
	Name        string `xml:"name,attr"                      json:"name"`
	Description string `xml:"description,attr,omitempty"     json:"description,omitempty"`
	Version     int    `xml:"version,attr"                   json:"version"`
	ActiveEnv   string `xml:"activeEnv,attr,omitempty"       json:"activeEnv,omitempty"`
	Nodes       []Node `xml:"nodes>node"                     json:"nodes"`
	Edges       []Edge `xml:"edges>edge"                     json:"edges"`
}

// Node représente un bloc dans le graphe ETL.
type Node struct {
	ID      string  `xml:"id,attr"                     json:"id"`
	Type    string  `xml:"type,attr"                    json:"type"`
	Label   string  `xml:"label,attr,omitempty"         json:"label,omitempty"`
	ConnRef string  `xml:"connectionRef,attr,omitempty" json:"connectionRef,omitempty"`
	PosX    float64 `xml:"posX,attr,omitempty"          json:"posX,omitempty"`
	PosY    float64 `xml:"posY,attr,omitempty"          json:"posY,omitempty"`
	Params  []Param `xml:"params>param,omitempty"       json:"params,omitempty"`
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
// Phase 11 : ajout de SourcePort et TargetPort pour les blocs multi-ports (Fork/Merge).
// Les champs SourcePort/TargetPort sont omitempty pour assurer la rétro-compatibilité
// avec tous les fichiers XML générés par les phases 5 à 10.
type Edge struct {
	ID         string `xml:"id,attr,omitempty"           json:"id,omitempty"`
	Source     string `xml:"source,attr"                 json:"source"`       // Phase 11 (= ancien From)
	Target     string `xml:"target,attr"                 json:"target"`       // Phase 11 (= ancien To)
	From       string `xml:"from,attr,omitempty"         json:"from,omitempty"`   // compat phases 5-10
	To         string `xml:"to,attr,omitempty"           json:"to,omitempty"`     // compat phases 5-10
	SourcePort string `xml:"sourcePort,attr,omitempty"   json:"sourcePort,omitempty"` // Phase 11 Fork/Merge
	TargetPort string `xml:"targetPort,attr,omitempty"   json:"targetPort,omitempty"` // Phase 11 Fork/Merge
	FromPort   string `xml:"fromPort,attr,omitempty"     json:"fromPort,omitempty"`   // compat phases 5-10
	ToPort     string `xml:"toPort,attr,omitempty"       json:"toPort,omitempty"`     // compat phases 5-10
	Disabled   bool   `xml:"disabled,attr,omitempty"    json:"disabled,omitempty"`
}

// EffectiveSource retourne le nœud source quel que soit le format XML (Phase 5-10 ou Phase 11).
func (e *Edge) EffectiveSource() string {
	if e.Source != "" {
		return e.Source
	}
	return e.From
}

// EffectiveTarget retourne le nœud cible quel que soit le format XML (Phase 5-10 ou Phase 11).
func (e *Edge) EffectiveTarget() string {
	if e.Target != "" {
		return e.Target
	}
	return e.To
}

// EffectiveSourcePort retourne le port source quel que soit le format XML.
func (e *Edge) EffectiveSourcePort() string {
	if e.SourcePort != "" {
		return e.SourcePort
	}
	return e.FromPort
}

// EffectiveTargetPort retourne le port cible quel que soit le format XML.
func (e *Edge) EffectiveTargetPort() string {
	if e.TargetPort != "" {
		return e.TargetPort
	}
	return e.ToPort
}

// ProjectTemplate représente un pipeline réutilisable paramétrable.
// Ajouté Phase 11 — Templates axis.
type ProjectTemplate struct {
	ID          string          `xml:"id,attr"          json:"id"`
	Name        string          `xml:"name,attr"        json:"name"`
	Description string          `xml:"description,attr" json:"description"`
	Category    string          `xml:"category,attr"    json:"category"`
	Params      []TemplateParam `xml:"params>param"     json:"params"`
	Nodes       []Node          `xml:"nodes>node"       json:"nodes"`
	Edges       []Edge          `xml:"edges>edge"       json:"edges"`
}

// TemplateParam décrit un paramètre substituable dans le template.
// Ajouté Phase 11 — Templates axis.
type TemplateParam struct {
	Key         string `xml:"key,attr"         json:"key"`
	Label       string `xml:"label,attr"       json:"label"`
	Description string `xml:"description,attr" json:"description"`
	Default     string `xml:"default,attr"     json:"default"`
	Required    bool   `xml:"required,attr"    json:"required"`
}
