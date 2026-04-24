-- 005_run_logs.sql
-- Ajouté Phase 10 : logs structurés par bloc pour le streaming WebSocket
CREATE TABLE IF NOT EXISTS run_logs (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id     UUID        NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    block_id   VARCHAR(255) NOT NULL,
    level      VARCHAR(10)  NOT NULL DEFAULT 'info', -- 'info' | 'warn' | 'error'
    message    TEXT         NOT NULL,
    ts         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_run_logs_run_id ON run_logs(run_id);
CREATE INDEX IF NOT EXISTS idx_run_logs_ts     ON run_logs(run_id, ts);
