package models

import (
	"encoding/json"
	"time"
)

type Pipeline struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Status       string          `json:"status"`
	SourceType   string          `json:"source_type"`
	TargetType   string          `json:"target_type"`
	SourceConfig json.RawMessage `json:"source_config,omitempty"`
	TargetConfig json.RawMessage `json:"target_config,omitempty"`
	Steps        json.RawMessage `json:"steps,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
