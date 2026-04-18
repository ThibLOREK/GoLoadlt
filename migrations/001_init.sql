CREATE TABLE IF NOT EXISTS pipelines (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    description   TEXT,
    status        TEXT NOT NULL DEFAULT 'draft',
    source_type   TEXT NOT NULL DEFAULT 'unknown',
    target_type   TEXT NOT NULL DEFAULT 'unknown',
    source_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    target_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    steps         JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
