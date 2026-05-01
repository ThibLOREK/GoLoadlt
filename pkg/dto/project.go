package dto

import "github.com/rinjold/go-etl-studio/internal/etl/contracts"

// ProjectCreateRequest est le corps de POST /api/v1/projects.
type ProjectCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ProjectUpdateRequest est le corps de PUT /api/v1/projects/{id}.
// Contient le graphe DAG complet sérialisé depuis l'UI.
type ProjectUpdateRequest struct {
	Project *contracts.Project `json:"project"`
}

// ProjectResponse est la représentation enrichie d'un projet retournée par l'API.
type ProjectResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	XMLPath     string `json:"xmlPath,omitempty"`
	SHA256      string `json:"sha256,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
	RunCount    int    `json:"runCount"`
	LastStatus  string `json:"lastStatus,omitempty"` // succeeded|failed|running|pending
}

// ProjectListResponse est la réponse paginable de GET /api/v1/projects.
type ProjectListResponse struct {
	Items []ProjectResponse `json:"items"`
	Total int               `json:"total"`
}
