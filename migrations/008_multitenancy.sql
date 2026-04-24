-- 008_multitenancy.sql
-- Phase 11 Axe 4 : Multi-tenant
-- Audit correction : renommé de 005 → 008 pour éviter le conflit avec 005_run_logs.sql
CREATE TABLE IF NOT EXISTS organizations (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    slug       VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    active     BOOLEAN      NOT NULL DEFAULT TRUE
);

-- Ajouter org_id sur toutes les tables existantes
ALTER TABLE projects ADD COLUMN IF NOT EXISTS org_id UUID REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE runs     ADD COLUMN IF NOT EXISTS org_id UUID REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE users    ADD COLUMN IF NOT EXISTS org_id UUID REFERENCES organizations(id) ON DELETE CASCADE;

-- Index pour toutes les requêtes filtrées par org
CREATE INDEX IF NOT EXISTS idx_projects_org ON projects(org_id);
CREATE INDEX IF NOT EXISTS idx_runs_org     ON runs(org_id);
CREATE INDEX IF NOT EXISTS idx_users_org    ON users(org_id);

-- Organisation par défaut pour les données existantes (migration non-breaking)
INSERT INTO organizations (id, name, slug)
    VALUES ('00000000-0000-0000-0000-000000000001', 'Default', 'default')
    ON CONFLICT DO NOTHING;

UPDATE projects SET org_id = '00000000-0000-0000-0000-000000000001' WHERE org_id IS NULL;
UPDATE runs     SET org_id = '00000000-0000-0000-0000-000000000001' WHERE org_id IS NULL;
UPDATE users    SET org_id = '00000000-0000-0000-0000-000000000001' WHERE org_id IS NULL;

ALTER TABLE projects ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE runs     ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE users    ALTER COLUMN org_id SET NOT NULL;
