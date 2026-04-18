package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rinjold/go-etl-studio/pkg/models"
)

var ErrPipelineNotFound = errors.New("pipeline not found")

type PipelineRepository struct {
	pool *pgxpool.Pool
}

func NewPipelineRepository(pool *pgxpool.Pool) *PipelineRepository {
	return &PipelineRepository{pool: pool}
}

func (r *PipelineRepository) List(ctx context.Context) ([]models.Pipeline, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, description, status, source_type, target_type, created_at, updated_at
		FROM pipelines
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pipelines := make([]models.Pipeline, 0)
	for rows.Next() {
		var p models.Pipeline
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Status, &p.SourceType, &p.TargetType, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		pipelines = append(pipelines, p)
	}

	return pipelines, rows.Err()
}

func (r *PipelineRepository) GetByID(ctx context.Context, id string) (models.Pipeline, error) {
	var p models.Pipeline
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, description, status, source_type, target_type, created_at, updated_at
		FROM pipelines
		WHERE id = $1
	`, id).Scan(&p.ID, &p.Name, &p.Description, &p.Status, &p.SourceType, &p.TargetType, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return models.Pipeline{}, mapPGError(err)
	}

	return p, nil
}

func (r *PipelineRepository) Create(ctx context.Context, p models.Pipeline) (models.Pipeline, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO pipelines (id, name, description, status, source_type, target_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`, p.ID, p.Name, p.Description, p.Status, p.SourceType, p.TargetType).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return models.Pipeline{}, err
	}

	return p, nil
}

func (r *PipelineRepository) Update(ctx context.Context, p models.Pipeline) (models.Pipeline, error) {
	err := r.pool.QueryRow(ctx, `
		UPDATE pipelines
		SET name = $2,
		    description = $3,
		    status = $4,
		    source_type = $5,
		    target_type = $6,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING created_at, updated_at
	`, p.ID, p.Name, p.Description, p.Status, p.SourceType, p.TargetType).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return models.Pipeline{}, mapPGError(err)
	}

	return p, nil
}

func (r *PipelineRepository) Delete(ctx context.Context, id string) error {
	cmd, err := r.pool.Exec(ctx, `DELETE FROM pipelines WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrPipelineNotFound
	}
	return nil
}
