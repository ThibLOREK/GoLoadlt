package dto

import "github.com/rinjold/go-etl-studio/internal/etl/engine"

// RunRequest est le corps optionnel de POST /api/v1/projects/{id}/runs.
// Le champ Env permet de surcharger l'environnement actif pour ce run uniquement.
type RunRequest struct {
	Env string `json:"env,omitempty"`
}

// RunResponse est la réponse retournée après le lancement d'un run.
type RunResponse struct {
	RunID     string                  `json:"runId"`
	ProjectID string                  `json:"projectId"`
	Status    string                  `json:"status"` // pending|running|succeeded|failed|cancelled
	StartedAt string                  `json:"startedAt"`
	Report    *engine.ExecutionReport `json:"report,omitempty"`
}

// RunLogEntry représente une ligne de log structurée produite par un bloc.
type RunLogEntry struct {
	BlockID   string `json:"blockId"`
	BlockType string `json:"blockType"`
	Level     string `json:"level"` // info|warn|error
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	RowsIn    int64  `json:"rowsIn"`
	RowsOut   int64  `json:"rowsOut"`
}
