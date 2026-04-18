package models

import "time"

type RunStatus string

const (
	RunPending   RunStatus = "pending"
	RunRunning   RunStatus = "running"
	RunSucceeded RunStatus = "succeeded"
	RunFailed    RunStatus = "failed"
	RunCancelled RunStatus = "cancelled"
)

type Run struct {
	ID          string    `json:"id"`
	PipelineID  string    `json:"pipeline_id"`
	Status      RunStatus `json:"status"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
	ErrorMsg    string    `json:"error_msg,omitempty"`
	RecordsRead int64     `json:"records_read"`
	RecordsLoad int64     `json:"records_loaded"`
	CreatedAt   time.Time `json:"created_at"`
}
