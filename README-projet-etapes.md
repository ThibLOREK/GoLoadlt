# GoLoadIt — Architecture et feuille de route

## Vision
Construire une plateforme ETL visuelle en Go, inspirée d'Alteryx pour le design par blocs
et de Talend pour la gestion des connexions multi-environnements. Chaque projet ETL est
modélisé comme un graphe orienté de blocs fonctionnels, persisté en XML côté serveur,
exécutable de façon autonome par un worker Go.

## Objectifs produit
- Designer des flux ETL via une UI orientée blocs interconnectés (style Alteryx)
- Persister chaque projet sous forme de fichier XML versionné côté serveur (style Jenkins)
- Gérer des connexions réutilisables entre projets avec profils Dev / Préprod / Prod (style Talend)
- Exécuter les pipelines de façon fiable, traçable et reprenante
- Supporter de nombreux blocs de transformation : split, pivot, dépivot, colonne calculée, jointure...
- Permettre l'observabilité, la reprise sur erreur et la planification cron

## Concepts fondamentaux

### Projet ETL
Un projet est un graphe acyclique dirigé (DAG) composé de **blocs** (nodes) reliés par des
**liens** (edges). Chaque bloc encapsule une opération atomique : lire une source, transformer
des données, écrire vers une cible.

Le projet est sauvegardé en XML sur le serveur à chaque modification depuis l'UI, sur le modèle
des jobs Jenkins. Ce fichier XML est la source de vérité : il peut être versionné, importé,
exporté ou rejoué indépendamment de la base de métadonnées.

### Connexion réutilisable multi-env
Une connexion (base de données, API, fichier distant) est définie une fois et disponible dans
tous les projets. Elle embarque plusieurs profils d'environnement (Dev, Préprod, Prod).
Un switch global d'environnement permet de basculer tous les projets simultanément sans
modifier leur définition XML — exactement comme le contexte d'environnement dans Talend.

### Blocs de transformation
Chaque bloc expose :
- un **type** unique (`transform.pivot`, `transform.split`...)
- des **paramètres** configurables depuis l'UI
- un ou plusieurs **ports d'entrée** et **ports de sortie** (le bloc `split` a 1 entrée et N sorties)
- un **contrat de données** : schéma entrant / sortant

## Architecture cible

### Vue d'ensemble
Le projet est découpé en 4 couches :

1. **Presentation layer** : UI React Flow, API HTTP, WebSocket/SSE temps réel
2. **Application layer** : orchestration des jobs, gestion des connexions, switch env
3. **Domain ETL layer** : moteur DAG, catalogue de blocs, parser XML, contrats
4. **Infrastructure layer** : connecteurs, stockage XML/DB, logs, sécurité

### Structure projet
```text
GoLoadIt/
├── cmd/
│   ├── server/                  # Point d'entrée API + UI + supervision
│   ├── worker/                  # Exécuteur de projets ETL asynchrones
│   └── migrate/                 # Migrations de la DB de métadonnées
├── internal/
│   ├── app/                     # Bootstrap et wiring
│   ├── config/                  # Configuration multi-env (yaml + env vars)
│   ├── logger/                  # Logs structurés (zerolog)
│   ├── telemetry/               # Metrics, traces, healthchecks
│   ├── errors/                  # Erreurs métier et techniques
│   ├── etl/
│   │   ├── project/             # Modèle de projet ETL (DAG de blocs)
│   │   ├── blocks/              # Catalogue et registre de tous les blocs
│   │   │   ├── sources/         # source.csv, source.postgres, source.mssql...
│   │   │   ├── transforms/      # filter, select, cast, add_column, split,
│   │   │   │                    # pivot, unpivot, join, dedup, sort, aggregate
│   │   │   └── targets/         # target.postgres, target.csv, target.rest...
│   │   ├── engine/              # Moteur d'exécution DAG (parcours topologique)
│   │   ├── contracts/           # Interfaces Block, Port, Schema, DataRow
│   │   ├── scheduler/           # Planification cron
│   │   └── validation/          # Validation de graphe et de schéma
│   ├── xml/
│   │   ├── parser/              # XML → DAG de blocs en mémoire
│   │   ├── serializer/          # DAG → XML persisté
│   │   └── store/               # Stockage fichiers XML projets et connexions
│   ├── connections/
│   │   ├── manager/             # CRUD connexions, switch d'environnement global
│   │   ├── resolver/            # Résolution env actif → paramètres de connexion
│   │   └── secrets/             # Intégration vault / env vars pour les credentials
│   ├── orchestrator/            # Gestion globale des exécutions
│   ├── jobs/                    # États : pending, running, failed, succeeded
│   ├── security/                # AuthN, AuthZ, gestion des secrets
│   └── storage/                 # Accès PostgreSQL métadonnées, Redis cache
├── pkg/
│   ├── models/                  # Types partagés (Project, Block, Edge, Connection)
│   ├── dto/                     # Contrats API (JSON in/out)
│   └── utils/                   # Helpers : expression evaluator, type caster...
├── api/
│   ├── openapi/                 # Contrat OpenAPI
│   ├── handlers/                # Handlers HTTP
│   └── middleware/              # Auth, CORS, logging
├── web/
│   ├── ui/                      # Frontend React + TypeScript + React Flow
│   └── assets/
├── projects/                    # Fichiers XML des projets ETL (généré par le serveur)
│   └── proj-001/
│       ├── project.xml          # Définition du graphe de blocs
│       └── history/             # Versions précédentes (versionning XML)
├── connections/                 # Fichiers XML des connexions réutilisables
│   └── conn-crm.xml
├── deploy/
│   ├── docker/
│   └── k8s/
├── configs/
│   ├── config.dev.yaml
│   ├── config.preprod.yaml
│   └── config.prod.yaml
├── migrations/
├── scripts/
├── tests/
│   ├── integration/
│   ├── e2e/
│   └── fixtures/
└── docs/
    ├── adr/
    ├── architecture/
    ├── api/
    └── runbooks/
```

## Modèle de données clé

### Block (nœud du graphe)
```go
type Block struct {
    ID         string            `xml:"id,attr"`
    Type       string            `xml:"type,attr"`  // ex: "transform.pivot"
    Label      string            `xml:"label,attr"`
    Params     map[string]string `xml:"params>param"`
    ConnRef    string            `xml:"connectionRef,attr,omitempty"`
    InputPorts []Port
    OutputPorts []Port
}

type Port struct {
    ID     string
    Schema Schema // colonnes et types attendus
}
```

### Connection (multi-env)
```go
type Connection struct {
    ID       string                 `xml:"id,attr"`
    Name     string                 `xml:"name,attr"`
    Type     string                 `xml:"type,attr"` // postgres, mysql, mssql, rest...
    Envs     map[string]ConnEnv     `xml:"environments>env"`
}

type ConnEnv struct {
    Name      string `xml:"name,attr"`
    Host      string `xml:"host,attr"`
    Port      int    `xml:"port,attr"`
    Database  string `xml:"db,attr"`
    User      string `xml:"user,attr"`
    SecretRef string `xml:"secretRef,attr"` // référence vault ou env var
}
```

## Catalogue des blocs MVP

### Sources
| Bloc | Type | Paramètres clés |
|---|---|---|
| CSV | `source.csv` | path, delimiter, encoding, hasHeader |
| PostgreSQL | `source.postgres` | connectionRef, query, params |
| MySQL | `source.mysql` | connectionRef, query |
| SQL Server | `source.mssql` | connectionRef, query |
| API REST | `source.rest` | url, method, headers, pagination |

### Transformations
| Bloc | Type | Paramètres clés |
|---|---|---|
| Filtre | `transform.filter` | condition (ex: `amount > 100`) |
| Sélection colonnes | `transform.select` | columns (liste + renommage) |
| Cast type | `transform.cast` | column, targetType |
| Colonne calculée | `transform.add_column` | name, expression (ex: `price * qty`) |
| Split | `transform.split` | conditions[] → 1 port de sortie par condition |
| Pivot | `transform.pivot` | groupBy, valueColumn, aggregation |
| Dépivot | `transform.unpivot` | columns[], keyName, valueName |
| Jointure | `transform.join` | type (inner/left/right), leftKey, rightKey |
| Déduplication | `transform.dedup` | keys[] |
| Tri | `transform.sort` | columns[], order (asc/desc) |
| Agrégation | `transform.aggregate` | groupBy[], aggregations[] |

### Cibles
| Bloc | Type | Paramètres clés |
|---|---|---|
| PostgreSQL | `target.postgres` | connectionRef, table, mode (insert/upsert/truncate) |
| CSV | `target.csv` | path, delimiter, append |
| API REST | `target.rest` | url, method, bodyTemplate |

## Persistance XML des projets

Le répertoire `projects/` sur le serveur joue le même rôle que `$JENKINS_HOME/jobs/`.
À chaque sauvegarde depuis l'UI :
1. Le serveur sérialise le DAG en XML
2. L'ancienne version est archivée dans `projects/{id}/history/v{n}.xml`
3. La nouvelle version remplace `projects/{id}/project.xml`
4. Le hash SHA256 du fichier est stocké en base pour détecter toute modification externe

Le worker charge et parse le XML pour construire le DAG en mémoire avant exécution.
Aucune logique métier n'est stockée en base : la base ne contient que les métadonnées
d'exécution (runs, logs, statuts) et les références aux fichiers XML.

## Gestion des connexions multi-environnements

### Principe
Inspiré du **contexte Talend** :
- Toutes les connexions sont définies dans `connections/*.xml`
- Chaque connexion a un profil par environnement
- Un paramètre global `ACTIVE_ENV` (Dev / Préprod / Prod) détermine quel profil utiliser
- Les credentials (mots de passe) ne sont jamais écrits en clair : on utilise des références
  à des variables d'environnement ou à un vault (ex: `${DB_PASSWORD}` ou `vault:secret/crm`)

### Switch d'environnement

PUT /api/v1/environment
{ "env": "prod" }

Ce switch est global et immédiat : tous les projets utilisent désormais les paramètres de
connexion de production, sans modifier un seul fichier XML de projet.

### Résolution à l'exécution

connection "conn-crm" + ACTIVE_ENV="prod"
→ host: prod-db.internal, db: crm_prod, user: prod_user
→ password: résolu depuis vault ou env var au moment du run


## Principes d'architecture

- **Modularité des blocs** : tout nouveau bloc s'enregistre dans le catalogue sans toucher au moteur
- **XML comme source de vérité** : le projet est portable, versionnable, importable/exportable
- **Connexions découplées des projets** : un projet référence une connexion par ID, jamais les credentials
- **Switch env sans recompilation** : le profil actif est résolu à l'exécution par le `resolver`
- **Idempotence** : une relance de run ne corrompt pas la cible
- **Observabilité native** : chaque bloc tracé individuellement (lignes lues, écrites, erreurs)
- **Testabilité** : chaque bloc est testable de façon isolée avec des fixtures

## Stack technique

### Backend Go
- Go 1.24+
- Router HTTP : `chi`
- DB métadonnées : PostgreSQL (`pgx` / `sqlx`)
- Évaluation d'expressions : `expr-lang/expr` pour les colonnes calculées et filtres
- Config multi-env : `viper` + yaml par environnement
- Logs : `zerolog`
- Observabilité : OpenTelemetry + Prometheus
- XML : `encoding/xml` natif Go

### Frontend visuel
- React + TypeScript
- **React Flow** : canvas de blocs interconnectés avec palette latérale
- UI kit : shadcn/ui ou Ant Design
- Communication temps réel : WebSocket (suivi d'exécution bloc par bloc)

### Déploiement
- Docker Compose en local (server + worker + postgres + redis)
- Kubernetes pour la production
- CI/CD : GitHub Actions (lint + tests + build multi-stage)

## Feuille de route

### Phase 0 — Cadrage ✅
- Définir la vision blocs + XML + multi-env
- Identifier le catalogue de blocs MVP
- Formaliser les modèles `Project`, `Block`, `Connection`
- Produire les ADR initiaux

### Phase 1 — Fondation technique
- Initialiser la structure du repo
- Config multi-env (dev/preprod/prod yaml)
- Logger, erreurs, healthchecks
- Docker Compose local
- PostgreSQL métadonnées + migrations

### Phase 2 — Modèle XML et persistance
- Définir le schéma XML des projets et des connexions
- Implémenter le serializer DAG → XML
- Implémenter le parser XML → DAG
- Gérer le versionnement des fichiers XML (`history/`)
- Exposer les endpoints import/export XML

### Phase 3 — Moteur d'exécution DAG
- Définir les interfaces `Block`, `Port`, `Schema`, `DataRow`
- Implémenter le registre de blocs (catalogue)
- Développer le moteur de parcours topologique
- Gérer le contexte, timeout, retry, annulation
- Tracer les métriques par bloc (lignes in/out, durée, erreurs)

### Phase 4 — Blocs sources et cibles MVP
- `source.csv`, `source.postgres`, `source.mssql`
- `target.postgres`, `target.csv`
- Tests d'intégration par connecteur

### Phase 5 — Blocs de transformation MVP
- `transform.filter`, `transform.select`, `transform.cast`
- `transform.add_column` avec évaluateur d'expressions
- `transform.split` (1 entrée → N sorties conditionnelles)
- `transform.pivot`, `transform.unpivot`
- `transform.join` (jointure de deux flux)
- `transform.dedup`, `transform.sort`, `transform.aggregate`

### Phase 6 — Gestionnaire de connexions multi-env
- CRUD connexions XML
- Résolution env actif → paramètres de connexion
- Intégration secrets (env vars, vault)
- Switch global d'environnement via API
- Test de connexion par profil

### Phase 7 — API de pilotage complète
- CRUD projets (avec sauvegarde XML automatique)
- CRUD connexions
- Lancement / annulation de runs
- Historique et logs d'exécution
- Switch d'environnement global
- Documentation OpenAPI

### Phase 8 — Interface visuelle MVP
- Authentification
- Palette de blocs (drag & drop vers le canvas)
- Canvas React Flow : blocs interconnectés, configuration au clic
- Gestion des connexions avec profils d'environnement
- Exécution temps réel avec suivi bloc par bloc (WebSocket)
- Historique des runs et logs

### Phase 9 — Orchestration et scheduling
- Planification cron par projet
- Worker asynchrone avec file d'attente
- Retry policy configurable par projet
- Limitation de concurrence

### Phase 10 — Qualité et exploitation
- Tests e2e sur les blocs clés
- Profiling mémoire (gros volumes)
- Observabilité complète (traces par run)
- Versionnement et rollback de projet XML
- Runbooks d'exploitation

### Phase 11 — Évolutions avancées
- Blocs DAG multi-branches (fork / merge)
- CDC (Change Data Capture)
- Templates de projets réutilisables
- Multi-tenant
- RBAC avancé par projet / connexion

## MVP recommandé
| Composant | Contenu MVP |
|---|---|
| UI | Canvas React Flow + palette 10 blocs essentiels |
| Backend | API Go avec CRUD projets, connexions, runs |
| Worker | Exécuteur XML → DAG → run |
| Persistance | XML projets + XML connexions + PostgreSQL métadonnées |
| Blocs sources | CSV + PostgreSQL |
| Blocs transforms | filter, select, cast, add_column, split |
| Blocs cibles | PostgreSQL + CSV |
| Connexions | Multi-env Dev/Préprod/Prod + switch global |
| Scheduling | Exécution manuelle + cron simple |
| Observabilité | Logs par bloc, statuts de run, historique |

## Livrables par sprint
- **Sprint 1** : structure repo + Docker Compose + métadonnées DB
- **Sprint 2** : modèle XML (parser + serializer) + API projets/connexions
- **Sprint 3** : moteur DAG + blocs sources/cibles CSV et PostgreSQL
- **Sprint 4** : blocs de transformation MVP (filter, select, cast, add_column, split, pivot)
- **Sprint 5** : gestionnaire connexions multi-env + switch environnement
- **Sprint 6** : UI visuelle MVP (canvas React Flow + palette + configuration)
- **Sprint 7** : scheduling, worker asynchrone, retries, logs temps réel

## Risques à anticiper
- Complexité prématurée du designer visuel (commencer par un canvas minimaliste)
- Couplage entre l'UI et la structure XML (passer par un DTO intermédiaire)
- Gestion des flux multi-sorties du bloc `split` dans le moteur DAG
- Sécurité des credentials dans les fichiers XML (ne jamais écrire en clair)
- Performance sur les gros volumes (pipeline en streaming plutôt qu'en batch mémoire)
- Versionnement XML non maîtrisé si plusieurs utilisateurs éditent simultanément

## Recommandation d'expert
Séparer très tôt **le moteur ETL** (qui lit le XML et exécute les blocs), **l'orchestrateur**
(qui gère les runs et le scheduling) et **la couche UI/API** (qui édite et sauvegarde le XML).
Le fichier XML est le contrat entre l'UI et le moteur : ni l'un ni l'autre ne doit contenir
de logique cachée qui n'y soit pas représentée.

Commencer par faire tourner un projet XML simple en ligne de commande avant de construire
l'UI : cela valide le moteur indépendamment et permet des tests rapides.