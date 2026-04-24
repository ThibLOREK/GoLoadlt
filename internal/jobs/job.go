package jobs

import (
	"context"
	"time"
)

// Status représente l'état d'un run.
type Status string

const (
	Pending   Status = "pending"
	Running   Status = "running"
	Failed    Status = "failed"
	Succeeded Status = "succeeded"
	Cancelled Status = "cancelled"
)

// Run représente une exécution d'un projet ETL.
type Run struct {
	ID        string     `db:"id"         json:"id"`
	ProjectID string     `db:"project_id" json:"projectId"`
	OrgID     string     `db:"org_id"     json:"orgId"`     // Phase 11 multi-tenant
	Status    Status     `db:"status"     json:"status"`
	StartedAt time.Time  `db:"started_at" json:"startedAt"`
	EndedAt   *time.Time `db:"ended_at"   json:"endedAt,omitempty"`
	Error     string     `db:"error"      json:"error,omitempty"`
}

// RunLogEntry représente une entrée de log structurée pour un run.
// Ajouté Phase 10 — requis par le streaming de logs via WebSocket.
type RunLogEntry struct {
	RunID     string    `db:"run_id"    json:"runId"`
	BlockID   string    `db:"block_id"  json:"blockId"`
	Level     string    `db:"level"     json:"level"`   // "info" | "warn" | "error"
	Message   string    `db:"message"   json:"message"`
	Timestamp time.Time `db:"ts"        json:"timestamp"`
}

// Repository définit les opérations de persistance des runs.
// Phase 7 : Create, SetStatus, GetByID, ListByProject
// Phase 7 : GetLogs (appelé dans RunHandler.GetLogs — audit rupture #2)
// Phase 10 : AppendLog (streaming de logs structurés)
type Repository interface {
	Create(ctx context.Context, projectID string) (*Run, error)
	SetStatus(ctx context.Context, runID string, status Status) error
	GetByID(ctx context.Context, runID string) (*Run, error)
	ListByProject(ctx context.Context, projectID string) ([]Run, error)
	GetLogs(ctx context.Context, runID string) ([]RunLogEntry, error)
	AppendLog(ctx context.Context, entry RunLogEntry) error
}
