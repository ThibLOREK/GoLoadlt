package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rinjold/go-etl-studio/pkg/models"
)

var ErrScheduleNotFound = errors.New("schedule not found")

type ScheduleRepository struct {
	pool *pgxpool.Pool
}

func NewScheduleRepository(pool *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{pool: pool}
}

func (r *ScheduleRepository) Upsert(ctx context.Context, s models.Schedule) (models.Schedule, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO schedules (id, pipeline_id, cron_expr, enabled, next_run_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (pipeline_id) DO UPDATE
		SET cron_expr = EXCLUDED.cron_expr,
		    enabled = EXCLUDED.enabled,
		    next_run_at = EXCLUDED.next_run_at,
		    updated_at = NOW()
		RETURNING created_at, updated_at
	`, s.ID, s.PipelineID, s.CronExpr, s.Enabled, s.NextRunAt).Scan(&s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *ScheduleRepository) GetByPipeline(ctx context.Context, pipelineID string) (models.Schedule, error) {
	var s models.Schedule
	err := r.pool.QueryRow(ctx, `
		SELECT id, pipeline_id, cron_expr, enabled, last_run_at, next_run_at, created_at, updated_at
		FROM schedules WHERE pipeline_id = $1
	`, pipelineID).Scan(&s.ID, &s.PipelineID, &s.CronExpr, &s.Enabled,
		&s.LastRunAt, &s.NextRunAt, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return s, ErrScheduleNotFound
	}
	return s, err
}

func (r *ScheduleRepository) ListDue(ctx context.Context, now time.Time) ([]models.Schedule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, pipeline_id, cron_expr, enabled, last_run_at, next_run_at, created_at, updated_at
		FROM schedules
		WHERE enabled = TRUE AND next_run_at <= $1
		ORDER BY next_run_at ASC
	`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []models.Schedule
	for rows.Next() {
		var s models.Schedule
		if err := rows.Scan(&s.ID, &s.PipelineID, &s.CronExpr, &s.Enabled,
			&s.LastRunAt, &s.NextRunAt, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}
	return schedules, rows.Err()
}

func (r *ScheduleRepository) UpdateAfterFire(ctx context.Context, id string, lastRun time.Time, nextRun *time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE schedules SET last_run_at = $2, next_run_at = $3, updated_at = NOW()
		WHERE id = $1
	`, id, lastRun, nextRun)
	return err
}

func (r *ScheduleRepository) Delete(ctx context.Context, pipelineID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM schedules WHERE pipeline_id = $1`, pipelineID)
	return err
}
