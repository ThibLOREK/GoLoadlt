# Projet ETL en Golang avec interface visuelle

## Vision
Construire une plateforme ETL modulaire en Go avec une interface visuelle permettant de concevoir, exécuter, superviser et historiser des pipelines de données.

## Objectifs produit
- Concevoir des flux ETL via une UI visuelle
- Exécuter des pipelines de façon fiable et traçable
- Isoler clairement extraction, transformation et chargement
- Supporter plusieurs connecteurs (DB, fichiers, API)
- Permettre l'observabilité, la reprise sur erreur et la planification
- Préparer une architecture extensible vers l'ELT, le streaming et le CDC

## Architecture cible

### Vue d'ensemble
Le projet est découpé en 4 couches principales :

1. **Presentation layer** : UI web, API HTTP, WebSocket/SSE pour le suivi temps réel
2. **Application layer** : orchestration des jobs, gestion des exécutions, règles métier
3. **Domain ETL layer** : moteur de pipeline, contrats extract/transform/load, validation
4. **Infrastructure layer** : connecteurs, stockage, logs, observabilité, sécurité, déploiement

### Structure proposée
```text
go-etl-studio/
├── cmd/
│   ├── server/                  # Point d'entrée API + UI + supervision
│   ├── worker/                  # Exécuteur de jobs ETL asynchrones
│   └── migrate/                 # Outil de migrations DB
├── internal/
│   ├── app/                     # Bootstrap applicatif et wiring
│   ├── config/                  # Chargement et validation de configuration
│   ├── logger/                  # Journalisation structurée
│   ├── telemetry/               # Metrics, traces, healthchecks
│   ├── errors/                  # Erreurs métier et techniques
│   ├── etl/
│   │   ├── pipeline/            # Définition des pipelines et DAG simple
│   │   ├── extractors/          # Implémentations des extracteurs
│   │   ├── transformers/        # Transformations unitaires et chaînables
│   │   ├── loaders/             # Implémentations des loaders
│   │   ├── contracts/           # Interfaces métier ETL
│   │   ├── engine/              # Moteur d'exécution d'un pipeline
│   │   ├── scheduler/           # Planification locale / cron
│   │   └── validation/          # Validation de schéma, règles, config
│   ├── connectors/
│   │   ├── sql/                 # Abstractions SQL communes
│   │   ├── postgres/
│   │   ├── mysql/
│   │   ├── mssql/
│   │   ├── csv/
│   │   ├── api/
│   │   └── s3/
│   ├── metadata/
│   │   ├── catalog/             # Métadonnées de pipeline et sources/cibles
│   │   ├── lineage/             # Traçabilité des flux et dépendances
│   │   └── repository/          # Persistance des définitions/exécutions
│   ├── orchestrator/            # Gestion globale des exécutions et dépendances
│   ├── jobs/                    # File de jobs, retries, états d'exécution
│   ├── security/                # AuthN, AuthZ, gestion secrets
│   └── storage/                 # Accès aux stores (Postgres, Redis, blob...)
├── pkg/
│   ├── models/                  # Types partageables hors internal
│   ├── dto/                     # Contrats API externes
│   └── utils/                   # Helpers génériques réutilisables
├── api/
│   ├── openapi/                 # Contrat OpenAPI
│   ├── handlers/                # Handlers HTTP
│   └── middleware/              # Middleware HTTP
├── web/
│   ├── ui/                      # Frontend visuel (React/Vue/Svelte conseillé)
│   └── assets/                  # Assets statiques
├── deploy/
│   ├── docker/                  # Dockerfiles / compose
│   └── k8s/                     # Manifests Kubernetes
├── configs/                     # Fichiers de configuration par environnement
├── migrations/                  # Schéma BDD applicative
├── scripts/                     # Scripts dev / CI / bootstrap
├── tests/
│   ├── integration/
│   ├── e2e/
│   └── fixtures/
└── docs/
    ├── adr/                     # Architecture Decision Records
    ├── architecture/            # Diagrammes et vues d'architecture
    ├── api/
    └── runbooks/
```

## Découpage des responsabilités

### `cmd/`
- `server` démarre l'API, expose l'UI, publie les endpoints d'administration et l'observabilité
- `worker` exécute les jobs ETL en arrière-plan
- `migrate` applique les migrations du stockage de métadonnées

### `internal/etl/`
C'est le coeur métier.
- `contracts` définit les interfaces `Extractor`, `Transformer`, `Loader`, `PipelineRunner`
- `engine` enchaîne les étapes, gère le contexte, les erreurs et les retries
- `pipeline` modélise le pipeline, ses noeuds, ses dépendances et ses paramètres
- `validation` contrôle les définitions envoyées par l'UI avant exécution

### `internal/connectors/`
Chaque connecteur encapsule :
- l'initialisation de la connexion
- la lecture/écriture
- le mapping de schémas
- la gestion des erreurs techniques
- l'optimisation spécifique (batch, pagination, bulk insert, transactions)

### `internal/orchestrator` et `internal/jobs`
- `orchestrator` décide quoi exécuter et quand
- `jobs` gère les états : `pending`, `running`, `failed`, `succeeded`, `cancelled`
- prévoir dès le départ la stratégie de retry, idempotence et reprise

### `web/ui`
L'interface visuelle doit offrir :
- un designer de pipeline orienté graphe/noeuds
- un écran de configuration des sources, cibles et transformations
- un tableau de bord d'exécution
- un écran de logs et diagnostics
- une gestion des versions de pipeline

## Principes d'architecture recommandés
- **Modularité** : chaque connecteur et chaque transformation doit être extensible sans toucher au coeur
- **Contrats clairs** : interfaces Go petites et stables
- **Séparation métier / infra** : ne pas mélanger logique ETL et détails HTTP/SQL
- **Idempotence** : une relance ne doit pas corrompre la cible
- **Observabilité native** : logs structurés, métriques, traces, audit
- **Testabilité** : composants injectables, mocks simples, tests d'intégration réels
- **Sécurité** : secrets externalisés, RBAC, chiffrement des données sensibles

## Stack technique recommandée

### Backend Go
- Go 1.24+
- Router HTTP : `chi` ou `gin` (préférence `chi` si tu veux une base sobre et testable)
- DB metadata : PostgreSQL
- Queue/cache : Redis ou NATS selon la complexité
- ORM/SQL : `sqlx` ou `pgx` côté PostgreSQL, éviter un ORM trop opaque au début
- Config : `viper` ou config maison stricte avec env + yaml
- Logs : `zap` ou `zerolog`
- Observabilité : OpenTelemetry + Prometheus

### Frontend visuel
- React + TypeScript
- Librairie de graphe : React Flow
- UI kit : MUI, Ant Design ou shadcn/ui
- Communication temps réel : WebSocket ou SSE

### Déploiement
- Docker Compose en local
- Kubernetes plus tard si besoin de scalabilité
- CI/CD avec tests + lint + build multi-stage

## Étapes de développement

### Phase 0 — Cadrage
- Définir le périmètre MVP
- Lister les connecteurs prioritaires
- Définir les types de transformations du MVP
- Formaliser les cas d'usage principaux
- Produire les ADR initiaux

### Phase 1 — Fondation technique
- Initialiser le monorepo / repo principal
- Mettre en place la structure projet
- Ajouter config, logger, gestion d'erreurs, healthchecks
- Préparer Docker Compose local
- Mettre en place la base PostgreSQL de métadonnées
- Ajouter migrations et seed minimal

### Phase 2 — Noyau moteur ETL
- Définir les interfaces métier ETL
- Implémenter le modèle de pipeline
- Développer le moteur d'exécution séquentiel
- Ajouter gestion de contexte, timeout, retry, annulation
- Tracer les états d'exécution et logs techniques

### Phase 3 — Connecteurs MVP
- Source CSV
- Source PostgreSQL
- Cible PostgreSQL
- Cible fichier CSV
- Source API REST simple
- Ajouter tests d'intégration par connecteur

### Phase 4 — Transformations MVP
- Mapping de colonnes
- Cast de types
- Filtrage
- Enrichissement simple
- Validation de schéma
- Chaînage de transformations

### Phase 5 — API de pilotage
- CRUD pipelines
- CRUD connexions / credentials référencés
- Lancement manuel d'un job
- Consultation de l'historique d'exécution
- Consultation des logs et statuts
- Documentation OpenAPI

### Phase 6 — Interface visuelle MVP
- Authentification
- Liste des pipelines
- Designer visuel type node-based
- Formulaires de configuration source/target/transforms
- Page d'exécution temps réel
- Historique et détail d'un run

### Phase 7 — Orchestration et scheduling
- Planification cron
- Exécution asynchrone via worker
- File d'attente des jobs
- Retry policy configurable
- Limitation de concurrence

### Phase 8 — Qualité et exploitation
- Tests e2e
- Profiling et optimisation mémoire
- Observabilité complète
- Gestion fine des erreurs utilisateur / techniques
- Runbooks d'exploitation
- Politique de versionnement des pipelines

### Phase 9 — Évolutions avancées
- DAG multi-branches
- CDC
- Streaming
- Templates de pipelines
- Multi-tenant
- RBAC avancé
- Versioning et rollback de pipeline

## MVP recommandé
Pour aller vite sans te disperser, le MVP devrait couvrir :
- 1 UI web
- 1 API backend Go
- 1 worker Go
- PostgreSQL pour les métadonnées
- Connecteurs : CSV + PostgreSQL
- Transformations simples : map, filter, cast
- Exécution manuelle + scheduling cron
- Historique d'exécution + logs

## Ordre d'implémentation conseillé
1. Initialiser la structure du repo
2. Monter le backend HTTP minimal
3. Brancher la base de métadonnées
4. Développer le modèle de pipeline
5. Implémenter le moteur ETL simple
6. Ajouter un connecteur source puis un loader cible
7. Exposer le pilotage via API
8. Construire l'UI visuelle MVP
9. Ajouter queue, scheduling et retries
10. Renforcer tests, sécurité et observabilité

## Livrables attendus par sprint
- **Sprint 1** : squelette projet + infra locale + metadata DB
- **Sprint 2** : moteur ETL minimal exécutable
- **Sprint 3** : connecteurs CSV/PostgreSQL + transformations de base
- **Sprint 4** : API complète MVP
- **Sprint 5** : UI visuelle MVP
- **Sprint 6** : scheduling, logs, hardening

## Risques à anticiper
- complexité trop tôt du designer visuel
- couplage fort entre UI, API et moteur ETL
- absence de modèle de métadonnées stable
- mauvaise gestion des erreurs/reprises
- connecteurs trop spécifiques et peu réutilisables
- sous-estimation de l'observabilité

## Recommandation d'expert
Le meilleur choix pour un ETL Go avec interface visuelle est de séparer très tôt **le moteur ETL**, **l'orchestrateur**, et **la couche UI/API**. Le moteur doit pouvoir s'exécuter sans interface graphique. L'UI ne doit être qu'un client d'édition et de supervision branché sur l'API.

## Étape suivante conseillée
Après validation de cette architecture, l'étape la plus rentable est de produire :
- le schéma des entités métier (`Pipeline`, `Run`, `Connector`, `Step`, `Schedule`)
- les interfaces Go du moteur ETL
- le contrat API initial
- un `docker-compose.yml` de développement
