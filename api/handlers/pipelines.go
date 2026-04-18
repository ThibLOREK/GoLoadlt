package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rinjold/go-etl-studio/internal/services"
	"github.com/rinjold/go-etl-studio/internal/storage"
	"github.com/rinjold/go-etl-studio/pkg/dto"
	"github.com/rinjold/go-etl-studio/pkg/models"
)

func ListPipelines(service *services.PipelineService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pipelines, err := service.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := make([]dto.PipelineResponse, 0, len(pipelines))
		for _, pipeline := range pipelines {
			response = append(response, toPipelineResponse(pipeline))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}

func GetPipeline(service *services.PipelineService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "pipelineID")
		pipeline, err := service.GetByID(r.Context(), id)
		if err != nil {
			writePipelineError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(toPipelineResponse(pipeline))
	}
}

func CreatePipeline(service *services.PipelineService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request dto.CreatePipelineRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		pipeline, err := service.Create(r.Context(), services.CreatePipelineInput{
			Name:        request.Name,
			Description: request.Description,
			SourceType:  request.SourceType,
			TargetType:  request.TargetType,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(toPipelineResponse(pipeline))
	}
}

func UpdatePipeline(service *services.PipelineService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "pipelineID")
		var request dto.UpdatePipelineRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		pipeline, err := service.Update(r.Context(), id, services.UpdatePipelineInput{
			Name:        request.Name,
			Description: request.Description,
			Status:      request.Status,
			SourceType:  request.SourceType,
			TargetType:  request.TargetType,
		})
		if err != nil {
			writePipelineError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(toPipelineResponse(pipeline))
	}
}

func DeletePipeline(service *services.PipelineService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "pipelineID")
		if err := service.Delete(r.Context(), id); err != nil {
			writePipelineError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func RunPipeline(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"message":"pipeline run scheduled"}`))
}

func toPipelineResponse(p models.Pipeline) dto.PipelineResponse {
	return dto.PipelineResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Status:      p.Status,
		SourceType:  p.SourceType,
		TargetType:  p.TargetType,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}

func writePipelineError(w http.ResponseWriter, err error) {
	if errors.Is(err, storage.ErrPipelineNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
