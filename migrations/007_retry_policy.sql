-- 007_retry_policy.sql
-- Phase 9 : politiques de retry par projet
-- Audit correction : SQL tronqué dans ETAPE-9-DETAIL.md (parenthèse fermante manquante)
CREATE TABLE IF NOT EXISTS retry_policies (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id   UUID        NOT NULL UNIQUE REFERENCES projects(id) ON DELETE CASCADE,
    max_attempts INT         NOT NULL DEFAULT 1,
    delay_ms     BIGINT      NOT NULL DEFAULT 30000,
    backoff      NUMERIC(4,2) NOT NULL DEFAULT 2.0,
    max_delay_ms BIGINT      NOT NULL DEFAULT 600000,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
