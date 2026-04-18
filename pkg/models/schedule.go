package models

import "time"

type Schedule struct {
	ID         string    `json:"id"`
	PipelineID string    `json:"pipeline_id"`
	CronExpr   string    `json:"cron_expr"`
	Enabled    bool      `json:"enabled"`
	LastRunAt  *time.Time `json:"last_run_at,omitempty"`
	NextRunAt  *time.Time `json:"next_run_at,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
