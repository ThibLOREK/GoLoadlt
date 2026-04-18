package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rinjold/go-etl-studio/pkg/models"
)

var ErrRunNotFound = errors.New("run not found")

type RunRepository struct {
	pool *pgxpool.Pool
}

func NewRunRepository(pool *pgxpool.Pool) *RunRepository {
	return &RunRepository{pool: pool}
}

func (r *RunRepository) Create(ctx context.Context, run models.Run) (models.Run, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO runs (id, pipeline_id, status)
		VALUES ($1, $2, $3)
		RETURNING created_at
	`, run.ID, run.PipelineID, run.Status).Scan(&run.CreatedAt)
	return run, err
}

func (r *RunRepository) GetByID(ctx context.Context, id string) (models.Run, error) {
	var run models.Run
	err := r.pool.QueryRow(ctx, `
		SELECT id, pipeline_id, status, started_at, finished_at, error_msg,
		       records_read, records_loaded, created_at
		FROM runs WHERE id = $1
	`, id).Scan(
		&run.ID, &run.PipelineID, &run.Status,
		&run.StartedAt, &run.FinishedAt, &run.ErrorMsg,
		&run.RecordsRead, &run.RecordsLoad, &run.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return run, ErrRunNotFound
	}
	return run, err
}

func (r *RunRepository) ListByPipeline(ctx context.Context, pipelineID string) ([]models.Run, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, pipeline_id, status, started_at, finished_at, error_msg,
		       records_read, records_loaded, created_at
		FROM runs WHERE pipeline_id = $1
		ORDER BY created_at DESC
	`, pipelineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := make([]models.Run, 0)
	for rows.Next() {
		var run models.Run
		if err := rows.Scan(
			&run.ID, &run.PipelineID, &run.Status,
			&run.StartedAt, &run.FinishedAt, &run.ErrorMsg,
			&run.RecordsRead, &run.RecordsLoad, &run.CreatedAt,
		); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func (r *RunRepository) UpdateStatus(ctx context.Context, id string, status models.RunStatus, errMsg string) error {
	now := time.Now().UTC()
	var q string
	var args []any

	switch status {
	case models.RunRunning:
		q = `UPDATE runs SET status = $2, started_at = $3 WHERE id = $1`
		args = []any{id, status, now}
	case models.RunSucceeded, models.RunFailed, models.RunCancelled:
		q = `UPDATE runs SET status = $2, finished_at = $3, error_msg = $4 WHERE id = $1`
		args = []any{id, status, now, errMsg}
	default:
		q = `UPDATE runs SET status = $2 WHERE id = $1`
		args = []any{id, status}
	}

	_, err := r.pool.Exec(ctx, q, args...)
	return err
}

func (r *RunRepository) UpdateCounts(ctx context.Context, id string, read, loaded int64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE runs SET records_read = $2, records_loaded = $3 WHERE id = $1`,
		id, read, loaded,
	)
	return err
}
