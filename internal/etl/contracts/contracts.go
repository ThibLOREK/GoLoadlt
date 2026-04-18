package contracts

import "context"

type Record map[string]any

type Extractor interface {
	Extract(ctx context.Context) ([]Record, error)
}

type Transformer interface {
	Transform(ctx context.Context, in []Record) ([]Record, error)
}

type Loader interface {
	Load(ctx context.Context, in []Record) error
}
