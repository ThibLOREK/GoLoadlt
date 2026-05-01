// Package spec expose le contenu de openapi.yaml via go:embed.
// Ce sous-package est nécessaire car //go:embed ne peut référencer
// que des chemins relatifs situés sous le répertoire du fichier Go.
// openapi_handler.go (dans api/handlers/) importe ce package pour
// servir la spec sans dépendance de build supplémentaire.
package spec

import _ "embed"

//go:embed ../openapi.yaml
var Spec []byte
