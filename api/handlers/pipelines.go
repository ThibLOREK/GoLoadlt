package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/rinjold/go-etl-studio/pkg/dto"
)

func ListPipelines(w http.ResponseWriter, r *http.Request) {
	response := []dto.PipelineResponse{}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func CreatePipeline(w http.ResponseWriter, r *http.Request) {
	var request dto.CreatePipelineRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(dto.PipelineResponse{
		ID:          "todo-id",
		Name:        request.Name,
		Description: request.Description,
		Status:      "draft",
	})
}

func RunPipeline(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"message":"pipeline run scheduled"}`))
}
