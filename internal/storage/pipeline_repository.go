package storage

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
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
		SELECT id, name, description, status, source_type, target_type,
		       source_config, target_config, steps, created_at, updated_at
		FROM pipelines ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pipelines []models.Pipeline
	for rows.Next() {
		p, err := scanPipeline(rows.Scan)
		if err != nil {
			return nil, err
		}
		pipelines = append(pipelines, p)
	}
	return pipelines, rows.Err()
}

func (r *PipelineRepository) GetByID(ctx context.Context, id string) (models.Pipeline, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, description, status, source_type, target_type,
		       source_config, target_config, steps, created_at, updated_at
		FROM pipelines WHERE id = $1
	`, id)
	p, err := scanPipeline(row.Scan)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, ErrPipelineNotFound
	}
	return p, err
}

func (r *PipelineRepository) Create(ctx context.Context, p models.Pipeline) (models.Pipeline, error) {
	sc, _ := json.Marshal(p.SourceConfig)
	tc, _ := json.Marshal(p.TargetConfig)
	st, _ := json.Marshal(p.Steps)
	err := r.pool.QueryRow(ctx, `
		INSERT INTO pipelines (id, name, description, status, source_type, target_type,
		                       source_config, target_config, steps)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING created_at, updated_at
	`, p.ID, p.Name, p.Description, p.Status, p.SourceType, p.TargetType,
		sc, tc, st).Scan(&p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *PipelineRepository) Update(ctx context.Context, p models.Pipeline) (models.Pipeline, error) {
	sc, _ := json.Marshal(p.SourceConfig)
	tc, _ := json.Marshal(p.TargetConfig)
	st, _ := json.Marshal(p.Steps)
	err := r.pool.QueryRow(ctx, `
		UPDATE pipelines SET name=$2, description=$3, status=$4, source_type=$5,
		       target_type=$6, source_config=$7, target_config=$8, steps=$9, updated_at=NOW()
		WHERE id=$1 RETURNING created_at, updated_at
	`, p.ID, p.Name, p.Description, p.Status, p.SourceType, p.TargetType,
		sc, tc, st).Scan(&p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, ErrPipelineNotFound
	}
	return p, err
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

func scanPipeline(scan func(...any) error) (models.Pipeline, error) {
	var p models.Pipeline
	var sc, tc, st []byte
	err := scan(&p.ID, &p.Name, &p.Description, &p.Status,
		&p.SourceType, &p.TargetType, &sc, &tc, &st,
		&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return p, err
	}
	_ = json.Unmarshal(sc, &p.SourceConfig)
	_ = json.Unmarshal(tc, &p.TargetConfig)
	_ = json.Unmarshal(st, &p.Steps)
	return p, nil
}
