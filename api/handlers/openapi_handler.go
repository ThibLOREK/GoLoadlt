package handlers

import (
	"net/http"

	"github.com/rinjold/go-etl-studio/api/openapi/spec"
)

// OpenAPIHandler sert la spécification OpenAPI et la Swagger UI.
type OpenAPIHandler struct{}

// NewOpenAPIHandler instancie l'OpenAPIHandler (sans dépendances).
func NewOpenAPIHandler() *OpenAPIHandler { return &OpenAPIHandler{} }

// ServeSpec sert le fichier openapi.yaml embarqué via go:embed.
// GET /api/v1/openapi.yaml
func (h *OpenAPIHandler) ServeSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(spec.Spec)
}

// SwaggerUI sert l'interface Swagger UI via CDN unpkg (sans dépendance npm).
// GET /api/docs
func (h *OpenAPIHandler) SwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html lang="fr">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>GoLoadIt — API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    body { margin: 0; }
    #swagger-ui .topbar { background-color: #1a1a2e; }
    #swagger-ui .topbar-wrapper .link { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function () {
      SwaggerUIBundle({
        url: "/api/v1/openapi.yaml",
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      });
    };
  </script>
</body>
</html>`))
}
