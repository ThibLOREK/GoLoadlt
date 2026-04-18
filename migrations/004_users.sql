CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'viewer',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed admin par défaut (password: admin123)
INSERT INTO users (id, email, password_hash, role)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'admin@etl.local',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    'admin'
) ON CONFLICT DO NOTHING;
