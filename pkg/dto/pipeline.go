package dto

type CreatePipelineRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SourceType  string `json:"source_type"`
	TargetType  string `json:"target_type"`
}

type UpdatePipelineRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	SourceType  string `json:"source_type"`
	TargetType  string `json:"target_type"`
}

type PipelineResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	SourceType  string `json:"source_type"`
	TargetType  string `json:"target_type"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
