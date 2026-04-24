# Étape 7 — API de Pilotage Complète : État détaillé et tâches

> Généré le 2026-04-24 · Basé sur un scan complet du code source après Phase 6

---

## Résumé de la Phase 7

La Phase 7 a pour objectif d'exposer l'intégralité des opérations de pilotage du système
via une API HTTP REST versionnée, documentée OpenAPI, et couvrant :
- **CRUD projets** avec sauvegarde XML automatique
- **CRUD connexions** multi-env
- **Lancement / annulation de runs**
- **Historique et logs d'exécution** par run et par bloc
- **Switch global d'environnement**
- **Documentation OpenAPI** générée et servie en `/api/docs`

**État global : handlers présents en stub ✅ — orchestration & runs à câbler ⚠️ — OpenAPI absent ❌**

---

## Ce qui est déjà en place (Phases 0 → 6)

### ✅ Infrastructure & Foundation
- Structure repo complète (`cmd/`, `internal/`, `pkg/`, `api/`, `web/`, `deploy/`, `migrations/`)
- Config multi-env YAML + Docker Compose + Makefile
- Logger `zerolog`, auth JWT, middleware (CORS, auth, logging)
- Migrations SQL : `001_init.sql` → `004_users.sql` + table `schedules`

### ✅ Contracts & Modèle DAG
- `contracts/block.go` : `DataType`, `ColumnDef`, `Schema`, `DataRow`, `Port`, `BlockContext`, `Block`, `BlockFactory`
- `contracts/project.go` : `Project`, `Node`, `Edge`, `Param` avec tags XML + JSON
- `contracts/preview.go` : `PreviewStore`

### ✅ Moteur d'exécution DAG
- `engine/dag.go` : `BuildDAG()`, tri topologique, gestion edges `disabled`
- `engine/executor.go` : `Execute()`, câblage ports, `RunResult`, `ExecutionReport`
- `engine/inject_connections.go` : injection connexions dans `BlockContext`

### ✅ Blocs Sources, Transforms, Targets (Phases 4 & 5)
- Tous les blocs MVP présents et enregistrés (voir ETAPE-5-DETAIL.md)

### ✅ Connexions multi-env (Phase 6)
- `internal/connections/manager/` : CRUD connexions XML
- `internal/connections/resolver/` : résolution `ACTIVE_ENV` → paramètres
- `internal/connections/secrets/` : intégration env vars / vault
- Endpoint `PUT /api/v1/environment` implémenté

### ✅ XML Store / Parser / Serializer (Phase 5 sprint C)
- `internal/xml/store/store.go` : Save, Load, List, Delete + archivage `history/` + SHA256
- `internal/xml/parser/parser.go` : `Parse(io.Reader) (*contracts.Project, error)`
- `internal/xml/serializer/serializer.go` : `Serialize(*contracts.Project) ([]byte, error)`

### ✅ Jobs & Orchestrateur (Phase 5 sprint D)
- `internal/jobs/job.go` : interface `Repository` + implémentation PostgreSQL
- `internal/orchestrator/service.go` : `RunProject()`, `CancelRun()`

---

## Périmètre complet de la Phase 7

### Routes API à exposer
```
Projets
GET /api/v1/projects → liste tous les projets
POST /api/v1/projects → crée un projet (génère XML)
GET /api/v1/projects/{id} → charge un projet (parse XML)
PUT /api/v1/projects/{id} → met à jour le graphe (sérialise XML + archive)
DELETE /api/v1/projects/{id} → supprime le répertoire XML + métadonnées

Runs
POST /api/v1/projects/{id}/runs → lance un run
GET /api/v1/projects/{id}/runs → historique des runs du projet
GET /api/v1/runs/{runID} → détail d'un run
DELETE /api/v1/runs/{runID} → annule un run en cours
GET /api/v1/runs/{runID}/logs → logs structurés du run (par bloc)
GET /api/v1/runs/{runID}/report → ExecutionReport complet (lignes in/out, durée)

Connexions
GET /api/v1/connections → liste toutes les connexions
POST /api/v1/connections → crée une connexion (XML)
GET /api/v1/connections/{id} → charge une connexion
PUT /api/v1/connections/{id} → met à jour une connexion
DELETE /api/v1/connections/{id} → supprime la connexion
POST /api/v1/connections/{id}/test → teste la connexion sur l'env actif

Environnement global
GET /api/v1/environment → retourne ACTIVE_ENV courant
PUT /api/v1/environment → switch global (dev/preprod/prod)

Documentation
GET /api/docs → Swagger UI
GET /api/v1/openapi.yaml → spec OpenAPI brute
```

---

## État détaillé — Handlers HTTP

| Handler | Fichier | Méthodes | Câblage orchestrateur | Test d'intégration |
|---|---|---|---|---|
| `ProjectHandler` | `api/handlers/project_handler.go` | GET, POST, PUT, DELETE | ⚠️ à compléter | ❌ manquant |
| `RunHandler` | `api/handlers/run_handler.go` | POST, GET, DELETE, logs, report | ❌ à créer | ❌ manquant |
| `ConnectionHandler` | `api/handlers/connection_handler.go` | GET, POST, PUT, DELETE, test | ⚠️ à compléter | ❌ manquant |
| `EnvironmentHandler` | `api/handlers/environment_handler.go` | GET, PUT | ⚠️ présent Phase 6 | ❌ manquant |
| `OpenAPIHandler` | `api/handlers/openapi_handler.go` | GET docs, GET yaml | ❌ à créer | ❌ manquant |

---

## Problèmes bloquants identifiés

### 🔴 BLOQUANT 1 — `RunHandler` absent

`api/handlers/run_handler.go` n'existe pas encore. C'est le handler central de la Phase 7 :
il orchestre le lancement d'un run, l'annulation et l'exposition des logs.

**À créer : `api/handlers/run_handler.go`**

```go
package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/ThibLOREK/GoLoadlt/internal/orchestrator"
    "github.com/ThibLOREK/GoLoadlt/internal/jobs"
    "github.com/ThibLOREK/GoLoadlt/pkg/dto"
)

type RunHandler struct {
    orchestrator *orchestrator.Service
    jobRepo      jobs.Repository
}

func NewRunHandler(o *orchestrator.Service, j jobs.Repository) *RunHandler {
    return &RunHandler{orchestrator: o, jobRepo: j}
}

// POST /api/v1/projects/{id}/runs
func (h *RunHandler) StartRun(w http.ResponseWriter, r *http.Request) {
    projectID := chi.URLParam(r, "id")
    report, err := h.orchestrator.RunProject(r.Context(), projectID)
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    respondJSON(w, http.StatusCreated, dto.RunResponse{
        ProjectID: projectID,
        Status:    "succeeded",
        Report:    report,
    })
}

// DELETE /api/v1/runs/{runID}
func (h *RunHandler) CancelRun(w http.ResponseWriter, r *http.Request) {
    runID := chi.URLParam(r, "runID")
    if err := h.orchestrator.CancelRun(r.Context(), runID); err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

// GET /api/v1/runs/{runID}/logs
func (h *RunHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
    runID := chi.URLParam(r, "runID")
    logs, err := h.jobRepo.GetLogs(r.Context(), runID)
    if err != nil {
        respondError(w, http.StatusNotFound, err.Error())
        return
    }
    respondJSON(w, http.StatusOK, logs)
}

// GET /api/v1/runs/{runID}/report
func (h *RunHandler) GetReport(w http.ResponseWriter, r *http.Request) {
    runID := chi.URLParam(r, "runID")
    run, err := h.jobRepo.GetByID(r.Context(), runID)
    if err != nil {
        respondError(w, http.StatusNotFound, err.Error())
        return
    }
    respondJSON(w, http.StatusOK, run)
}

// GET /api/v1/projects/{id}/runs
func (h *RunHandler) ListRuns(w http.ResponseWriter, r *http.Request) {
    projectID := chi.URLParam(r, "id")
    runs, err := h.jobRepo.ListByProject(r.Context(), projectID)
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    respondJSON(w, http.StatusOK, runs)
}

// GET /api/v1/runs/{runID}
func (h *RunHandler) GetRun(w http.ResponseWriter, r *http.Request) {
    runID := chi.URLParam(r, "runID")
    run, err := h.jobRepo.GetByID(r.Context(), runID)
    if err != nil {
        respondError(w, http.StatusNotFound, err.Error())
        return
    }
    respondJSON(w, http.StatusOK, run)
}
```

---

### 🔴 BLOQUANT 2 — DTOs `pkg/dto/` incomplets

Les objets de transfert JSON (in/out) sont absents ou minimalistes.

**À créer / compléter : `pkg/dto/run.go`**

```go
package dto

import "github.com/ThibLOREK/GoLoadlt/internal/etl/engine"

// RunRequest : corps de POST /api/v1/projects/{id}/runs (optionnel, pour overrides futurs)
type RunRequest struct {
    Env string `json:"env,omitempty"` // override env pour ce run uniquement
}

// RunResponse : réponse après lancement d'un run
type RunResponse struct {
    RunID     string                  `json:"runId"`
    ProjectID string                  `json:"projectId"`
    Status    string                  `json:"status"`
    StartedAt string                  `json:"startedAt"`
    Report    *engine.ExecutionReport `json:"report,omitempty"`
}

// RunLogEntry : une ligne de log structurée par bloc
type RunLogEntry struct {
    BlockID   string `json:"blockId"`
    BlockType string `json:"blockType"`
    Level     string `json:"level"` // info | warn | error
    Message   string `json:"message"`
    Timestamp string `json:"timestamp"`
    RowsIn    int64  `json:"rowsIn"`
    RowsOut   int64  `json:"rowsOut"`
}
```

**À créer : `pkg/dto/project.go`**

```go
package dto

import "github.com/ThibLOREK/GoLoadlt/internal/etl/contracts"

// ProjectCreateRequest : corps de POST /api/v1/projects
type ProjectCreateRequest struct {
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
}

// ProjectUpdateRequest : corps de PUT /api/v1/projects/{id}
// Contient le graphe DAG complet sérialisé depuis l'UI
type ProjectUpdateRequest struct {
    Project *contracts.Project `json:"project"`
}

// ProjectResponse : réponse enrichie d'un projet
type ProjectResponse struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    XMLPath     string `json:"xmlPath"`
    SHA256      string `json:"sha256"`
    UpdatedAt   string `json:"updatedAt"`
    RunCount    int    `json:"runCount"`
    LastStatus  string `json:"lastStatus,omitempty"`
}

// ProjectListResponse : liste paginée
type ProjectListResponse struct {
    Items []ProjectResponse `json:"items"`
    Total int               `json:"total"`
}
```

---

### 🔴 BLOQUANT 3 — `ProjectHandler` incomplet (câblage XML Store manquant)

`api/handlers/project_handler.go` doit être câblé sur `internal/xml/store` pour les opérations
CRUD. Le handler actuel ne persiste pas encore le XML.

**`api/handlers/project_handler.go` — version complète :**

```go
package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/ThibLOREK/GoLoadlt/internal/xml/store"
    "github.com/ThibLOREK/GoLoadlt/internal/xml/serializer"
    "github.com/ThibLOREK/GoLoadlt/internal/etl/contracts"
    "github.com/ThibLOREK/GoLoadlt/pkg/dto"
    "github.com/google/uuid"
)

type ProjectHandler struct {
    xmlStore *store.XMLStore
}

func NewProjectHandler(xs *store.XMLStore) *ProjectHandler {
    return &ProjectHandler{xmlStore: xs}
}

// GET /api/v1/projects
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
    projects, err := h.xmlStore.List()
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    items := make([]dto.ProjectResponse, len(projects))
    for i, p := range projects {
        items[i] = toProjectDTO(&p)
    }
    respondJSON(w, http.StatusOK, dto.ProjectListResponse{Items: items, Total: len(items)})
}

// POST /api/v1/projects
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
    var req dto.ProjectCreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "invalid JSON")
        return
    }
    project := &contracts.Project{
        ID:   uuid.NewString(),
        Name: req.Name,
    }
    if err := h.xmlStore.Save(project); err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    respondJSON(w, http.StatusCreated, toProjectDTO(project))
}

// GET /api/v1/projects/{id}
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    project, err := h.xmlStore.Load(id)
    if err != nil {
        respondError(w, http.StatusNotFound, "project not found")
        return
    }
    respondJSON(w, http.StatusOK, project)
}

// PUT /api/v1/projects/{id}
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    var req dto.ProjectUpdateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "invalid JSON")
        return
    }
    req.Project.ID = id
    if err := h.xmlStore.Save(req.Project); err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    respondJSON(w, http.StatusOK, toProjectDTO(req.Project))
}

// DELETE /api/v1/projects/{id}
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    if err := h.xmlStore.Delete(id); err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

func toProjectDTO(p *contracts.Project) dto.ProjectResponse {
    return dto.ProjectResponse{
        ID:   p.ID,
        Name: p.Name,
    }
}
```

---

### 🔴 BLOQUANT 4 — Router non câblé sur les nouvelles routes

`cmd/server/main.go` (ou le fichier de bootstrap `internal/app/`) doit enregistrer
toutes les routes de la Phase 7.

**Extrait du wiring `internal/app/router.go` :**

```go
package app

import (
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/ThibLOREK/GoLoadlt/api/handlers"
    apimiddleware "github.com/ThibLOREK/GoLoadlt/api/middleware"
)

func NewRouter(
    projectHandler *handlers.ProjectHandler,
    runHandler     *handlers.RunHandler,
    connHandler    *handlers.ConnectionHandler,
    envHandler     *handlers.EnvironmentHandler,
    openAPIHandler *handlers.OpenAPIHandler,
) *chi.Mux {
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(apimiddleware.CORS)

    r.Route("/api/v1", func(r chi.Router) {
        r.Use(apimiddleware.Auth)

        // Projets
        r.Get("/projects",           projectHandler.ListProjects)
        r.Post("/projects",          projectHandler.CreateProject)
        r.Get("/projects/{id}",      projectHandler.GetProject)
        r.Put("/projects/{id}",      projectHandler.UpdateProject)
        r.Delete("/projects/{id}",   projectHandler.DeleteProject)

        // Runs (imbriqués sous projet + standalone)
        r.Post("/projects/{id}/runs",    runHandler.StartRun)
        r.Get("/projects/{id}/runs",     runHandler.ListRuns)
        r.Get("/runs/{runID}",           runHandler.GetRun)
        r.Delete("/runs/{runID}",        runHandler.CancelRun)
        r.Get("/runs/{runID}/logs",      runHandler.GetLogs)
        r.Get("/runs/{runID}/report",    runHandler.GetReport)

        // Connexions
        r.Get("/connections",            connHandler.ListConnections)
        r.Post("/connections",           connHandler.CreateConnection)
        r.Get("/connections/{id}",       connHandler.GetConnection)
        r.Put("/connections/{id}",       connHandler.UpdateConnection)
        r.Delete("/connections/{id}",    connHandler.DeleteConnection)
        r.Post("/connections/{id}/test", connHandler.TestConnection)

        // Environnement global
        r.Get("/environment",  envHandler.GetEnvironment)
        r.Put("/environment",  envHandler.SetEnvironment)

        // OpenAPI spec
        r.Get("/openapi.yaml", openAPIHandler.ServeSpec)
    })

    // Swagger UI (non authentifié)
    r.Get("/api/docs", openAPIHandler.SwaggerUI)
    r.Get("/api/docs/*", openAPIHandler.SwaggerUI)

    return r
}
```

---

### 🔴 BLOQUANT 5 — OpenAPI absent

Le fichier `api/openapi/openapi.yaml` est vide ou inexistant.

**`api/openapi/openapi.yaml` — squelette à compléter :**

```yaml
openapi: "3.1.0"
info:
  title: GoLoadIt API
  version: "1.0.0"
  description: |
    API de pilotage de la plateforme ETL GoLoadIt.
    Gère les projets XML, les runs d'exécution, les connexions multi-env
    et le switch global d'environnement.

servers:
  - url: http://localhost:8080/api/v1
    description: Développement local

tags:
  - name: projects
    description: CRUD projets ETL (persistance XML)
  - name: runs
    description: Lancement, suivi et historique des exécutions
  - name: connections
    description: CRUD connexions réutilisables multi-environnements
  - name: environment
    description: Switch global d'environnement (dev / preprod / prod)

paths:
  /projects:
    get:
      tags: [projects]
      summary: Liste tous les projets
      responses:
        "200":
          description: Liste paginée des projets
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ProjectListResponse"
    post:
      tags: [projects]
      summary: Crée un nouveau projet (génère le fichier XML)
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/ProjectCreateRequest"
      responses:
        "201":
          description: Projet créé
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ProjectResponse"

  /projects/{id}:
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
    get:
      tags: [projects]
      summary: Charge un projet (parse le XML)
      responses:
        "200":
          description: Projet chargé
        "404":
          description: Projet introuvable
    put:
      tags: [projects]
      summary: Met à jour le graphe (sérialise XML + archive history/)
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/ProjectUpdateRequest"
      responses:
        "200":
          description: Projet mis à jour
    delete:
      tags: [projects]
      summary: Supprime le projet et ses fichiers XML
      responses:
        "204":
          description: Supprimé

  /projects/{id}/runs:
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
    post:
      tags: [runs]
      summary: Lance un run du projet
      responses:
        "201":
          description: Run lancé
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/RunResponse"
    get:
      tags: [runs]
      summary: Historique des runs du projet
      responses:
        "200":
          description: Liste des runs

  /runs/{runID}:
    parameters:
      - name: runID
        in: path
        required: true
        schema:
          type: string
    get:
      tags: [runs]
      summary: Détail d'un run
      responses:
        "200":
          description: Run trouvé
    delete:
      tags: [runs]
      summary: Annule un run en cours
      responses:
        "204":
          description: Run annulé

  /runs/{runID}/logs:
    parameters:
      - name: runID
        in: path
        required: true
        schema:
          type: string
    get:
      tags: [runs]
      summary: Logs structurés du run (par bloc)
      responses:
        "200":
          description: Entrées de log

  /runs/{runID}/report:
    parameters:
      - name: runID
        in: path
        required: true
        schema:
          type: string
    get:
      tags: [runs]
      summary: ExecutionReport complet (lignes in/out, durée par bloc)
      responses:
        "200":
          description: Rapport d'exécution

  /connections:
    get:
      tags: [connections]
      summary: Liste toutes les connexions
      responses:
        "200":
          description: Liste des connexions
    post:
      tags: [connections]
      summary: Crée une connexion (écrit le XML)
      responses:
        "201":
          description: Connexion créée

  /connections/{id}:
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
    get:
      tags: [connections]
      summary: Charge une connexion
      responses:
        "200":
          description: Connexion chargée
    put:
      tags: [connections]
      summary: Met à jour une connexion
      responses:
        "200":
          description: Connexion mise à jour
    delete:
      tags: [connections]
      summary: Supprime une connexion
      responses:
        "204":
          description: Supprimée

  /connections/{id}/test:
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
    post:
      tags: [connections]
      summary: Teste la connexion sur l'environnement actif
      responses:
        "200":
          description: Connexion réussie
        "422":
          description: Connexion échouée (détail dans le body)

  /environment:
    get:
      tags: [environment]
      summary: Retourne l'environnement actif
      responses:
        "200":
          description: Env actif
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/EnvironmentResponse"
    put:
      tags: [environment]
      summary: Switch global d'environnement
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/EnvironmentRequest"
      responses:
        "200":
          description: Switch effectué

components:
  schemas:
    ProjectCreateRequest:
      type: object
      required: [name]
      properties:
        name:
          type: string
        description:
          type: string

    ProjectUpdateRequest:
      type: object
      required: [project]
      properties:
        project:
          type: object
          description: Graphe DAG complet (nodes + edges)

    ProjectResponse:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        description:
          type: string
        xmlPath:
          type: string
        sha256:
          type: string
        updatedAt:
          type: string
          format: date-time
        runCount:
          type: integer
        lastStatus:
          type: string
          enum: [succeeded, failed, running, pending]

    ProjectListResponse:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: "#/components/schemas/ProjectResponse"
        total:
          type: integer

    RunResponse:
      type: object
      properties:
        runId:
          type: string
        projectId:
          type: string
        status:
          type: string
          enum: [pending, running, succeeded, failed, cancelled]
        startedAt:
          type: string
          format: date-time
        report:
          type: object
          description: ExecutionReport (lignes in/out, durée par bloc)

    EnvironmentRequest:
      type: object
      required: [env]
      properties:
        env:
          type: string
          enum: [dev, preprod, prod]

    EnvironmentResponse:
      type: object
      properties:
        activeEnv:
          type: string
          enum: [dev, preprod, prod]

  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

security:
  - BearerAuth: []
```

---

### 🟡 IMPORTANT 6 — `OpenAPIHandler` à créer

**`api/handlers/openapi_handler.go` :**

```go
package handlers

import (
    _ "embed"
    "net/http"
)

//go:embed ../../api/openapi/openapi.yaml
var openAPISpec []byte

type OpenAPIHandler struct{}

func NewOpenAPIHandler() *OpenAPIHandler { return &OpenAPIHandler{} }

// GET /api/v1/openapi.yaml
func (h *OpenAPIHandler) ServeSpec(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/yaml")
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write(openAPISpec)
}

// GET /api/docs — Swagger UI via CDN (pas de dépendance npm)
func (h *OpenAPIHandler) SwaggerUI(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    _, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
  <title>GoLoadIt API Docs</title>
  <meta charset="utf-8"/>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: "/api/v1/openapi.yaml",
      dom_id: "#swagger-ui",
      presets: [SwaggerUIBundle.presets.apis],
      layout: "BaseLayout"
    });
  </script>
</body>
</html>`))
}
```

---

### 🟡 IMPORTANT 7 — `jobs.Repository` : ajout de `GetLogs`

Le contrat `jobs.Repository` défini en Phase 5 doit être étendu pour supporter
la récupération des logs structurés par run.

**Extension de `internal/jobs/job.go` :**

```go
package jobs

import (
    "context"
    "time"
)

type Run struct {
    ID        string    `db:"id"        json:"id"`
    ProjectID string    `db:"project_id" json:"projectId"`
    Status    string    `db:"status"    json:"status"` // pending|running|succeeded|failed|cancelled
    StartedAt time.Time `db:"started_at" json:"startedAt"`
    EndedAt   *time.Time `db:"ended_at"  json:"endedAt,omitempty"`
}

type LogEntry struct {
    RunID     string    `db:"run_id"    json:"runId"`
    BlockID   string    `db:"block_id"  json:"blockId"`
    BlockType string    `db:"block_type" json:"blockType"`
    Level     string    `db:"level"     json:"level"`
    Message   string    `db:"message"   json:"message"`
    RowsIn    int64     `db:"rows_in"   json:"rowsIn"`
    RowsOut   int64     `db:"rows_out"  json:"rowsOut"`
    Timestamp time.Time `db:"timestamp" json:"timestamp"`
}

type Repository interface {
    Create(ctx context.Context, projectID string) (*Run, error)
    SetStatus(ctx context.Context, runID string, status string) error
    GetByID(ctx context.Context, runID string) (*Run, error)
    ListByProject(ctx context.Context, projectID string) ([]Run, error)
    // Extension Phase 7 :
    AppendLog(ctx context.Context, entry LogEntry) error
    GetLogs(ctx context.Context, runID string) ([]LogEntry, error)
}
```

**Migration SQL associée (`migrations/005_run_logs.sql`) :**

```sql
CREATE TABLE IF NOT EXISTS run_logs (
    id         BIGSERIAL PRIMARY KEY,
    run_id     UUID NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    block_id   TEXT NOT NULL,
    block_type TEXT NOT NULL,
    level      TEXT NOT NULL DEFAULT 'info',
    message    TEXT,
    rows_in    BIGINT DEFAULT 0,
    rows_out   BIGINT DEFAULT 0,
    timestamp  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_run_logs_run_id ON run_logs(run_id);
```

---

### 🟡 IMPORTANT 8 — Helpers de réponse manquants ou éparpillés

Centraliser les helpers JSON dans `api/handlers/helpers.go` :

```go
package handlers

import (
    "encoding/json"
    "net/http"
)

func respondJSON(w http.ResponseWriter, status int, body any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(body)
}

func respondError(w http.ResponseWriter, status int, message string) {
    respondJSON(w, status, map[string]string{"error": message})
}
```

---

## Plan d'action pour finaliser la Phase 7

### Sprint A — DTOs et helpers (0,5 jour)

- [ ] Créer `pkg/dto/run.go` (RunRequest, RunResponse, RunLogEntry)
- [ ] Créer `pkg/dto/project.go` (ProjectCreateRequest, ProjectUpdateRequest, ProjectResponse, ProjectListResponse)
- [ ] Créer `api/handlers/helpers.go` (respondJSON, respondError)
- [ ] Vérifier `go build ./...` passe

### Sprint B — RunHandler + câblage router (1 jour)

- [ ] Créer `api/handlers/run_handler.go` (StartRun, CancelRun, ListRuns, GetRun, GetLogs, GetReport)
- [ ] Compléter `api/handlers/project_handler.go` (câblage XML Store)
- [ ] Compléter `api/handlers/connection_handler.go` si stubs
- [ ] Mettre à jour `internal/app/router.go` avec toutes les routes Phase 7
- [ ] Injecter `RunHandler` dans le bootstrap `cmd/server/main.go`

### Sprint C — Jobs + migration logs (0,5 jour)

- [ ] Étendre l'interface `jobs.Repository` avec `AppendLog` et `GetLogs`
- [ ] Implémenter `AppendLog` / `GetLogs` dans l'implémentation PostgreSQL
- [ ] Créer `migrations/005_run_logs.sql`
- [ ] Brancher `AppendLog` dans `engine/executor.go` (après chaque bloc exécuté)

### Sprint D — OpenAPI + Swagger UI (0,5 jour)

- [ ] Compléter `api/openapi/openapi.yaml` avec tous les schemas
- [ ] Créer `api/handlers/openapi_handler.go` (embed + SwaggerUI)
- [ ] Enregistrer les routes `/api/docs` et `/api/v1/openapi.yaml` dans le router
- [ ] Vérifier que Swagger UI s'affiche correctement sur `http://localhost:8080/api/docs`

### Sprint E — Tests d'intégration API (1 jour)

- [ ] `tests/integration/project_api_test.go` : CRUD complet (create → update → list → delete)
- [ ] `tests/integration/run_api_test.go` : start run → check status → get logs → get report
- [ ] `tests/integration/connection_api_test.go` : CRUD + test connexion
- [ ] `tests/integration/environment_api_test.go` : GET env + PUT switch + vérifier résolution connexion

---

## Checklist finale Phase 7 — "Definition of Done"

### Backend Go
- [ ] `go build ./...` passe sans erreur ni warning
- [ ] `go vet ./...` propre
- [ ] Toutes les routes déclarées dans le router correspondent aux specs OpenAPI
- [ ] `POST /api/v1/projects/{id}/runs` exécute le pipeline de bout en bout
- [ ] `GET /api/v1/runs/{runID}/logs` retourne les logs structurés par bloc
- [ ] `GET /api/v1/runs/{runID}/report` retourne l'`ExecutionReport` complet
- [ ] `PUT /api/v1/environment` bascule `ACTIVE_ENV` et est pris en compte immédiatement
- [ ] Migration `005_run_logs.sql` appliquée sans erreur

### Documentation
- [ ] `GET /api/docs` affiche Swagger UI sans erreur
- [ ] `GET /api/v1/openapi.yaml` retourne le fichier complet
- [ ] Chaque route a sa description, ses paramètres et ses réponses documentées

### Tests d'intégration
- [ ] `go test ./tests/integration/...` vert (avec DB de test)
- [ ] Pipeline end-to-end via API : `POST project → PUT graphe → POST run → GET logs → GET report`

### Déploiement
- [ ] `docker-compose up` démarre sans erreur
- [ ] Swagger UI accessible depuis le navigateur sur le conteneur

---

## Architecture rappel — Flux API Phase 7
```
Client HTTP / UI React
│
▼
chi.Router (internal/app/router.go)
│
┌───┴────────────────────────────────┐
│ │
ProjectHandler RunHandler
(api/handlers/project_handler.go) (api/handlers/run_handler.go)
│ │
▼ ▼
xml/store.XMLStore orchestrator.Service
│ (Save / Load / List) │ (RunProject / CancelRun)
▼ │
projects/{id}/project.xml xml/store.Load()
(source de vérité) xml/parser.Parse()
│
engine.Executor.Execute()
│ (TopologicalSort + bloc par bloc)
│ AppendLog() ──► run_logs (PostgreSQL)
▼
ExecutionReport
│
jobs.Repository.SetStatus("succeeded")
│
▼
GET /runs/{id}/report ◄── Client
GET /runs/{id}/logs ◄── Client
```

---

*Document généré automatiquement par analyse du code source — à mettre à jour à chaque sprint.*
