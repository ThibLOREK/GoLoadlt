-- Migration 002 : connexions réutilisables et switch d'environnement

-- Table des connexions référencées dans les projets ETL.
-- Les paramètres détaillés par env sont dans les fichiers XML connections/.
CREATE TABLE IF NOT EXISTS connections (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    type        VARCHAR(64)  NOT NULL,  -- postgres, mysql, mssql, rest, csv
    xml_path    TEXT         NOT NULL,  -- chemin vers le fichier XML de la connexion
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- Table du contexte d'environnement global.
-- Une seule ligne : l'environnement actif pour toute la plateforme.
CREATE TABLE IF NOT EXISTS environment_context (
    id          SMALLINT PRIMARY KEY DEFAULT 1,  -- toujours 1 (singleton)
    active_env  VARCHAR(32)  NOT NULL DEFAULT 'dev',  -- dev | preprod | prod
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_by  VARCHAR(128),
    CONSTRAINT chk_single_row CHECK (id = 1)
);

-- Insérer la ligne singleton si elle n'existe pas.
INSERT INTO environment_context (id, active_env)
VALUES (1, 'dev')
ON CONFLICT (id) DO NOTHING;

-- Historique des changements d'environnement.
CREATE TABLE IF NOT EXISTS environment_history (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    previous_env VARCHAR(32) NOT NULL,
    new_env      VARCHAR(32) NOT NULL,
    changed_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    changed_by   VARCHAR(128)
);

-- Table des projets ETL (métadonnées, le graphe est dans le XML).
CREATE TABLE IF NOT EXISTS projects (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    xml_path    TEXT         NOT NULL,  -- chemin vers projects/{id}/project.xml
    xml_sha256  VARCHAR(64),            -- hash du fichier XML (intégrité)
    version     INT          NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- Index pour accélérer les recherches par nom.
CREATE INDEX IF NOT EXISTS idx_projects_name ON projects(name);
CREATE INDEX IF NOT EXISTS idx_connections_name ON connections(name);
