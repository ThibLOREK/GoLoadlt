-- 006_schedules_full.sql
-- Phase 9 : table schedules complète avec expression cron et champs retry
CREATE TABLE IF NOT EXISTS schedules (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID        NOT NULL UNIQUE REFERENCES projects(id) ON DELETE CASCADE,
    cron_expr   VARCHAR(100) NOT NULL,              -- ex: "0 2 * * *"
    enabled     BOOLEAN      NOT NULL DEFAULT TRUE,
    timezone    VARCHAR(50)  NOT NULL DEFAULT 'UTC',
    last_run_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_schedules_next_run ON schedules(next_run_at) WHERE enabled = TRUE;
