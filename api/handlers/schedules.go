package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rinjold/go-etl-studio/internal/services"
	"github.com/rinjold/go-etl-studio/internal/storage"
)

type UpsertScheduleRequest struct {
	CronExpr string `json:"cron_expr"`
	Enabled  bool   `json:"enabled"`
}

func UpsertSchedule(svc *services.ScheduleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pipelineID := chi.URLParam(r, "pipelineID")
		var req UpsertScheduleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		sched, err := svc.Upsert(r.Context(), pipelineID, req.CronExpr, req.Enabled)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sched)
	}
}

func GetSchedule(svc *services.ScheduleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pipelineID := chi.URLParam(r, "pipelineID")
		sched, err := svc.GetByPipeline(r.Context(), pipelineID)
		if err != nil {
			if err == storage.ErrScheduleNotFound {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sched)
	}
}

func DeleteSchedule(svc *services.ScheduleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pipelineID := chi.URLParam(r, "pipelineID")
		if err := svc.Delete(r.Context(), pipelineID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
