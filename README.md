# GoLoadIt — Plateforme ETL moderne en Go

> Outil ETL modulaire, rapide et sans prise de tête, conçu en Go pour déplacer,
> transformer et charger tes données avec efficacité.

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](LICENSE)

---

## Sommaire

- [Présentation](#présentation)
- [Architecture](#architecture)
- [Stack technique](#stack-technique)
- [Prérequis](#prérequis)
- [Installation et lancement](#installation-et-lancement)
- [Configuration](#configuration)
- [Structure du projet](#structure-du-projet)
- [Moteur ETL — Concepts clés](#moteur-etl--concepts-clés)
- [Connecteurs disponibles](#connecteurs-disponibles)
- [Commandes Make](#commandes-make)
- [Observabilité](#observabilité)
- [Feuille de route](#feuille-de-route)

---

## Présentation

**GoLoadIt** est une plateforme ETL (Extract, Transform, Load) conçue pour les
développeurs et équipes data. Elle permet de :

- Concevoir des flux de données via une **interface visuelle** type node-based
- Exécuter des pipelines de façon **fiable, traçable et répétable**
- Connecter plusieurs sources et cibles (bases de données, fichiers CSV, APIs REST)
- Planifier des exécutions via un **scheduler cron**
- Superviser les runs en **temps réel** avec logs structurés et métriques

Le projet est découpé en trois processus indépendants :

| Processus | Rôle |
|-----------|------|
| `server`  | API HTTP REST + exposition de l'UI web + endpoints d'observabilité |
| `worker`  | Exécution asynchrone des jobs ETL en arrière-plan |
| `migrate` | Application des migrations du schéma PostgreSQL |

---

## Architecture
```text
┌─────────────────────────────────────────────────────┐
│ Presentation Layer │
│ UI Web (React) · API HTTP (chi) │
│ WebSocket / SSE pour suivi temps réel │
└───────────────────────┬─────────────────────────────┘
│
┌───────────────────────▼─────────────────────────────┐
│ Application Layer │
│ Orchestrateur · Gestion des jobs · Scheduler │
└───────────────────────┬─────────────────────────────┘
│
┌───────────────────────▼─────────────────────────────┐
│ Domain ETL Layer │
│ Contracts (interfaces) · Engine · Transformers │
└───────────────────────┬─────────────────────────────┘
│
┌───────────────────────▼─────────────────────────────┐
│ Infrastructure Layer │
│ Connecteurs · PostgreSQL · Redis · Logs · Telemetry│
└─────────────────────────────────────────────────────┘
```

---

## Stack technique

### Backend
| Composant       | Technologie                          |
|-----------------|--------------------------------------|
| Langage         | Go 1.24+                             |
| Router HTTP     | `chi` v5                             |
| Base de données | PostgreSQL 16 (via `pgx/v5`)         |
| Cache / Queue   | Redis 7                              |
| Configuration   | `viper` + `.env` via `godotenv`      |
| Logs            | `zerolog` (JSON structuré)           |
| Observabilité   | OpenTelemetry + Prometheus + Grafana |
| Auth            | JWT (`golang-jwt/jwt/v5`)            |

### Frontend
| Composant    | Technologie                        |
|--------------|------------------------------------|
| Framework    | React + TypeScript                 |
| Designer     | React Flow (éditeur de graphes)    |
| UI Kit       | shadcn/ui ou MUI                   |
| Temps réel   | WebSocket / SSE                    |

### Déploiement
- **Local** : Docker Compose (tout-en-un)
- **Production** : Kubernetes (manifests dans `deploy/k8s/`)

---

## Prérequis

Avant de lancer le projet, assure-toi d'avoir installé :

- [Go 1.24+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) et [Docker Compose](https://docs.docker.com/compose/)
- `make` (disponible nativement sur Linux/macOS, [Git Bash](https://gitforwindows.org/) sur Windows)

---

## Installation et lancement

### Option 1 — Docker Compose (recommandé)

Lance l'intégralité de la stack (PostgreSQL, Redis, Server, Worker, Prometheus, Grafana)
en une seule commande :

```bash
# 1. Cloner le dépôt
git clone https://github.com/ThibLOREK/GoLoadlt.git
cd GoLoadlt

# 2. Copier et adapter la configuration
cp configs/.env.example configs/.env
# Édite configs/.env selon ton environnement

# 3. Démarrer tous les services
docker compose -f deploy/docker/docker-compose.yml up --build
```

Les services sont ensuite accessibles à :

| Service    | URL                          |
|------------|------------------------------|
| API Server | http://localhost:8080        |
| Prometheus | http://localhost:9090        |
| Grafana    | http://localhost:3001        |
| PostgreSQL | localhost:5432               |
| Redis      | localhost:6379               |

> Identifiants Grafana par défaut : `admin` / `admin`

---

### Option 2 — Lancement natif (développement)

```bash
# 1. Cloner le dépôt
git clone https://github.com/ThibLOREK/GoLoadlt.git
cd GoLoadlt

# 2. Installer les dépendances Go
make tidy

# 3. Démarrer PostgreSQL et Redis (via Docker suffit)
docker compose -f deploy/docker/docker-compose.yml up postgres redis -d

# 4. Copier et adapter la configuration
cp configs/.env.example configs/.env

# 5. Appliquer les migrations
go run ./cmd/migrate

# 6. Lancer le serveur API
make run-server

# 7. Dans un second terminal, lancer le worker
make run-worker
```

---

### Option 3 — Build et exécution des binaires

```bash
# Compiler les binaires dans bin/
make build

# Exécuter
./bin/server
./bin/worker
```

---

## Configuration

Toute la configuration se fait via le fichier `configs/.env`
(copié depuis `configs/.env.example`) :

```env
APP_NAME=go-etl-studio        # Nom de l'application
APP_ENV=development            # Environnement : development | production
HTTP_PORT=8080                 # Port d'écoute du serveur HTTP
POSTGRES_DSN=postgres://etl:etl@localhost:5432/etl_studio?sslmode=disable
REDIS_ADDR=localhost:6379      # Adresse Redis
JWT_SECRET=change-me           # Clé secrète JWT — à changer en production !
FRONTEND_ORIGIN=http://localhost:3000  # Origine autorisée (CORS)
```

> ⚠️ **Ne jamais committer le fichier `.env` réel.** Seul `.env.example` doit être versionné.

---

## Structure du projet
```text
GoLoadlt/
├── cmd/
│ ├── server/ # Point d'entrée : API HTTP + supervision
│ ├── worker/ # Point d'entrée : exécuteur de jobs ETL
│ └── migrate/ # Point d'entrée : migrations BDD
├── internal/
│ ├── app/ # Bootstrap applicatif et câblage des dépendances
│ ├── config/ # Chargement et validation de la configuration
│ ├── logger/ # Initialisation du logger zerolog
│ ├── telemetry/ # OpenTelemetry + Prometheus (métriques, traces, healthchecks)
│ ├── security/ # JWT, AuthN/AuthZ, gestion des secrets
│ ├── etl/
│ │ ├── contracts/ # Interfaces Go : Extractor, Transformer, Loader
│ │ ├── engine/ # Moteur d'exécution ETL (executor, builder)
│ │ ├── pipeline/ # Modèle de pipeline et ses étapes
│ │ ├── scheduler/ # Planification cron
│ │ └── transformers/ # Transformations chaînables (map, filter, cast)
│ ├── connectors/ # Connecteurs sources/cibles (PostgreSQL, CSV, API...)
│ ├── orchestrator/ # Gestion globale des exécutions et dépendances
│ ├── jobs/ # File de jobs, états, retries
│ ├── services/ # Logique métier applicative
│ └── storage/ # Accès aux stores (PostgreSQL, Redis)
├── pkg/
│ ├── models/ # Types partageables
│ ├── dto/ # Contrats API (request/response)
│ └── utils/ # Helpers génériques
├── api/
│ ├── handlers/ # Handlers HTTP
│ ├── middleware/ # Middlewares (auth, logging, cors, recovery)
│ └── openapi/ # Contrat OpenAPI / Swagger
├── web/
│ └── ui/ # Frontend React (designer visuel de pipelines)
├── deploy/
│ └── docker/
│ ├── Dockerfile.server
│ ├── Dockerfile.worker
│ ├── docker-compose.yml
│ └── prometheus.yml
├── configs/
│ └── .env.example # Template de configuration
├── migrations/ # Fichiers SQL de migration
├── scripts/ # Scripts CI / bootstrap / dev
├── tests/
│ ├── integration/
│ ├── e2e/
│ └── fixtures/
└── docs/
├── adr/ # Architecture Decision Records
├── architecture/ # Diagrammes
└── runbooks/ # Guides d'exploitation
```

---

## Moteur ETL — Concepts clés

Le cœur du projet repose sur **trois interfaces Go** définies dans `internal/etl/contracts/contracts.go` :

```go
// Un enregistrement de données = map clé/valeur
type Record map[string]any

// Extractor lit les données depuis une source
type Extractor interface {
    Extract(ctx context.Context) ([]Record, error)
}

// Transformer modifie ou filtre les enregistrements
type Transformer interface {
    Transform(ctx context.Context, in []Record) ([]Record, error)
}

// Loader écrit les données vers une cible
type Loader interface {
    Load(ctx context.Context, in []Record) error
}
```

### Flux d'exécution d'un pipeline
```text
[Source] → Extract() → []Record
↓
Transform() → []Record (optionnel)
↓
Load() → [Cible]
```

L'`Executor` (`internal/etl/engine/executor.go`) orchestre ces trois étapes
avec gestion des erreurs, traces OpenTelemetry et logs structurés à chaque phase.

### États d'un job
L'`Executor` (`internal/etl/engine/executor.go`) orchestre ces trois étapes
avec gestion des erreurs, traces OpenTelemetry et logs structurés à chaque phase.

### États d'un job
```text
pending → running → succeeded
↘ failed → (retry) → running
↘ cancelled
```

---

## Connecteurs disponibles

| Connecteur       | Type   | Direction    | Statut  |
|------------------|--------|--------------|---------|
| PostgreSQL       | SQL    | Source/Cible | ✅ MVP  |
| CSV              | Fichier| Source/Cible | ✅ MVP  |
| API REST         | HTTP   | Source       | 🔜 Prévu |
| MySQL            | SQL    | Source/Cible | 🔜 Prévu |
| SQL Server       | SQL    | Source/Cible | 🔜 Prévu |
| S3 / Blob        | Objet  | Source/Cible | 🔜 Prévu |

---

## Commandes Make

```bash
make run-server   # Lance le serveur API (go run ./cmd/server)
make run-worker   # Lance le worker ETL (go run ./cmd/worker)
make build        # Compile les binaires dans bin/
make test         # Exécute tous les tests (go test ./...)
make fmt          # Formate le code source (gofmt)
make tidy         # Synchronise go.mod et go.sum (go mod tidy)
```

---

## Observabilité

GoLoadIt embarque une stack d'observabilité complète dès le départ :

- **Logs JSON structurés** via `zerolog` — chaque étape ETL est loggée avec contexte
- **Traces distribuées** via OpenTelemetry — chaque pipeline génère un span `pipeline.execute`
  avec sous-spans `extract`, `transform`, `load`
- **Métriques Prometheus** — exposées sur `/metrics`
- **Dashboards Grafana** — accessibles sur http://localhost:3001 après `docker compose up`
- **Healthcheck** — endpoint `/health` pour vérifier l'état du serveur

---

## Feuille de route

| Phase | Contenu | Statut |
|-------|---------|--------|
| 0 | Cadrage et ADR initiaux | ✅ |
| 1 | Fondation : structure, config, Docker, PostgreSQL | ✅ |
| 2 | Moteur ETL : interfaces, engine, executor | ✅ |
| 3 | Connecteurs CSV + PostgreSQL | 🚧 En cours |
| 4 | Transformations : map, filter, cast, chaînage | 🚧 En cours |
| 5 | API REST complète (CRUD pipelines, runs, logs) | 🔜 |
| 6 | Interface visuelle MVP (React Flow) | 🔜 |
| 7 | Scheduling cron + queue de jobs + retries | 🔜 |
| 8 | Tests e2e, optimisation, runbooks | 🔜 |
| 9 | DAG, CDC, streaming, multi-tenant | 🔜 |

---

## Contribuer

1. Fork le dépôt
2. Crée une branche : `git checkout -b feature/ma-fonctionnalite`
3. Commite tes changements : `git commit -m "feat: description"`
4. Pousse : `git push origin feature/ma-fonctionnalite`
5. Ouvre une Pull Request

Merci de suivre les conventions de nommage Go standard et d'ajouter des tests
pour toute nouvelle fonctionnalité.

---

## Licence

Ce projet est distribué sous licence **Apache 2.0** — voir le fichier [LICENSE](LICENSE).