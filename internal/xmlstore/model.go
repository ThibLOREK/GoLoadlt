package xmlstore

import "encoding/xml"

// XMLPipeline représente la structure racine d'un pipeline ETL sérialisé en XML.
type XMLPipeline struct {
	XMLName xml.Name  `xml:"pipeline"`
	ID      string    `xml:"id,attr"`
	Name    string    `xml:"name,attr"`
	Version int       `xml:"version,attr"`
	Nodes   []XMLNode `xml:"nodes>node"`
	Edges   []XMLEdge `xml:"edges>edge"`
}

// XMLNode représente un bloc (nœud) dans le DAG du pipeline.
type XMLNode struct {
	XMLName xml.Name   `xml:"node"`
	ID      string     `xml:"id,attr"`
	Type    string     `xml:"type,attr"`
	Label   string     `xml:"label,attr"`
	Params  []XMLParam `xml:"params>param"`
}

// XMLParam est une paire clé/valeur attachée à un nœud.
type XMLParam struct {
	XMLName xml.Name `xml:"param"`
	Key     string   `xml:"key,attr"`
	Value   string   `xml:"value,attr"`
}

// XMLEdge représente une connexion orientée entre deux nœuds du DAG.
type XMLEdge struct {
	XMLName  xml.Name `xml:"edge"`
	From     string   `xml:"from,attr"`
	To       string   `xml:"to,attr"`
	FromPort int      `xml:"fromPort,attr"`
	ToPort   int      `xml:"toPort,attr"`
}
