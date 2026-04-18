package engine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	apiconn "github.com/rinjold/go-etl-studio/internal/connectors/api"
	csvconn "github.com/rinjold/go-etl-studio/internal/connectors/csv"
	pgconn "github.com/rinjold/go-etl-studio/internal/connectors/postgres"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
	"github.com/rinjold/go-etl-studio/internal/etl/pipeline"
	"github.com/rinjold/go-etl-studio/internal/etl/transformers"
	"github.com/rinjold/go-etl-studio/internal/storage"
	"github.com/rs/zerolog"
)

func BuildExecutor(ctx context.Context, def pipeline.Definition, pool *pgxpool.Pool, log zerolog.Logger) (Executor, error) {
	extractor, err := buildExtractor(ctx, def, pool)
	if err != nil {
		return Executor{}, fmt.Errorf("build extractor: %w", err)
	}
	loader, err := buildLoader(ctx, def, pool)
	if err != nil {
		return Executor{}, fmt.Errorf("build loader: %w", err)
	}
	chain, err := buildTransformChain(def.Steps)
	if err != nil {
		return Executor{}, fmt.Errorf("build transformers: %w", err)
	}
	return Executor{Extractor: extractor, Transformer: chain, Loader: loader, Log: log}, nil
}

func buildExtractor(ctx context.Context, def pipeline.Definition, pool *pgxpool.Pool) (contracts.Extractor, error) {
	switch def.SourceType {
	case pipeline.SourceCSV:
		var cfg csvconn.ExtractorConfig
		if err := json.Unmarshal(def.SourceConfig, &cfg); err != nil {
			return nil, err
		}
		return csvconn.NewExtractor(cfg), nil

	case pipeline.SourcePostgres:
		var cfg pgconn.ExtractorConfig
		if err := json.Unmarshal(def.SourceConfig, &cfg); err != nil {
			return nil, err
		}
		if cfg.DSN != "" {
			srcPool, err := storage.NewPostgresPool(ctx, cfg.DSN)
			if err != nil {
				return nil, err
			}
			cfg.Pool = srcPool
		} else {
			cfg.Pool = pool
		}
		return pgconn.NewExtractor(cfg), nil

	case pipeline.SourceAPI:
		var cfg apiconn.ExtractorConfig
		if err := json.Unmarshal(def.SourceConfig, &cfg); err != nil {
			return nil, err
		}
		return apiconn.NewExtractor(cfg), nil

	default:
		return nil, fmt.Errorf("unknown source type: %s", def.SourceType)
	}
}

func buildLoader(ctx context.Context, def pipeline.Definition, pool *pgxpool.Pool) (contracts.Loader, error) {
	switch def.TargetType {
	case pipeline.TargetPostgres:
		var cfg pgconn.LoaderConfig
		if err := json.Unmarshal(def.TargetConfig, &cfg); err != nil {
			return nil, err
		}
		if cfg.Pool == nil {
			cfg.Pool = pool
		}
		return pgconn.NewLoader(cfg), nil
	default:
		return nil, fmt.Errorf("unknown target type: %s", def.TargetType)
	}
}

func buildTransformChain(steps []pipeline.TransformStep) (contracts.Transformer, error) {
	if len(steps) == 0 {
		return transformers.Noop{}, nil
	}
	chain := transformers.Chain{}
	for _, step := range steps {
		t, err := buildTransformer(step)
		if err != nil {
			return nil, err
		}
		chain.Steps = append(chain.Steps, t)
	}
	return chain, nil
}

func buildTransformer(step pipeline.TransformStep) (contracts.Transformer, error) {
	switch step.Type {
	case "mapper":
		var cfg pipeline.MapperConfig
		if err := json.Unmarshal(step.Config, &cfg); err != nil {
			return nil, err
		}
		return transformers.Mapper{Mapping: cfg.Mapping}, nil
	case "filter":
		var cfg pipeline.FilterConfig
		if err := json.Unmarshal(step.Config, &cfg); err != nil {
			return nil, err
		}
		rules := make([]transformers.FilterRule, len(cfg.Rules))
		for i, r := range cfg.Rules {
			rules[i] = transformers.FilterRule{Column: r.Column, Operator: transformers.Operator(r.Operator), Value: r.Value}
		}
		return transformers.Filter{Rules: rules}, nil
	case "caster":
		var cfg pipeline.CasterConfig
		if err := json.Unmarshal(step.Config, &cfg); err != nil {
			return nil, err
		}
		rules := make([]transformers.CastRule, len(cfg.Rules))
		for i, r := range cfg.Rules {
			rules[i] = transformers.CastRule{Column: r.Column, CastTo: transformers.CastType(r.CastTo)}
		}
		return transformers.Caster{Rules: rules}, nil
	default:
		return nil, fmt.Errorf("unknown transformer type: %s", step.Type)
	}
}
