package dto

type CreatePipelineRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PipelineResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}
