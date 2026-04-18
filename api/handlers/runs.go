package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rinjold/go-etl-studio/internal/services"
	"github.com/rinjold/go-etl-studio/pkg/models"
)

func ScheduleRun(service *services.RunService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pipelineID := chi.URLParam(r, "pipelineID")

		run, err := service.Schedule(r.Context(), pipelineID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(toRunResponse(run))
	}
}

func GetRun(service *services.RunService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		runID := chi.URLParam(r, "runID")

		run, err := service.GetByID(r.Context(), runID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(toRunResponse(run))
	}
}

func ListRuns(service *services.RunService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pipelineID := chi.URLParam(r, "pipelineID")

		runs, err := service.ListByPipeline(r.Context(), pipelineID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := make([]map[string]any, 0, len(runs))
		for _, run := range runs {
			response = append(response, toRunResponse(run))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}

func toRunResponse(run models.Run) map[string]any {
	return map[string]any{
		"id":             run.ID,
		"pipeline_id":    run.PipelineID,
		"status":         run.Status,
		"started_at":     run.StartedAt,
		"finished_at":    run.FinishedAt,
		"error_msg":      run.ErrorMsg,
		"records_read":   run.RecordsRead,
		"records_loaded": run.RecordsLoad,
		"created_at":     run.CreatedAt,
	}
}
