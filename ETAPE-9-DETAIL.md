# Étape 9 — Orchestration et Scheduling : État détaillé et tâches

> Généré le 2026-04-24 · Basé sur un scan complet du code source après Phase 8

---

## Résumé de la Phase 9

La Phase 9 a pour objectif de mettre en place l'**orchestration asynchrone complète** des runs ETL :
planification cron par projet, worker Go avec file d'attente persistée, retry policy configurable,
limitation de concurrence, et suivi temps réel des runs en cours.

**État global : scheduler stub présent ✅ — worker asynchrone absent ❌ — queue absente ❌ — retry absent ❌**

---

## Ce qui est déjà en place (Phases 0 → 8)

### ✅ Infrastructure & Foundation
- Structure repo complète (`cmd/`, `internal/`, `pkg/`, `api/`, `web/`, `deploy/`, `migrations/`)
- Config multi-env YAML + Docker Compose + Makefile
- Logger `zerolog`, auth JWT, middleware (CORS, auth, logging)
- Migrations SQL : `001_init.sql` → `005_run_logs.sql`

### ✅ Contracts & Modèle DAG
- `contracts/block.go` : `DataType`, `ColumnDef`, `Schema`, `DataRow`, `Port`, `BlockContext`, `Block`, `BlockFactory`
- `contracts/project.go` : `Project`, `Node`, `Edge`, `Param` avec tags XML + JSON
- `contracts/preview.go` : `PreviewStore`

### ✅ Moteur d'exécution DAG (Phase 3)
- `engine/dag.go` : `BuildDAG()`, tri topologique, gestion edges `disabled`
- `engine/executor.go` : `Execute()`, câblage ports, `RunResult`, `ExecutionReport`
- `engine/inject_connections.go` : injection connexions dans `BlockContext`

### ✅ Blocs Sources, Transforms, Targets (Phases 4 & 5)
- Tous les blocs MVP enregistrés et fonctionnels

### ✅ Connexions multi-env (Phase 6)
- `internal/connections/manager/` : CRUD XML + persistance `ACTIVE_ENV`
- `internal/connections/resolver/` : résolution env actif → DSN
- `internal/connections/secrets/` : env vars + stub Vault

### ✅ XML Store / Parser / Serializer (Phase 5)
- `internal/xml/store/store.go` : Save, Load, List, Delete + archivage `history/` + SHA256
- `internal/xml/parser/parser.go` et `internal/xml/serializer/serializer.go` : fonctionnels

### ✅ Jobs & Orchestrateur synchrone (Phase 5 & 7)
- `internal/jobs/job.go` : interface `Repository` + implémentation PostgreSQL
- `internal/orchestrator/service.go` : `RunProject()` synchrone, `CancelRun()`

### ✅ API de pilotage (Phase 7)
- CRUD projets, connexions, runs, environment
- `GET /api/v1/runs/{runID}/logs` et `/report`
- OpenAPI documenté, Swagger UI

### ✅ Interface visuelle (Phase 8)
- Canvas React Flow fonctionnel, palette blocs, configuration, exécution manuelle
- Suivi temps réel via WebSocket (statut par bloc)

### ⚠️ Scheduler stub présent mais vide
- `internal/etl/scheduler/` : répertoire présent, fichiers minimalistes
- Migration `003_schedules.sql` : table `schedules` existante en base
- Aucune goroutine de scheduling démarrée au boot du serveur

---

## Périmètre complet de la Phase 9

### Nouvelles routes API à exposer
```
Scheduling (cron par projet)
GET /api/v1/projects/{id}/schedule → récupère le cron du projet
PUT /api/v1/projects/{id}/schedule → active / modifie le cron
DELETE /api/v1/projects/{id}/schedule → désactive le scheduling

Queue & Worker
GET /api/v1/queue → état de la file d'attente (jobs pending/running)
GET /api/v1/worker/status → statut du worker (active goroutines, slots libres)

Retry policy par projet
GET /api/v1/projects/{id}/retry-policy → récupère la retry policy
PUT /api/v1/projects/{id}/retry-policy → configure la retry policy
```

### Nouvelles migrations SQL
```
migrations/006_schedules_full.sql → enrichissement de la table schedules
migrations/007_retry_policy.sql → table retry_policies
```

---

## État détaillé — Composants Phase 9

| Composant | Fichier | État |
|---|---|---|
| Scheduler cron | `internal/etl/scheduler/scheduler.go` | ❌ stub vide |
| Worker asynchrone | `internal/worker/worker.go` | ❌ absent |
| File d'attente | `internal/worker/queue.go` | ❌ absente |
| Retry engine | `internal/worker/retry.go` | ❌ absent |
| Concurrency limiter | `internal/worker/limiter.go` | ❌ absent |
| ScheduleHandler | `api/handlers/schedule_handler.go` | ❌ absent |
| WorkerHandler | `api/handlers/worker_handler.go` | ❌ absent |
| Orchestrateur async | `internal/orchestrator/service.go` | ⚠️ synchrone uniquement |
| UI Scheduling | `web/ui/src/pages/SchedulePage.tsx` | ❌ absente |
| Migration schedules enrichie | `migrations/006_schedules_full.sql` | ❌ absente |
| Migration retry_policies | `migrations/007_retry_policy.sql` | ❌ absente |

---

## Problèmes bloquants identifiés

### 🔴 BLOQUANT 1 — Worker asynchrone absent

Le serveur actuel exécute les runs de façon **synchrone** dans le handler HTTP :
l'appelant attend la fin du run pour recevoir la réponse. Sur un pipeline long,
le client HTTP timeout et le run devient inobservable.

**Architecture cible :** le handler `POST /runs` enfile le run dans une queue en mémoire
(buffered channel) et répond immédiatement `202 Accepted`. Un pool de goroutines worker
consomme la queue et exécute les runs en parallèle avec une limite de concurrence.

**À créer : `internal/worker/worker.go`**

```go
package worker

import (
    "context"
    "sync"

    "github.com/rs/zerolog"
    "github.com/ThibLOREK/GoLoadlt/internal/orchestrator"
    "github.com/ThibLOREK/GoLoadlt/internal/jobs"
)

// Job représente une demande d'exécution mise en file d'attente.
type Job struct {
    ProjectID string
    RunID     string
    Ctx       context.Context
}

// Worker consomme la queue et exécute les runs via l'orchestrateur.
type Worker struct {
    queue        <-chan Job
    orchestrator *orchestrator.Service
    jobRepo      jobs.Repository
    log          zerolog.Logger
    concurrency  int
    wg           sync.WaitGroup
}

func New(
    queue <-chan Job,
    orch *orchestrator.Service,
    jobRepo jobs.Repository,
    concurrency int,
    log zerolog.Logger,
) *Worker {
    return &Worker{
        queue:        queue,
        orchestrator: orch,
        jobRepo:      jobRepo,
        log:          log,
        concurrency:  concurrency,
    }
}

// Start lance N goroutines worker et bloque jusqu'à ce que ctx soit annulé.
func (w *Worker) Start(ctx context.Context) {
    sem := make(chan struct{}, w.concurrency)
    for {
        select {
        case <-ctx.Done():
            w.wg.Wait()
            return
        case job, ok := <-w.queue:
            if !ok {
                w.wg.Wait()
                return
            }
            sem <- struct{}{} // acquiert un slot
            w.wg.Add(1)
            go func(j Job) {
                defer w.wg.Done()
                defer func() { <-sem }() // libère le slot
                w.execute(j)
            }(job)
        }
    }
}

func (w *Worker) execute(job Job) {
    log := w.log.With().Str("projectID", job.ProjectID).Str("runID", job.RunID).Logger()
    log.Info().Msg("worker: démarrage du run")

    _ = w.jobRepo.SetStatus(job.Ctx, job.RunID, "running")
    report, err := w.orchestrator.Execute(job.Ctx, job.ProjectID)
    if err != nil {
        log.Error().Err(err).Msg("worker: run échoué")
        _ = w.jobRepo.SetStatus(job.Ctx, job.RunID, "failed")
        return
    }
    log.Info().
        Int64("rowsTotal", report.TotalRows).
        Dur("duration", report.Duration).
        Msg("worker: run terminé avec succès")
    _ = w.jobRepo.SetStatus(job.Ctx, job.RunID, "succeeded")
}
```

---

### 🔴 BLOQUANT 2 — File d'attente (Queue) absente

**À créer : `internal/worker/queue.go`**

```go
package worker

import (
    "context"
    "fmt"
    "sync/atomic"
    "time"
)

// Queue est un canal bufferisé de Jobs avec métriques intégrées.
type Queue struct {
    ch      chan Job
    pending atomic.Int64
    total   atomic.Int64
}

func NewQueue(capacity int) *Queue {
    return &Queue{ch: make(chan Job, capacity)}
}

// Enqueue tente d'ajouter un job à la file. Retourne une erreur si la file est pleine.
func (q *Queue) Enqueue(job Job) error {
    select {
    case q.ch <- job:
        q.pending.Add(1)
        q.total.Add(1)
        return nil
    default:
        return fmt.Errorf("queue pleine (%d/%d) — réessayer plus tard", q.pending.Load(), cap(q.ch))
    }
}

// Chan retourne le channel en lecture seule pour le worker.
func (q *Queue) Chan() <-chan Job { return q.ch }

// Stats retourne les métriques actuelles de la file.
func (q *Queue) Stats() QueueStats {
    return QueueStats{
        Pending:  int(q.pending.Load()),
        Capacity: cap(q.ch),
        Total:    int(q.total.Load()),
    }
}

// QueueStats expose les métriques de la file pour l'API.
type QueueStats struct {
    Pending  int `json:"pending"`
    Capacity int `json:"capacity"`
    Total    int `json:"total"`  // total depuis le démarrage
}
```

---

### 🔴 BLOQUANT 3 — Orchestrateur non adapté à l'exécution asynchrone

`internal/orchestrator/service.go` expose `RunProject()` qui est synchrone.
Il faut le scinder en deux : `Enqueue()` (retourne immédiatement un `runID`)
et `Execute()` (appelé par le worker).

**`internal/orchestrator/service.go` — refactoring Phase 9 :**

```go
package orchestrator

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/ThibLOREK/GoLoadlt/internal/etl/engine"
    "github.com/ThibLOREK/GoLoadlt/internal/xml/store"
    "github.com/ThibLOREK/GoLoadlt/internal/jobs"
    "github.com/ThibLOREK/GoLoadlt/internal/worker"
)

type Service struct {
    executor *engine.Executor
    xmlStore *store.XMLStore
    jobRepo  jobs.Repository
    queue    *worker.Queue
}

func NewService(
    executor *engine.Executor,
    xmlStore *store.XMLStore,
    jobRepo jobs.Repository,
    queue *worker.Queue,
) *Service {
    return &Service{executor: executor, xmlStore: xmlStore, jobRepo: jobRepo, queue: queue}
}

// Enqueue crée un run en base (status=pending) et l'enfile dans la queue worker.
// Retourne le runID immédiatement — non bloquant.
func (s *Service) Enqueue(ctx context.Context, projectID string) (string, error) {
    run, err := s.jobRepo.Create(ctx, projectID)
    if err != nil {
        return "", fmt.Errorf("orchestrator.Enqueue: %w", err)
    }
    job := worker.Job{
        ProjectID: projectID,
        RunID:     run.ID,
        Ctx:       context.Background(), // contexte détaché du HTTP — run survit à la déconnexion
    }
    if err := s.queue.Enqueue(job); err != nil {
        _ = s.jobRepo.SetStatus(ctx, run.ID, "failed")
        return "", fmt.Errorf("orchestrator.Enqueue: queue pleine: %w", err)
    }
    return run.ID, nil
}

// Execute charge le XML, parse le DAG et exécute. Appelé par le Worker (pas le handler HTTP).
func (s *Service) Execute(ctx context.Context, projectID string) (*engine.ExecutionReport, error) {
    project, err := s.xmlStore.Load(projectID)
    if err != nil {
        return nil, fmt.Errorf("orchestrator.Execute: load XML: %w", err)
    }
    return s.executor.Execute(ctx, project)
}

// CancelRun annule un run en cours (best-effort via context cancel — voir Phase 9 limiter).
func (s *Service) CancelRun(ctx context.Context, runID string) error {
    return s.jobRepo.SetStatus(ctx, runID, "cancelled")
}
```

**Adapter `RunHandler.StartRun` pour utiliser `Enqueue` :**

```go
// ✅ Phase 9 — réponse 202 Accepted immédiate
func (h *RunHandler) StartRun(w http.ResponseWriter, r *http.Request) {
    projectID := chi.URLParam(r, "id")
    runID, err := h.orchestrator.Enqueue(r.Context(), projectID)
    if err != nil {
        respondError(w, http.StatusServiceUnavailable, err.Error())
        return
    }
    respondJSON(w, http.StatusAccepted, dto.RunResponse{
        RunID:     runID,
        ProjectID: projectID,
        Status:    "pending",
        StartedAt: time.Now().UTC().Format(time.RFC3339),
    })
}
```

---

### 🔴 BLOQUANT 4 — Scheduler cron vide

`internal/etl/scheduler/scheduler.go` est un stub. Il doit :
1. Se lancer au boot du serveur (dans `cmd/server/main.go`)
2. Charger tous les schedules actifs depuis la base
3. Pour chaque schedule, calculer le prochain `next_run` via la lib `robfig/cron`
4. Enqueuer un run au bon moment via l'orchestrateur

**À implémenter : `internal/etl/scheduler/scheduler.go`**

```go
package scheduler

import (
    "context"
    "time"

    "github.com/robfig/cron/v3"
    "github.com/rs/zerolog"
    "github.com/ThibLOREK/GoLoadlt/internal/orchestrator"
    "github.com/ThibLOREK/GoLoadlt/internal/jobs"
)

// Schedule représente une planification cron pour un projet.
type Schedule struct {
    ID        string
    ProjectID string
    CronExpr  string    // ex: "0 2 * * *" — tous les jours à 2h
    Active    bool
    NextRun   time.Time
    LastRunAt *time.Time
}

// Repository persiste les schedules en PostgreSQL.
type Repository interface {
    ListActive(ctx context.Context) ([]Schedule, error)
    UpdateNextRun(ctx context.Context, id string, nextRun time.Time) error
    Create(ctx context.Context, s Schedule) (*Schedule, error)
    GetByProject(ctx context.Context, projectID string) (*Schedule, error)
    Upsert(ctx context.Context, s Schedule) (*Schedule, error)
    Delete(ctx context.Context, projectID string) error
}

// Scheduler orchestre les cron jobs de tous les projets.
type Scheduler struct {
    cron         *cron.Cron
    repo         Repository
    orchestrator *orchestrator.Service
    log          zerolog.Logger
    entryIDs     map[string]cron.EntryID // projectID → entryID cron
}

func New(repo Repository, orch *orchestrator.Service, log zerolog.Logger) *Scheduler {
    return &Scheduler{
        cron:         cron.New(cron.WithSeconds()),
        repo:         repo,
        orchestrator: orch,
        log:          log,
        entryIDs:     make(map[string]cron.EntryID),
    }
}

// Start charge les schedules actifs et démarre le cron en arrière-plan.
func (s *Scheduler) Start(ctx context.Context) error {
    schedules, err := s.repo.ListActive(ctx)
    if err != nil {
        return err
    }
    for _, sched := range schedules {
        if err := s.register(sched); err != nil {
            s.log.Warn().Err(err).Str("projectID", sched.ProjectID).Msg("scheduler: schedule invalide ignoré")
        }
    }
    s.cron.Start()
    s.log.Info().Int("schedules", len(schedules)).Msg("scheduler: démarré")
    go func() {
        <-ctx.Done()
        s.cron.Stop()
        s.log.Info().Msg("scheduler: arrêté")
    }()
    return nil
}

// register ajoute ou remplace un schedule dans le cron interne.
func (s *Scheduler) register(sched Schedule) error {
    if id, exists := s.entryIDs[sched.ProjectID]; exists {
        s.cron.Remove(id)
    }
    entryID, err := s.cron.AddFunc(sched.CronExpr, func() {
        ctx := context.Background()
        runID, err := s.orchestrator.Enqueue(ctx, sched.ProjectID)
        if err != nil {
            s.log.Error().Err(err).Str("projectID", sched.ProjectID).Msg("scheduler: échec enqueue")
            return
        }
        s.log.Info().Str("projectID", sched.ProjectID).Str("runID", runID).Msg("scheduler: run planifié lancé")
        next := s.cron.Entry(s.entryIDs[sched.ProjectID]).Next
        _ = s.repo.UpdateNextRun(ctx, sched.ID, next)
    })
    if err != nil {
        return err
    }
    s.entryIDs[sched.ProjectID] = entryID
    return nil
}

// Upsert crée ou met à jour le schedule d'un projet et l'active immédiatement.
func (s *Scheduler) Upsert(ctx context.Context, projectID, cronExpr string) (*Schedule, error) {
    sched, err := s.repo.Upsert(ctx, Schedule{ProjectID: projectID, CronExpr: cronExpr, Active: true})
    if err != nil {
        return nil, err
    }
    if err := s.register(*sched); err != nil {
        return nil, err
    }
    return sched, nil
}

// Delete supprime le schedule d'un projet et arrête son cron entry.
func (s *Scheduler) Delete(ctx context.Context, projectID string) error {
    if id, exists := s.entryIDs[projectID]; exists {
        s.cron.Remove(id)
        delete(s.entryIDs, projectID)
    }
    return s.repo.Delete(ctx, projectID)
}

// GetByProject retourne le schedule d'un projet.
func (s *Scheduler) GetByProject(ctx context.Context, projectID string) (*Schedule, error) {
    return s.repo.GetByProject(ctx, projectID)
}
```

**Dépendance à ajouter dans `go.mod` :**

```bash
go get github.com/robfig/cron/v3
```

---

### 🔴 BLOQUANT 5 — Retry engine absent

Sans retry, tout run en échec reste en `failed` sans tentative de récupération automatique.

**À créer : `internal/worker/retry.go`**

```go
package worker

import (
    "context"
    "fmt"
    "math"
    "time"

    "github.com/rs/zerolog"
)

// RetryPolicy configure le comportement de réessai d'un run.
type RetryPolicy struct {
    MaxAttempts int           // nombre maximal de tentatives (1 = pas de retry)
    Delay       time.Duration // délai initial entre tentatives
    Backoff     float64       // multiplicateur exponentiel (1.0 = délai constant)
    MaxDelay    time.Duration // délai maximum après backoff
}

// DefaultRetryPolicy est appliqué si aucune policy n'est configurée sur le projet.
var DefaultRetryPolicy = RetryPolicy{
    MaxAttempts: 1,
    Delay:       30 * time.Second,
    Backoff:     2.0,
    MaxDelay:    10 * time.Minute,
}

// RunWithRetry exécute fn jusqu'à MaxAttempts fois en cas d'erreur.
func RunWithRetry(ctx context.Context, policy RetryPolicy, log zerolog.Logger, fn func(ctx context.Context) error) error {
    var lastErr error
    for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
        if err := ctx.Err(); err != nil {
            return fmt.Errorf("retry: contexte annulé après %d tentatives: %w", attempt-1, err)
        }
        lastErr = fn(ctx)
        if lastErr == nil {
            return nil
        }
        log.Warn().Err(lastErr).Int("attempt", attempt).Int("maxAttempts", policy.MaxAttempts).
            Msg("worker: tentative échouée")
        if attempt == policy.MaxAttempts {
            break
        }
        delay := time.Duration(float64(policy.Delay) * math.Pow(policy.Backoff, float64(attempt-1)))
        if delay > policy.MaxDelay {
            delay = policy.MaxDelay
        }
        log.Info().Dur("delay", delay).Msg("worker: attente avant réessai")
        select {
        case <-ctx.Done():
            return fmt.Errorf("retry: annulé pendant attente: %w", ctx.Err())
        case <-time.After(delay):
        }
    }
    return fmt.Errorf("retry: échec après %d tentatives: %w", policy.MaxAttempts, lastErr)
}
```

**Intégration dans `worker.execute()` :**

```go
// worker.go — modifier la méthode execute() pour utiliser RunWithRetry
func (w *Worker) execute(job Job) {
    log := w.log.With().Str("projectID", job.ProjectID).Str("runID", job.RunID).Logger()

    policy := w.retryRepo.GetPolicy(job.ProjectID) // récupère la policy ou DefaultRetryPolicy

    _ = w.jobRepo.SetStatus(job.Ctx, job.RunID, "running")

    err := RunWithRetry(job.Ctx, policy, log, func(ctx context.Context) error {
        _, execErr := w.orchestrator.Execute(ctx, job.ProjectID)
        return execErr
    })

    status := "succeeded"
    if err != nil {
        log.Error().Err(err).Msg("worker: run définitivement échoué")
        status = "failed"
    }
    _ = w.jobRepo.SetStatus(job.Ctx, job.RunID, status)
}
```

---

### 🔴 BLOQUANT 6 — Migrations incomplètes pour le scheduling

**`migrations/006_schedules_full.sql` — enrichissement de la table existante :**

```sql
-- Enrichissement de la table schedules (003_schedules.sql crée le squelette)
ALTER TABLE schedules
    ADD COLUMN IF NOT EXISTS active      BOOLEAN     NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS next_run    TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_run_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_schedules_active     ON schedules(active) WHERE active = TRUE;
CREATE INDEX IF NOT EXISTS idx_schedules_project_id ON schedules(project_id);
```

**`migrations/007_retry_policy.sql` — table de retry policy :**

```sql
CREATE TABLE IF NOT EXISTS retry_policies (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id   UUID        NOT NULL UNIQUE REFERENCES projects(id) ON DELETE CASCADE,
    max_attempts INT         NOT NULL DEFAULT 1,
    delay_ms     BIGINT      NOT NULL DEFAULT 30000,  -- millisecondes
    backoff      NUMERIC     NOT NULL DEFAULT 2.0,
    max_delay_ms BIGINT      NOT NULL DEFAULT 600000, -- 10 minutes
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW
``` 

CREATE TABLE IF NOT EXISTS retry_policies (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id   UUID        NOT NULL UNIQUE REFERENCES projects(id) ON DELETE CASCADE,
    max_attempts INT         NOT NULL DEFAULT 1,
    delay_ms     BIGINT      NOT NULL DEFAULT 30000,  -- millisecondes
    backoff      NUMERIC     NOT NULL DEFAULT 2.0,
    max_delay_ms BIGINT      NOT NULL DEFAULT 600000, -- 10 minutes
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE retry_policies IS 'Politique de réessai automatique par projet ETL';
COMMENT ON COLUMN retry_policies.max_attempts IS '1 = pas de retry';
COMMENT ON COLUMN retry_policies.backoff IS 'Multiplicateur exponentiel du délai (1.0 = délai constant)';
```

---

### 🔴 BLOQUANT 7 — ScheduleHandler et WorkerHandler absents

**À créer : `api/handlers/schedule_handler.go`**

```go
package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/ThibLOREK/GoLoadlt/internal/etl/scheduler"
    "github.com/ThibLOREK/GoLoadlt/internal/worker"
)

type ScheduleHandler struct {
    scheduler *scheduler.Scheduler
    queue     *worker.Queue
}

func NewScheduleHandler(s *scheduler.Scheduler, q *worker.Queue) *ScheduleHandler {
    return &ScheduleHandler{scheduler: s, queue: q}
}

// GET /api/v1/projects/{id}/schedule
func (h *ScheduleHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
    projectID := chi.URLParam(r, "id")
    sched, err := h.scheduler.GetByProject(r.Context(), projectID)
    if err != nil {
        respondError(w, http.StatusNotFound, "aucun schedule pour ce projet")
        return
    }
    respondJSON(w, http.StatusOK, sched)
}

// PUT /api/v1/projects/{id}/schedule
func (h *ScheduleHandler) UpsertSchedule(w http.ResponseWriter, r *http.Request) {
    projectID := chi.URLParam(r, "id")
    var body struct {
        CronExpr string `json:"cronExpr"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.CronExpr == "" {
        respondError(w, http.StatusBadRequest, "champ 'cronExpr' manquant (ex: '0 2 * * *')")
        return
    }
    sched, err := h.scheduler.Upsert(r.Context(), projectID, body.CronExpr)
    if err != nil {
        respondError(w, http.StatusUnprocessableEntity, "expression cron invalide: "+err.Error())
        return
    }
    respondJSON(w, http.StatusOK, sched)
}

// DELETE /api/v1/projects/{id}/schedule
func (h *ScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
    projectID := chi.URLParam(r, "id")
    if err := h.scheduler.Delete(r.Context(), projectID); err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

// GET /api/v1/queue
func (h *ScheduleHandler) GetQueueStatus(w http.ResponseWriter, r *http.Request) {
    respondJSON(w, http.StatusOK, h.queue.Stats())
}
```

**À créer : `api/handlers/worker_handler.go`**

```go
package handlers

import (
    "net/http"
    "runtime"

    "github.com/ThibLOREK/GoLoadlt/internal/worker"
)

type WorkerHandler struct {
    queue       *worker.Queue
    concurrency int
}

func NewWorkerHandler(q *worker.Queue, concurrency int) *WorkerHandler {
    return &WorkerHandler{queue: q, concurrency: concurrency}
}

// GET /api/v1/worker/status
func (h *WorkerHandler) Status(w http.ResponseWriter, r *http.Request) {
    stats := h.queue.Stats()
    respondJSON(w, http.StatusOK, map[string]any{
        "concurrency": h.concurrency,
        "queue":       stats,
        "goroutines":  runtime.NumGoroutine(),
    })
}
```

---

### 🟡 IMPORTANT 8 — Câblage complet dans `cmd/server/main.go`

Le serveur doit démarrer le `Scheduler` et le `Worker` au boot, avec graceful shutdown.

**Extrait du bootstrap `cmd/server/main.go` — Phase 9 :**

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"

    "github.com/ThibLOREK/GoLoadlt/internal/etl/scheduler"
    "github.com/ThibLOREK/GoLoadlt/internal/worker"
    // ... autres imports
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    // --- Wiring (injection de dépendances) ---
    cfg := config.Load()
    log := logger.New(cfg)
    db  := storage.NewPostgres(cfg)

    xmlStore    := xmlstore.New(cfg.ProjectsDir)
    jobRepo     := jobs.NewPostgresRepo(db)
    connMgr     := manager.New(cfg.ConnectionsDir, cfg.ActiveEnv)
    executor    := engine.NewExecutor(connMgr, log)

    // Phase 9 : queue + worker + scheduler
    queue       := worker.NewQueue(cfg.QueueCapacity)   // ex: 100
    orch        := orchestrator.NewService(executor, xmlStore, jobRepo, queue)
    schedRepo   := scheduler.NewPostgresRepo(db)
    sched       := scheduler.New(schedRepo, orch, log)
    w           := worker.New(queue.Chan(), orch, jobRepo, cfg.WorkerConcurrency, log)

    // Démarrage des services background
    if err := sched.Start(ctx); err != nil {
        log.Fatal().Err(err).Msg("impossible de démarrer le scheduler")
    }
    go w.Start(ctx)

    // Router HTTP
    schedHandler  := handlers.NewScheduleHandler(sched, queue)
    workerHandler := handlers.NewWorkerHandler(queue, cfg.WorkerConcurrency)
    router := app.NewRouter(/* ... handlers ... */, schedHandler, workerHandler)

    srv := &http.Server{Addr: cfg.Addr, Handler: router}
    go srv.ListenAndServe()
    log.Info().Str("addr", cfg.Addr).Msg("serveur démarré")

    <-ctx.Done()
    log.Info().Msg("arrêt gracieux en cours…")
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    _ = srv.Shutdown(shutdownCtx)
    log.Info().Msg("serveur arrêté")
}
```

**Nouvelles clés à ajouter dans `configs/config.dev.yaml` :**

```yaml
worker:
  concurrency: 3          # max 3 runs simultanés
  queue_capacity: 50      # max 50 runs en attente
  
scheduler:
  enabled: true           # désactivable en dev si besoin
```

---

### 🟡 IMPORTANT 9 — DTOs Phase 9 à ajouter dans `pkg/dto/`

**`pkg/dto/schedule.go` :**

```go
package dto

import "time"

// ScheduleRequest : corps de PUT /api/v1/projects/{id}/schedule
type ScheduleRequest struct {
    CronExpr string `json:"cronExpr"` // ex: "0 2 * * *"
}

// ScheduleResponse : réponse du scheduler
type ScheduleResponse struct {
    ProjectID string     `json:"projectId"`
    CronExpr  string     `json:"cronExpr"`
    Active    bool       `json:"active"`
    NextRun   *time.Time `json:"nextRun,omitempty"`
    LastRunAt *time.Time `json:"lastRunAt,omitempty"`
}

// RetryPolicyRequest : corps de PUT /api/v1/projects/{id}/retry-policy
type RetryPolicyRequest struct {
    MaxAttempts int     `json:"maxAttempts"`
    DelayMs     int64   `json:"delayMs"`
    Backoff     float64 `json:"backoff"`
    MaxDelayMs  int64   `json:"maxDelayMs"`
}
```

---

### 🟡 IMPORTANT 10 — UI Scheduling absente

**`web/ui/src/pages/SchedulePage.tsx` — page à créer :**

```tsx
// Route : /projects/:projectId/schedule
// Fonctionnalités :
// - Affiche le cron actuel et le prochain run calculé
// - Formulaire d'édition de l'expression cron (avec aide syntaxe)
// - Toggle ON/OFF du scheduling
// - Historique des derniers runs planifiés (lien vers RunHistory)
// - Indicateur visuel de la file d'attente (Pending / Capacity)
```

**Ajouter dans `App.tsx` :**

```tsx
<Route path="/projects/:projectId/schedule" element={<SchedulePage />} />
```

**Composant `CronInput.tsx` à créer :**

```tsx
// web/ui/src/components/scheduling/CronInput.tsx
// Input cron avec :
// - Validation de l'expression en temps réel
// - Affichage human-readable : "Tous les jours à 02:00"
// - Aide rapide : boutons prédéfinis (quotidien, hebdomadaire, mensuel)
// Librairie recommandée : cronstrue (npm install cronstrue)
```

---

### 🟡 IMPORTANT 11 — Ajout des routes Phase 9 dans le router

**Compléter `internal/app/router.go` :**

```go
// Scheduling
r.Get("/projects/{id}/schedule",      schedHandler.GetSchedule)
r.Put("/projects/{id}/schedule",      schedHandler.UpsertSchedule)
r.Delete("/projects/{id}/schedule",   schedHandler.DeleteSchedule)

// Retry policy
r.Get("/projects/{id}/retry-policy",  retryHandler.GetPolicy)
r.Put("/projects/{id}/retry-policy",  retryHandler.UpsertPolicy)

// Queue & Worker monitoring
r.Get("/queue",                       schedHandler.GetQueueStatus)
r.Get("/worker/status",               workerHandler.Status)
```

---

### 🟡 IMPORTANT 12 — `cmd/worker/` : binaire worker séparé (optionnel MVP)

Le `README-projet-etapes.md` mentionne `cmd/worker/` comme point d'entrée autonome.
Pour le MVP, le worker est embarqué dans le serveur. Pour les environnements de production,
il peut être extrait comme binaire distinct.

**`cmd/worker/main.go` — binaire autonome :**

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    // ... imports identiques à cmd/server/main.go pour le sous-ensemble worker
)

// Ce binaire ne démarre que le Worker et le Scheduler — pas le serveur HTTP.
// Il se connecte à la même PostgreSQL et au même répertoire XML.
// Permet le scaling horizontal : N replicas worker sans N replicas serveur HTTP.
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()
    // ... wiring identique, sans http.Server
    go w.Start(ctx)
    if err := sched.Start(ctx); err != nil {
        log.Fatal().Err(err).Msg("scheduler: démarrage échoué")
    }
    <-ctx.Done()
}
```

---

## Plan d'action pour finaliser la Phase 9

### Sprint A — Queue + Worker async (1 jour)

- [ ] Créer `internal/worker/queue.go` (chan bufferisé + stats)
- [ ] Créer `internal/worker/worker.go` (pool goroutines + semaphore)
- [ ] Refactorer `internal/orchestrator/service.go` : `Enqueue()` + `Execute()` séparés
- [ ] Adapter `api/handlers/run_handler.go` → répondre `202 Accepted` + `runID`
- [ ] Vérifier `go build ./...` passe

### Sprint B — Retry engine (0,5 jour)

- [ ] Créer `internal/worker/retry.go` (`RetryPolicy` + `RunWithRetry()`)
- [ ] Intégrer `RunWithRetry` dans `worker.execute()`
- [ ] Créer `migrations/007_retry_policy.sql`
- [ ] Implémenter `RetryPolicyRepository` (PostgreSQL) et les handlers GET/PUT

### Sprint C — Scheduler cron (1 jour)

- [ ] Implémenter `internal/etl/scheduler/scheduler.go` (`robfig/cron/v3`)
- [ ] Implémenter `scheduler.Repository` PostgreSQL (`ListActive`, `Upsert`, `Delete`, `UpdateNextRun`)
- [ ] Créer `migrations/006_schedules_full.sql`
- [ ] Câbler `sched.Start(ctx)` dans `cmd/server/main.go`
- [ ] Vérifier que le scheduler charge et active les crons au démarrage

### Sprint D — Handlers & Router (0,5 jour)

- [ ] Créer `api/handlers/schedule_handler.go`
- [ ] Créer `api/handlers/worker_handler.go`
- [ ] Ajouter toutes les routes Phase 9 dans `internal/app/router.go`
- [ ] Injecter les nouveaux handlers dans le bootstrap
- [ ] Compléter `pkg/dto/schedule.go`

### Sprint E — Configuration & Graceful shutdown (0,5 jour)

- [ ] Ajouter `worker.concurrency`, `worker.queue_capacity`, `scheduler.enabled`
  dans les 3 fichiers YAML (`config.dev.yaml`, `config.preprod.yaml`, `config.prod.yaml`)
- [ ] Implémenter le graceful shutdown : `http.Server.Shutdown()` + attendre `worker.wg.Wait()`
- [ ] Vérifier que `SIGTERM` arrête proprement le serveur sans couper un run en cours

### Sprint F — Tests (1 jour)

- [ ] `tests/unit/worker/queue_test.go` : Enqueue OK, Enqueue full → erreur, Stats cohérentes
- [ ] `tests/unit/worker/retry_test.go` : succès immédiat, 3 échecs puis succès, MaxAttempts épuisé, ctx annulé
- [ ] `tests/unit/scheduler/scheduler_test.go` : Upsert, register, Delete, cron déclenche un Enqueue
- [ ] `tests/integration/schedule_api_test.go` : PUT schedule → GET schedule → vérifier NextRun calculé
- [ ] `tests/integration/run_async_test.go` : POST run → 202 Accepted → poll GET run → succeeded

### Sprint G — UI Scheduling (1 jour)

- [ ] Créer `web/ui/src/components/scheduling/CronInput.tsx` avec validation + human-readable
- [ ] Créer `web/ui/src/pages/SchedulePage.tsx`
- [ ] Ajouter la route `/projects/:projectId/schedule` dans `App.tsx`
- [ ] Ajouter bouton "Planifier" sur `ProjectsPage` / `PipelineDesigner` → lien vers `SchedulePage`
- [ ] Vérifier `npm run build` passe

---

## Checklist finale Phase 9 — "Definition of Done"

### Backend Go

- [ ] `go build ./...` passe sans erreur ni warning
- [ ] `go vet ./...` passe proprement
- [ ] `POST /api/v1/projects/{id}/runs` répond `202 Accepted` avec `runID` (non bloquant)
- [ ] `GET /api/v1/runs/{runID}` retourne le statut `pending` → `running` → `succeeded/failed`
- [ ] `PUT /api/v1/projects/{id}/schedule` active le cron et calcule `nextRun`
- [ ] Le scheduler déclenche un run automatiquement selon l'expression cron configurée
- [ ] `RetryPolicy` est respectée : un run en échec est relancé N fois avec backoff
- [ ] `GET /api/v1/queue` retourne le nombre de jobs en attente
- [ ] `GET /api/v1/worker/status` retourne la concurrence et les goroutines actives
- [ ] Graceful shutdown : `SIGTERM` attend la fin du run en cours avant d'arrêter

### Tests

- [ ] `go test ./tests/unit/worker/...` vert
- [ ] `go test ./tests/unit/scheduler/...` vert
- [ ] `go test ./tests/integration/...` vert (avec DB de test)
- [ ] Run end-to-end planifié : schedule cron → run automatique → `succeeded` en base

### Frontend React

- [ ] `npm run build` passe sans erreur
- [ ] `SchedulePage` affiche le cron, le prochain run, et permet d'activer/désactiver
- [ ] `CronInput` valide l'expression et affiche la description human-readable
- [ ] L'UI affiche le statut du run en temps réel (WebSocket) pour les runs planifiés

### Déploiement

- [ ] `docker-compose up` démarre server + worker + postgres sans erreur
- [ ] Migrations `006` et `007` s'appliquent automatiquement
- [ ] Les variables `worker.concurrency` et `scheduler.enabled` sont lues depuis les YAML d'env

---

## Architecture rappel — Flux d'orchestration Phase 9
```
┌─────────────────────────────────────────────────────────┐
│ Déclencheurs │
│ │
│ UI / HTTP Client Scheduler (cron) │
│ POST /projects/{id}/runs robfig/cron.Cron │
│ │ │ │
└─────────┼───────────────────────┼────────────────────────┘
│ │
▼ ▼
orchestrator.Enqueue() orchestrator.Enqueue()
│ │
└───────────┬───────────┘
│ jobs.Create() → status=pending
▼
worker.Queue (chan Job, capacity=50)
│
┌────────────┼────────────┐
▼ ▼ ▼ ← concurrency = 3
goroutine 1 goroutine 2 goroutine 3
│
worker.execute(job)
│ jobRepo.SetStatus("running")
│
RunWithRetry(policy, fn)
│
orchestrator.Execute(ctx, projectID)
│ xmlStore.Load(projectID)
│ parser.Parse(xml)
│ engine.Execute(dag)
│ └─ bloc1 → bloc2 → bloc3 (topologique)
│ └─ jobRepo.AppendLog() par bloc
│
jobRepo.SetStatus("succeeded" | "failed")
│
▼
WebSocket SSE → UI (suivi temps réel)
```

---

## Fichiers impactés — récapitulatif

| Fichier | Action | Priorité |
|---|---|---|
| `internal/worker/queue.go` | **CRÉER** — Queue bufferisée + stats | 🔴 BLOQUANT |
| `internal/worker/worker.go` | **CRÉER** — Pool goroutines + semaphore | 🔴 BLOQUANT |
| `internal/worker/retry.go` | **CRÉER** — `RetryPolicy` + `RunWithRetry()` | 🔴 BLOQUANT |
| `internal/orchestrator/service.go` | Refactorer : `Enqueue()` + `Execute()` séparés | 🔴 BLOQUANT |
| `api/handlers/run_handler.go` | Adapter `StartRun` → `202 Accepted` | 🔴 BLOQUANT |
| `internal/etl/scheduler/scheduler.go` | **IMPLÉMENTER** — robfig/cron + repo + Start() | 🔴 BLOQUANT |
| `api/handlers/schedule_handler.go` | **CRÉER** — GET/PUT/DELETE schedule + queue stats | 🔴 BLOQUANT |
| `api/handlers/worker_handler.go` | **CRÉER** — GET /worker/status | 🔴 BLOQUANT |
| `migrations/006_schedules_full.sql` | **CRÉER** — enrichissement table schedules | 🔴 BLOQUANT |
| `migrations/007_retry_policy.sql` | **CRÉER** — table retry_policies | 🔴 BLOQUANT |
| `internal/app/router.go` | Ajouter routes scheduling + worker | 🔴 BLOQUANT |
| `cmd/server/main.go` | Câbler Queue + Worker + Scheduler au boot | 🔴 BLOQUANT |
| `configs/config.*.yaml` | Ajouter `worker.*` + `scheduler.*` | 🟡 IMPORTANT |
| `pkg/dto/schedule.go` | **CRÉER** — ScheduleRequest/Response + RetryPolicyRequest | 🟡 IMPORTANT |
| `cmd/worker/main.go` | **CRÉER** — binaire worker autonome (optionnel MVP) | 🟡 IMPORTANT |
| `web/ui/src/pages/SchedulePage.tsx` | **CRÉER** — page scheduling | 🟡 IMPORTANT |
| `web/ui/src/components/scheduling/CronInput.tsx` | **CRÉER** — input cron validé + human-readable | 🟡 IMPORTANT |
| `web/ui/src/App.tsx` | Ajouter route `/projects/:id/schedule` | 🟡 IMPORTANT |
| `tests/unit/worker/queue_test.go` | **CRÉER** | 🟡 IMPORTANT |
| `tests/unit/worker/retry_test.go` | **CRÉER** | 🟡 IMPORTANT |
| `tests/unit/scheduler/scheduler_test.go` | **CRÉER** | 🟡 IMPORTANT |
| `tests/integration/schedule_api_test.go` | **CRÉER** | 🟡 IMPORTANT |
| `tests/integration/run_async_test.go` | **CRÉER** | 🟡 IMPORTANT |

---

*Document généré automatiquement par analyse du code source — à mettre à jour à chaque sprint.*