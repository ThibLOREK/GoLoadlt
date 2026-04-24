# Étape 10 — Qualité et Exploitation : État détaillé et plan d'implémentation

> Généré le 2026-04-24 · Basé sur un scan complet du code source et la feuille de route Phase 10

---

## Résumé de la Phase 10

La Phase 10 a pour objectif de consolider la plateforme GoLoadIt autour de quatre axes majeurs :
1. **Tests e2e** sur les blocs et pipelines clés
2. **Profiling mémoire** pour supporter de gros volumes en streaming
3. **Observabilité complète** avec traces distribuées par run et par bloc
4. **Versionnement XML** et rollback de projet

**État global : phases 1-9 supposées stables ✅ — Phase 10 à implémenter intégralement 🔴**

---

## Ce qui est déjà en place (Phases 0 → 9)

### ✅ Infrastructure & Foundation
- Structure repo complète : `cmd/`, `internal/`, `pkg/`, `api/`, `web/`, `deploy/`, `migrations/`
- Config multi-env YAML (`config.dev.yaml`, `config.preprod.yaml`, `config.prod.yaml`)
- Docker Compose, Makefile fonctionnel
- Logger (`zerolog`), auth JWT, middleware, migrations SQL

### ✅ Contracts & Moteur DAG
- `contracts/block.go` : `DataType`, `ColumnDef`, `Schema`, `DataRow`, `Port`, `BlockContext`, `Block`, `BlockFactory`
- `engine/dag.go` : `BuildDAG()` avec tri topologique
- `engine/executor.go` : `Execute()` — `RunResult`, `ExecutionReport`
- `engine/inject_connections.go` : injection connexions dans `BlockContext`

### ✅ Blocs Sources, Transforms, Targets (Phases 4 & 5)
Tous les blocs MVP sont présents et enregistrés dans le catalogue.

### ✅ Connexions multi-env (Phase 6)
- CRUD XML connexions, résolution `ACTIVE_ENV`, secrets via env vars
- `connections/manager/`, `connections/resolver/`, `connections/secrets/`

### ✅ API complète (Phase 7)
- CRUD projets + connexions + runs
- Historique, logs d'exécution, switch d'env global
- Documentation OpenAPI générée

### ✅ UI React Flow (Phase 8)
- Canvas blocs + palette + `NodeConfigPanel`
- Suivi temps réel via WebSocket (bloc par bloc)
- `DataPreviewPanel` opérationnel

### ✅ Scheduling & Worker asynchrone (Phase 9)
- Planification cron par projet (`internal/etl/scheduler/`)
- Worker `cmd/worker/` avec file d'attente
- Retry policy configurable, limitation de concurrence

---

## Axes de la Phase 10 — Détail technique

---

### Axe 1 — Tests End-to-End (e2e)

#### Objectif
Valider des pipelines complets de bout en bout : de la lecture source à l'écriture target, en passant par plusieurs blocs de transformation enchaînés, avec assertions sur les données produites.

#### Architecture des tests e2e
```
tests/
└── e2e/
├── helpers/
│ ├── runner.go # Lance un pipeline XML → attend ExecutionReport
│ ├── pg_fixture.go # Démarre un conteneur PostgreSQL via testcontainers-go
│ └── csv_fixture.go # Génère des fichiers CSV de test temporaires
├── pipelines/
│ ├── filter_pipeline_test.go
│ ├── join_pipeline_test.go
│ ├── split_pipeline_test.go
│ ├── aggregate_pipeline_test.go
│ └── full_pipeline_test.go # CSV → filter → add_column → aggregate → postgres
└── fixtures/
├── projects/
│ ├── proj_filter.xml
│ ├── proj_join.xml
│ ├── proj_split.xml
│ └── proj_full.xml
└── csv/
├── sales.csv
└── products.csv
```

#### `tests/e2e/helpers/runner.go`

```go
package helpers

import (
    "context"
    "testing"

    "github.com/ThibLOREK/GoLoadIt/internal/etl/engine"
    "github.com/ThibLOREK/GoLoadIt/internal/xml/parser"
    "github.com/ThibLOREK/GoLoadIt/internal/xml/store"
    "github.com/stretchr/testify/require"
)

// PipelineRunner exécute un projet XML et retourne le rapport d'exécution.
type PipelineRunner struct {
    XMLStore *store.XMLStore
    Executor *engine.Executor
}

func NewPipelineRunner(baseDir string) *PipelineRunner {
    return &PipelineRunner{
        XMLStore: store.New(baseDir),
        Executor: engine.NewExecutor(),
    }
}

func (r *PipelineRunner) Run(t *testing.T, projectID string) *engine.ExecutionReport {
    t.Helper()
    project, err := r.XMLStore.Load(projectID)
    require.NoError(t, err, "chargement XML projet %s", projectID)

    report, err := r.Executor.Execute(context.Background(), project)
    require.NoError(t, err, "exécution pipeline %s", projectID)
    return report
}
```

#### `tests/e2e/helpers/pg_fixture.go`

```go
package helpers

import (
    "context"
    "fmt"
    "testing"

    "github.com/jackc/pgx/v5"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/stretchr/testify/require"
)

type PGFixture struct {
    Container testcontainers.Container
    DSN       string
}

// NewPGFixture démarre un conteneur PostgreSQL éphémère pour les tests e2e.
func NewPGFixture(t *testing.T) *PGFixture {
    t.Helper()
    ctx := context.Background()

    pgc, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:16-alpine"),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(/* readiness check */),
    )
    require.NoError(t, err)

    host, _ := pgc.Host(ctx)
    port, _ := pgc.MappedPort(ctx, "5432")
    dsn := fmt.Sprintf("postgres://test:test@%s:%s/testdb", host, port.Port())

    t.Cleanup(func() { _ = pgc.Terminate(ctx) })
    return &PGFixture{Container: pgc, DSN: dsn}
}

// QueryRows exécute une requête SELECT et retourne les lignes sous forme de maps.
func (f *PGFixture) QueryRows(t *testing.T, query string) []map[string]any {
    t.Helper()
    conn, err := pgx.Connect(context.Background(), f.DSN)
    require.NoError(t, err)
    defer conn.Close(context.Background())

    rows, err := conn.Query(context.Background(), query)
    require.NoError(t, err)
    defer rows.Close()

    var result []map[string]any
    for rows.Next() {
        vals, _ := rows.Values()
        fd := rows.FieldDescriptions()
        row := make(map[string]any, len(fd))
        for i, f := range fd {
            row[string(f.Name)] = vals[i]
        }
        result = append(result, row)
    }
    return result
}
```

#### Pipelines e2e prioritaires

| Test | Pipeline XML | Assertion clé |
|---|---|---|
| `TestFilter_E2E` | `source.csv → transform.filter → target.csv` | Lignes filtrées == attendu |
| `TestJoin_E2E` | `source.csv × source.csv → transform.join → target.csv` | Jointure inner correcte |
| `TestSplit_E2E` | `source.csv → transform.split → (target.csv × 2)` | 2 fichiers distincts |
| `TestAggregate_E2E` | `source.csv → transform.aggregate → target.postgres` | Sommes correctes en base |
| `TestFull_E2E` | Pipeline 5 blocs enchaînés | `ExecutionReport.Success == true`, 0 erreur |

#### Commandes

```bash
# Lancer tous les tests e2e (nécessite Docker)
go test ./tests/e2e/... -v -tags e2e -timeout 120s

# Lancer un test spécifique
go test ./tests/e2e/pipelines/ -run TestFull_E2E -v
```

---

### Axe 2 — Profiling Mémoire (Gros Volumes)

#### Objectif
Garantir que le moteur d'exécution traite les flux de données **en streaming pur** (un `DataRow` à la fois via les channels Go), sans jamais matérialiser l'intégralité du dataset en mémoire.

#### Contrat de streaming — `contracts/block.go`

La règle fondamentale est que chaque bloc lit depuis `Port.Ch` et écrit dans `Port.Ch` sans buffer intermédiaire. Le profiling vérifie que cette règle est respectée.

```go
// DataRow est la donnée unitaire qui transite entre blocs.
// Elle NE DOIT PAS être accumulée en slice entre deux blocs.
type DataRow map[string]any

// Port est le canal qui relie deux blocs dans le DAG.
type Port struct {
    ID     string
    Schema Schema
    Ch     chan DataRow // capacité configurable, jamais bufferisé à l'infini
}
```

#### `internal/etl/engine/executor.go` — Ajout du contexte de profiling

```go
// ExecutionOptions configure le comportement de l'exécuteur.
type ExecutionOptions struct {
    EnableProfiling bool          // active pprof pendant l'exécution
    ProfilePath     string        // dossier de sortie des profils (.pb.gz)
    ChannelBuffer   int           // capacité des channels entre blocs (défaut: 64)
    MaxMemoryMB     int           // seuil alerte mémoire (0 = désactivé)
}

// Execute exécute le DAG avec les options spécifiées.
func (e *Executor) ExecuteWithOptions(
    ctx context.Context,
    project *contracts.Project,
    opts ExecutionOptions,
) (*ExecutionReport, error) {
    if opts.EnableProfiling {
        stopCPU := e.startCPUProfile(opts.ProfilePath)
        defer stopCPU()
        defer e.writeMemProfile(opts.ProfilePath)
    }
    if opts.MaxMemoryMB > 0 {
        go e.watchMemory(ctx, opts.MaxMemoryMB)
    }
    return e.execute(ctx, project, opts.ChannelBuffer)
}
```

#### `internal/etl/engine/profiler.go` (nouveau fichier)

```go
package engine

import (
    "fmt"
    "os"
    "path/filepath"
    "runtime"
    "runtime/pprof"
    "time"
)

func (e *Executor) startCPUProfile(dir string) func() {
    _ = os.MkdirAll(dir, 0o755)
    f, err := os.Create(filepath.Join(dir, fmt.Sprintf("cpu_%d.pb.gz", time.Now().Unix())))
    if err != nil {
        return func() {}
    }
    _ = pprof.StartCPUProfile(f)
    return func() {
        pprof.StopCPUProfile()
        f.Close()
    }
}

func (e *Executor) writeMemProfile(dir string) {
    _ = os.MkdirAll(dir, 0o755)
    f, err := os.Create(filepath.Join(dir, fmt.Sprintf("mem_%d.pb.gz", time.Now().Unix())))
    if err != nil {
        return
    }
    defer f.Close()
    runtime.GC()
    _ = pprof.WriteHeapProfile(f)
}

// watchMemory émet un log d'alerte si l'utilisation mémoire dépasse maxMB.
func (e *Executor) watchMemory(ctx context.Context, maxMB int) {
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            var ms runtime.MemStats
            runtime.ReadMemStats(&ms)
            heapMB := int(ms.HeapInuse / 1024 / 1024)
            if heapMB > maxMB {
                e.logger.Warn().
                    Int("heap_mb", heapMB).
                    Int("limit_mb", maxMB).
                    Msg("ALERTE mémoire : seuil dépassé pendant l'exécution")
            }
        }
    }
}
```

#### Benchmark de référence — `tests/benchmarks/engine_bench_test.go`

```go
package benchmarks_test

import (
    "context"
    "testing"

    "github.com/ThibLOREK/GoLoadIt/internal/etl/engine"
    "github.com/ThibLOREK/GoLoadIt/tests/e2e/helpers"
)

// BenchmarkPipeline_1M simule un pipeline filter+aggregate sur 1 million de lignes.
func BenchmarkPipeline_1M(b *testing.B) {
    runner := helpers.NewPipelineRunner("fixtures/projects")
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        opts := engine.ExecutionOptions{
            ChannelBuffer: 256,
            MaxMemoryMB:   512,
        }
        _, err := runner.Executor.ExecuteWithOptions(context.Background(), syntheticProject(1_000_000), opts)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

#### Commandes de profiling

```bash
# Lancer le benchmark avec profiling
go test ./tests/benchmarks/ -bench=BenchmarkPipeline_1M -benchmem \
  -cpuprofile=profiles/cpu.pb.gz -memprofile=profiles/mem.pb.gz

# Analyser le profil mémoire interactivement
go tool pprof -http=:8080 profiles/mem.pb.gz

# Analyser le profil CPU
go tool pprof -http=:8081 profiles/cpu.pb.gz
```

---

### Axe 3 — Observabilité Complète (Traces par Run)

#### Objectif
Tracer chaque run et chaque bloc individuellement via **OpenTelemetry**, exposer les métriques Prometheus, et garantir la visibilité complète dans Grafana/Jaeger.

#### `internal/telemetry/tracer.go` — Configuration OpenTelemetry

```go
package telemetry

import (
    "context"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

const ServiceName = "goloadit"

// InitTracer initialise l'exporteur OTLP et retourne une fonction de shutdown.
func InitTracer(ctx context.Context, endpoint string) (func(context.Context) error, error) {
    exp, err := otlptracehttp.New(ctx,
        otlptracehttp.WithEndpoint(endpoint),
        otlptracehttp.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }

    res, _ := resource.New(ctx,
        resource.WithAttributes(semconv.ServiceName(ServiceName)),
    )

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exp),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
    )
    otel.SetTracerProvider(tp)
    return tp.Shutdown, nil
}
```

#### `internal/etl/engine/executor.go` — Instrumentation par bloc

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("goloadit/engine")

// runBlock exécute un bloc individuel dans une span OpenTelemetry dédiée.
func (e *Executor) runBlock(ctx context.Context, b contracts.Block, bctx *contracts.BlockContext) error {
    ctx, span := tracer.Start(ctx, "block."+bctx.BlockType,
        trace.WithAttributes(
            attribute.String("block.id", bctx.BlockID),
            attribute.String("block.type", bctx.BlockType),
            attribute.String("run.id", bctx.RunID),
        ),
    )
    defer span.End()

    err := b.Run(bctx)
    if err != nil {
        span.RecordError(err)
    }

    // Attacher les métriques de sortie à la span
    span.SetAttributes(
        attribute.Int64("block.rows_in", bctx.Metrics.RowsIn),
        attribute.Int64("block.rows_out", bctx.Metrics.RowsOut),
        attribute.Int64("block.duration_ms", bctx.Metrics.DurationMs),
    )
    return err
}
```

#### `contracts/block.go` — Ajout de `BlockMetrics`

```go
// BlockMetrics collecte les statistiques d'exécution d'un bloc.
// Renseigné par l'executor après chaque Run().
type BlockMetrics struct {
    RowsIn     int64
    RowsOut    int64
    RowsError  int64
    DurationMs int64
    MemoryBytes int64
}

// BlockContext — ajout du champ Metrics (non breaking, champ pointer)
type BlockContext struct {
    Ctx       context.Context
    BlockID   string
    BlockType string
    RunID     string
    Params    map[string]string
    Inputs    []*Port
    Outputs   []*Port
    ConnEnvs  map[string]ConnectionEnv
    Metrics   *BlockMetrics  // ← NOUVEAU — injecté par l'executor avant Run()
    Logger    zerolog.Logger
}
```

#### `internal/telemetry/metrics.go` — Métriques Prometheus

```go
package telemetry

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    RunsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "goloadit_runs_total",
        Help: "Nombre total de runs par statut (succeeded/failed/cancelled)",
    }, []string{"project_id", "status"})

    RunDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "goloadit_run_duration_seconds",
        Help:    "Durée d'exécution des runs en secondes",
        Buckets: prometheus.DefBuckets,
    }, []string{"project_id"})

    BlockRowsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "goloadit_block_rows_total",
        Help: "Lignes traitées par bloc (in/out)",
    }, []string{"project_id", "block_type", "direction"})

    BlockErrors = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "goloadit_block_errors_total",
        Help: "Erreurs par bloc",
    }, []string{"project_id", "block_type"})

    ActiveRuns = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "goloadit_active_runs",
        Help: "Nombre de runs en cours d'exécution",
    })
)
```

#### `internal/orchestrator/service.go` — Instrumentation des runs

```go
// RunProject — version instrumentée Phase 10
func (s *Service) RunProject(ctx context.Context, projectID string) (*engine.ExecutionReport, error) {
    ctx, span := tracer.Start(ctx, "orchestrator.RunProject",
        trace.WithAttributes(attribute.String("project.id", projectID)),
    )
    defer span.End()

    telemetry.ActiveRuns.Inc()
    defer telemetry.ActiveRuns.Dec()

    timer := prometheus.NewTimer(telemetry.RunDuration.WithLabelValues(projectID))
    defer timer.ObserveDuration()

    project, err := s.xmlStore.Load(projectID)
    if err != nil {
        span.RecordError(err)
        return nil, err
    }

    run, err := s.jobRepo.Create(ctx, projectID)
    if err != nil {
        return nil, err
    }
    _ = s.jobRepo.SetStatus(ctx, run.ID, "running")

    report, execErr := s.executor.ExecuteWithOptions(ctx, project, engine.ExecutionOptions{
        ChannelBuffer: 64,
        MaxMemoryMB:   1024,
    })

    status := "succeeded"
    if execErr != nil {
        status = "failed"
        span.RecordError(execErr)
    }
    _ = s.jobRepo.SetStatus(ctx, run.ID, status)
    telemetry.RunsTotal.WithLabelValues(projectID, status).Inc()

    // Propager les métriques par bloc vers Prometheus
    if report != nil {
        for _, r := range report.Results {
            telemetry.BlockRowsProcessed.WithLabelValues(projectID, r.BlockType, "in").Add(float64(r.RowsIn))
            telemetry.BlockRowsProcessed.WithLabelValues(projectID, r.BlockType, "out").Add(float64(r.RowsOut))
            if r.Error != nil {
                telemetry.BlockErrors.WithLabelValues(projectID, r.BlockType).Inc()
            }
        }
    }
    return report, execErr
}
```

#### Endpoint `/metrics` — `api/handlers/health_handler.go`

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

// MountMetrics expose le endpoint Prometheus standard.
func MountMetrics(r chi.Router) {
    r.Handle("/metrics", promhttp.Handler())
}
```

#### `deploy/docker/docker-compose.observability.yml` (nouveau fichier)

```yaml
version: "3.9"
services:
  jaeger:
    image: jaegertracing/all-in-one:1.57
    ports:
      - "16686:16686"  # UI Jaeger
      - "4318:4318"    # OTLP HTTP

  prometheus:
    image: prom/prometheus:v2.51.0
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9090:9090"

  grafana:
    image: grafana/grafana:10.4.0
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - ./grafana/dashboards:/var/lib/grafana/dashboards
      - ./grafana/provisioning:/etc/grafana/provisioning
```

---

### Axe 4 — Versionnement XML et Rollback de Projet

#### Objectif
Permettre de **lister l'historique des versions** d'un projet, d'**inspecter un état passé** et de **rollback** vers n'importe quelle version antérieure, sans perte de l'historique.

#### Modèle de versionning sur disque (déjà défini en Phase 2)
```
projects/
└── {project-id}/
├── project.xml # version courante (source de vérité)
└── history/
├── v1.xml
├── v2.xml
└── v{n}.xml # chaque save crée une nouvelle version
```

#### `internal/xml/store/store.go` — Méthodes de versionnement

```go
package store

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/xml"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
    "time"

    "github.com/ThibLOREK/GoLoadIt/internal/etl/contracts"
)

type XMLStore struct {
    BaseDir string
}

func New(baseDir string) *XMLStore {
    return &XMLStore{BaseDir: baseDir}
}

// ProjectVersion représente une entrée dans l'historique des versions.
type ProjectVersion struct {
    Version   int       `json:"version"`
    SavedAt   time.Time `json:"saved_at"`
    SHA256    string    `json:"sha256"`
    SizeBytes int64     `json:"size_bytes"`
}

// Save sérialise le projet, archive la version précédente, calcule le SHA256.
func (s *XMLStore) Save(project *contracts.Project) error {
    projectDir := filepath.Join(s.BaseDir, project.ID)
    historyDir := filepath.Join(projectDir, "history")
    currentPath := filepath.Join(projectDir, "project.xml")

    _ = os.MkdirAll(historyDir, 0o755)

    // Archiver la version courante si elle existe
    if _, err := os.Stat(currentPath); err == nil {
        nextVersion := s.nextVersionNumber(historyDir)
        archivePath := filepath.Join(historyDir, fmt.Sprintf("v%d.xml", nextVersion))
        if err := copyFile(currentPath, archivePath); err != nil {
            return fmt.Errorf("archivage version %d: %w", nextVersion, err)
        }
    }

    // Sérialiser la nouvelle version
    data, err := xml.MarshalIndent(project, "", "  ")
    if err != nil {
        return fmt.Errorf("sérialisation XML: %w", err)
    }
    data = append([]byte(xml.Header), data...)

    if err := os.WriteFile(currentPath, data, 0o644); err != nil {
        return fmt.Errorf("écriture project.xml: %w", err)
    }

    // Écrire le SHA256 dans un fichier sidecar
    hash := sha256sum(data)
    return os.WriteFile(currentPath+".sha256", []byte(hash), 0o644)
}

// Load charge et parse la version courante du projet.
func (s *XMLStore) Load(projectID string) (*contracts.Project, error) {
    path := filepath.Join(s.BaseDir, projectID, "project.xml")
    f, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("projet introuvable %s: %w", projectID, err)
    }
    defer f.Close()
    return parseReader(f)
}

// ListVersions retourne l'historique trié des versions du projet.
func (s *XMLStore) ListVersions(projectID string) ([]ProjectVersion, error) {
    historyDir := filepath.Join(s.BaseDir, projectID, "history")
    entries, err := os.ReadDir(historyDir)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, nil
        }
        return nil, err
    }

    var versions []ProjectVersion
    for _, e := range entries {
        if e.IsDir() || !strings.HasSuffix(e.Name(), ".xml") {
            continue
        }
        numStr := strings.TrimSuffix(strings.TrimPrefix(e.Name(), "v"), ".xml")
        n, err := strconv.Atoi(numStr)
        if err != nil {
            continue
        }
        info, _ := e.Info()
        data, _ := os.ReadFile(filepath.Join(historyDir, e.Name()))
        versions = append(versions, ProjectVersion{
            Version:   n,
            SavedAt:   info.ModTime(),
            SHA256:    sha256sum(data),
            SizeBytes: info.Size(),
        })
    }
    sort.Slice(versions, func(i, j int) bool {
        return versions[i].Version < versions[j].Version
    })
    return versions, nil
}

// LoadVersion charge une version spécifique depuis history/v{n}.xml.
func (s *XMLStore) LoadVersion(projectID string, version int) (*contracts.Project, error) {
    path := filepath.Join(s.BaseDir, projectID, "history", fmt.Sprintf("v%d.xml", version))
    f, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("version %d introuvable pour projet %s: %w", version, projectID, err)
    }
    defer f.Close()
    return parseReader(f)
}

// Rollback restaure une version passée comme version courante (archive d'abord la courante).
func (s *XMLStore) Rollback(projectID string, version int) error {
    target, err := s.LoadVersion(projectID, version)
    if err != nil {
        return err
    }
    // Save() archive automatiquement la version courante avant d'écrire
    return s.Save(target)
}

// --- helpers privés ---

func (s *XMLStore) nextVersionNumber(historyDir string) int {
    entries, _ := os.ReadDir(historyDir)
    max := 0
    for _, e := range entries {
        numStr := strings.TrimSuffix(strings.TrimPrefix(e.Name(), "v"), ".xml")
        if n, err := strconv.Atoi(numStr); err == nil && n > max {
            max = n
        }
    }
    return max + 1
}

func sha256sum(data []byte) string {
    h := sha256.Sum256(data)
    return hex.EncodeToString(h[:])
}

func copyFile(src, dst string) error {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer out.Close()
    _, err = io.Copy(out, in)
    return err
}

func parseReader(r io.Reader) (*contracts.Project, error) {
    var p contracts.Project
    return &p, xml.NewDecoder(r).Decode(&p)
}
```

#### Nouveaux endpoints API — `api/handlers/project_handler.go`

```go
// GET /api/v1/projects/{id}/versions
// → retourne []ProjectVersion (historique complet)
func (h *ProjectHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
    projectID := chi.URLParam(r, "id")
    versions, err := h.xmlStore.ListVersions(projectID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, versions)
}

// GET /api/v1/projects/{id}/versions/{version}
// → retourne le contenu XML de la version demandée
func (h *ProjectHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
    projectID := chi.URLParam(r, "id")
    version, err := strconv.Atoi(chi.URLParam(r, "version"))
    if err != nil {
        http.Error(w, "version invalide", http.StatusBadRequest)
        return
    }
    project, err := h.xmlStore.LoadVersion(projectID, version)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    writeJSON(w, project)
}

// POST /api/v1/projects/{id}/rollback
// Body: { "version": 3 }
// → restaure la version demandée comme version courante
func (h *ProjectHandler) Rollback(w http.ResponseWriter, r *http.Request) {
    projectID := chi.URLParam(r, "id")
    var body struct {
        Version int `json:"version"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Version < 1 {
        http.Error(w, "version invalide", http.StatusBadRequest)
        return
    }
    if err := h.xmlStore.Rollback(projectID, body.Version); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
    writeJSON(w, map[string]string{"status": "rolled_back", "version": strconv.Itoa(body.Version)})
}
```

#### Enregistrement des routes — `api/handlers/router.go`

```go
r.Route("/api/v1/projects/{id}", func(r chi.Router) {
    r.Get("/", h.GetProject)
    r.Put("/", h.SaveProject)
    r.Delete("/", h.DeleteProject)

    // Phase 10 — Versionning
    r.Get("/versions", h.ListVersions)
    r.Get("/versions/{version}", h.GetVersion)
    r.Post("/rollback", h.Rollback)
})
```

---

### Axe 5 — Runbooks d'Exploitation

#### `docs/runbooks/01-demarrage.md` (structure cible)

```markdown
# Runbook — Démarrage de la plateforme GoLoadIt

## Prérequis
- Docker 24+ et Docker Compose V2
- Go 1.24+ (dev uniquement)
- Make

## Démarrage complet

```bash
# 1. Copier la configuration
cp .env.example .env
# Renseigner les valeurs de .env

# 2. Démarrer les services
docker-compose up -d

# 3. Appliquer les migrations
make migrate

# 4. Vérifier la santé
curl http://localhost:8080/healthz
# Attendu: {"status":"ok"}

# 5. (Optionnel) Démarrer la stack d'observabilité
docker-compose -f deploy/docker/docker-compose.observability.yml up -d
# UI Jaeger: http://localhost:16686
# Prometheus: http://localhost:9090
# Grafana: http://localhost:3001 (admin/admin)
```

## Procédures d'urgence

### Pipeline bloqué
```bash
# Annuler un run en cours
curl -X POST http://localhost:8080/api/v1/runs/{run_id}/cancel

# Vérifier les logs du worker
docker-compose logs -f worker --tail=100
```

### Rollback d'un projet
```bash
# Lister les versions disponibles
curl http://localhost:8080/api/v1/projects/{project_id}/versions

# Revenir à la version 3
curl -X POST http://localhost:8080/api/v1/projects/{project_id}/rollback \
  -H "Content-Type: application/json" \
  -d '{"version": 3}'
```
```

---

## Plan d'action Phase 10

### Sprint A — Tests e2e (2 jours)
- [ ] Implémenter `tests/e2e/helpers/runner.go`, `pg_fixture.go`, `csv_fixture.go`
- [ ] Écrire les 5 pipelines XML de fixture dans `tests/e2e/fixtures/projects/`
- [ ] Implémenter les 5 tests e2e prioritaires
- [ ] Ajouter `go test ./tests/e2e/... -tags e2e` dans le CI GitHub Actions
- [ ] Configurer `testcontainers-go` dans `go.mod`

### Sprint B — Profiling mémoire (1 jour)
- [ ] Créer `internal/etl/engine/profiler.go`
- [ ] Ajouter `BlockMetrics` dans `contracts/block.go`
- [ ] Modifier `executor.go` pour injecter et collecter `BlockMetrics` par bloc
- [ ] Ajouter `ExecutionOptions` avec `EnableProfiling`, `MaxMemoryMB`, `ChannelBuffer`
- [ ] Écrire `tests/benchmarks/engine_bench_test.go` (1M lignes)
- [ ] Documenter les résultats du benchmark dans `docs/architecture/profiling.md`

### Sprint C — Observabilité complète (2 jours)
- [ ] Configurer `internal/telemetry/tracer.go` (OTLP HTTP vers Jaeger)
- [ ] Créer `internal/telemetry/metrics.go` (compteurs Prometheus)
- [ ] Instrumenter `engine/executor.go` : une span par bloc avec `block.rows_in/out`
- [ ] Instrumenter `orchestrator/service.go` : span globale du run + push métriques
- [ ] Monter `/metrics` dans le router chi
- [ ] Créer `deploy/docker/docker-compose.observability.yml`
- [ ] Créer un dashboard Grafana JSON dans `deploy/grafana/dashboards/goloadit.json`

### Sprint D — Versionnement XML et rollback (1 jour)
- [ ] Compléter `internal/xml/store/store.go` : `ListVersions`, `LoadVersion`, `Rollback`
- [ ] Vérifier que `Save()` archive bien `history/v{n}.xml` à chaque appel
- [ ] Ajouter les 3 nouveaux endpoints dans `project_handler.go`
- [ ] Enregistrer les routes dans `router.go`
- [ ] Écrire les tests unitaires `internal/xml/store/store_test.go`

### Sprint E — Runbooks (0.5 jour)
- [ ] Rédiger `docs/runbooks/01-demarrage.md`
- [ ] Rédiger `docs/runbooks/02-debug-pipeline.md`
- [ ] Rédiger `docs/runbooks/03-rollback-projet.md`
- [ ] Rédiger `docs/runbooks/04-scaling-worker.md`

---

## Impacts sur les interfaces existantes

| Interface / fichier | Modification Phase 10 | Breaking ? |
|---|---|---|
| `contracts/block.go` | Ajout champ `Metrics *BlockMetrics` dans `BlockContext` | ❌ Non (champ optionnel) |
| `engine/executor.go` | Ajout `ExecuteWithOptions()` + instrumentation spans | ❌ Non (nouvelle surcharge) |
| `internal/xml/store/store.go` | Ajout `ListVersions`, `LoadVersion`, `Rollback` | ❌ Non (nouvelles méthodes) |
| `api/handlers/project_handler.go` | Ajout 3 handlers versionnement | ❌ Non (nouvelles routes) |
| `internal/orchestrator/service.go` | Instrumentation OTel + Prometheus | ❌ Non (comportement inchangé) |
| `internal/telemetry/` | Nouveau package `metrics.go` + `tracer.go` | ❌ Non (nouveau package) |
| `go.mod` | Ajout `testcontainers-go`, `otel`, `prometheus/client_golang` | ❌ Additive |

---

## Checklist finale Phase 10 — "Definition of Done"

### Tests e2e
- [ ] `go test ./tests/e2e/... -tags e2e -timeout 120s` passe sans erreur (Docker requis)
- [ ] 5 pipelines e2e validés : filter, join, split, aggregate, full
- [ ] Tests e2e intégrés au CI GitHub Actions (job séparé avec service Docker)

### Profiling
- [ ] `go test ./tests/benchmarks/ -bench=BenchmarkPipeline_1M -benchmem` s'exécute
- [ ] Aucune fuite mémoire détectée sur 1M lignes (stable en heap)
- [ ] `watchMemory` émet des logs d'alerte si le seuil est dépassé
- [ ] Profils `.pb.gz` générés dans `profiles/` documentés dans `docs/architecture/profiling.md`

### Observabilité
- [ ] `GET /metrics` retourne les métriques Prometheus au démarrage
- [ ] Chaque run crée une trace visible dans Jaeger UI (`http://localhost:16686`)
- [ ] Chaque bloc apparaît comme une span enfant avec `block.rows_in`, `block.rows_out`
- [ ] Dashboard Grafana opérationnel : runs totaux, durée, lignes par bloc, erreurs

### Versionnement XML
- [ ] `Save()` crée bien `history/v{n}.xml` à chaque appel (testé unitairement)
- [ ] `GET /api/v1/projects/{id}/versions` retourne l'historique correct
- [ ] `POST /api/v1/projects/{id}/rollback` restaure le projet à la version demandée
- [ ] Le SHA256 du `project.xml` courant correspond au hash affiché dans `ListVersions`

### Runbooks
- [ ] 4 runbooks rédigés dans `docs/runbooks/`
- [ ] Procédure de rollback testée manuellement et documentée

---

## Architecture rappel — Flux complet Phase 10
```
UI (ReactFlow) ──save──▶ POST /api/v1/projects/{id}
│ xml/serializer → project.xml
│ archive → history/v{n}.xml [Phase 10 ✅]
│ sha256 → project.xml.sha256
▼
POST /api/v1/runs ──────▶ orchestrator.RunProject()
│ span OTel : "orchestrator.RunProject" [Phase 10 ✅]
│ ActiveRuns.Inc() [Prometheus] [Phase 10 ✅]
│ xml/store.Load() → contracts.Project
▼
engine.ExecuteWithOptions()
│ profiler CPU/mem si EnableProfiling [Phase 10 ✅]
│ pour chaque bloc :
│ span OTel : "block.{type}" [Phase 10 ✅]
│ Run(bctx) → bctx.Metrics renseigné [Phase 10 ✅]
▼
ExecutionReport { Results[], Preview }
│ BlockRowsProcessed.Add() [Prometheus] [Phase 10 ✅]
│ RunsTotal.Inc() [Prometheus] [Phase 10 ✅]
▼
UI ◀── WebSocket/SSE ── suivi temps réel
Jaeger ◀── OTLP traces [Phase 10 ✅]
Prometheus ◀── /metrics [Phase 10 ✅]

GET /api/v1/projects/{id}/versions → ListVersions() [Phase 10 ✅]
POST /api/v1/projects/{id}/rollback → Rollback(version) [Phase 10 ✅]
```

---

*Document généré automatiquement par analyse du code source — à mettre à jour à chaque sprint.*