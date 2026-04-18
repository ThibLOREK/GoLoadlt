package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

type PaginationMode string

const (
	PaginationNone   PaginationMode = "none"
	PaginationOffset PaginationMode = "offset"
	PaginationCursor PaginationMode = "cursor"
)

type ExtractorConfig struct {
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers"`
	DataPath       string            `json:"data_path"`   // ex: "data", "results", ""
	Pagination     PaginationMode    `json:"pagination"`
	PageParam      string            `json:"page_param"`  // ex: "page", "offset"
	LimitParam     string            `json:"limit_param"` // ex: "limit", "per_page"
	PageSize       int               `json:"page_size"`
	MaxPages       int               `json:"max_pages"`
	TimeoutSeconds int               `json:"timeout_seconds"`
}

type Extractor struct {
	cfg    ExtractorConfig
	client *http.Client
}

func NewExtractor(cfg ExtractorConfig) *Extractor {
	if cfg.Method == "" {
		cfg.Method = http.MethodGet
	}
	if cfg.PageSize == 0 {
		cfg.PageSize = 100
	}
	if cfg.MaxPages == 0 {
		cfg.MaxPages = 100
	}
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &Extractor{
		cfg:    cfg,
		client: &http.Client{Timeout: timeout},
	}
}

func (e *Extractor) Extract(ctx context.Context) ([]contracts.Record, error) {
	switch e.cfg.Pagination {
	case PaginationOffset:
		return e.extractPaginated(ctx)
	default:
		return e.extractSingle(ctx, e.cfg.URL)
	}
}

func (e *Extractor) extractSingle(ctx context.Context, url string) ([]contracts.Record, error) {
	req, err := http.NewRequestWithContext(ctx, e.cfg.Method, url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range e.cfg.Headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseBody(body, e.cfg.DataPath)
}

func (e *Extractor) extractPaginated(ctx context.Context) ([]contracts.Record, error) {
	var all []contracts.Record
	offset := 0

	for page := 0; page < e.cfg.MaxPages; page++ {
		url := fmt.Sprintf("%s?%s=%d&%s=%d",
			e.cfg.URL,
			e.cfg.PageParam, offset,
			e.cfg.LimitParam, e.cfg.PageSize,
		)
		records, err := e.extractSingle(ctx, url)
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			break
		}
		all = append(all, records...)
		if len(records) < e.cfg.PageSize {
			break
		}
		offset += e.cfg.PageSize
	}
	return all, nil
}

func parseBody(body []byte, dataPath string) ([]contracts.Record, error) {
	var raw any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if dataPath != "" {
		obj, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected JSON object, got array or scalar")
		}
		raw = obj[dataPath]
	}

	switch v := raw.(type) {
	case []any:
		records := make([]contracts.Record, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				rec := make(contracts.Record, len(m))
				for k, val := range m {
					rec[k] = val
				}
				records = append(records, rec)
			}
		}
		return records, nil
	case map[string]any:
		rec := make(contracts.Record, len(v))
		for k, val := range v {
			rec[k] = val
		}
		return []contracts.Record{rec}, nil
	default:
		return nil, fmt.Errorf("unexpected JSON structure at path %q", dataPath)
	}
}
