package pipeline

import "encoding/json"

type SourceType string
type TargetType string

const (
	SourceCSV      SourceType = "csv"
	SourcePostgres SourceType = "postgres"
	SourceAPI      SourceType = "api"

	TargetPostgres TargetType = "postgres"
	TargetCSV      TargetType = "csv"
)

type CSVSourceConfig struct {
	FilePath  string `json:"file_path"`
	Delimiter string `json:"delimiter"`
	HasHeader bool   `json:"has_header"`
}

type PostgresSourceConfig struct {
	DSN       string   `json:"dsn"`
	Schema    string   `json:"schema"`
	TableName string   `json:"table_name"`
	Columns   []string `json:"columns"`
	Where     string   `json:"where"`
}

type APISourceConfig struct {
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers"`
	DataPath       string            `json:"data_path"`
	Pagination     string            `json:"pagination"`
	PageParam      string            `json:"page_param"`
	LimitParam     string            `json:"limit_param"`
	PageSize       int               `json:"page_size"`
	MaxPages       int               `json:"max_pages"`
	TimeoutSeconds int               `json:"timeout_seconds"`
}

type PostgresTargetConfig struct {
	DSN       string   `json:"dsn"`
	Schema    string   `json:"schema"`
	TableName string   `json:"table_name"`
	Columns   []string `json:"columns"`
	BatchSize int      `json:"batch_size"`
}

type TransformStep struct {
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config"`
}

type MapperConfig struct {
	Mapping map[string]string `json:"mapping"`
}

type FilterConfig struct {
	Rules []FilterRule `json:"rules"`
}

type FilterRule struct {
	Column   string `json:"column"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type CastRule struct {
	Column string `json:"column"`
	CastTo string `json:"cast_to"`
}

type CasterConfig struct {
	Rules []CastRule `json:"rules"`
}

type Definition struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	SourceType   SourceType      `json:"source_type"`
	TargetType   TargetType      `json:"target_type"`
	SourceConfig json.RawMessage `json:"source_config"`
	TargetConfig json.RawMessage `json:"target_config"`
	Steps        []TransformStep `json:"steps"`
}
