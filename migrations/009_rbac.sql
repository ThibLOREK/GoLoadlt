-- 009_rbac.sql
-- Phase 11 Axe 5 : RBAC avancé
-- Audit correction : renommé de 006 → 009 pour éviter le conflit avec 006_schedules_full.sql
DO $$ BEGIN
    CREATE TYPE user_role AS ENUM ('admin', 'editor', 'runner', 'viewer', 'conn_admin');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

ALTER TABLE users ADD COLUMN IF NOT EXISTS role user_role NOT NULL DEFAULT 'viewer';

-- Permissions granulaires par ressource (niveau avancé Phase 11)
CREATE TABLE IF NOT EXISTS resource_permissions (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id)         ON DELETE CASCADE,
    org_id      UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    resource    VARCHAR(50) NOT NULL,    -- 'project' | 'connection'
    resource_id UUID,                   -- NULL = toutes les ressources du type
    permission  VARCHAR(50) NOT NULL,   -- ex: 'project:run', 'connection:write'
    granted_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, org_id, resource, resource_id, permission)
);

CREATE INDEX IF NOT EXISTS idx_resource_perms_user ON resource_permissions(user_id, org_id);
