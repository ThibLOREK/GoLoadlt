# Étape 11 — Évolutions Avancées : État détaillé et plan d'implémentation

> Généré le 2026-04-24 · Basé sur un scan complet du code source et la feuille de route Phase 11

---

## Résumé de la Phase 11

La Phase 11 a pour objectif d'amener GoLoadIt au niveau d'une plateforme ETL professionnelle multi-utilisateurs,
en implémentant cinq axes avancés :
1. **Blocs DAG multi-branches** — Fork et Merge de flux de données
2. **CDC (Change Data Capture)** — Ingestion des changements en temps réel
3. **Templates de projets réutilisables** — Bibliothèque de pipelines paramétrables
4. **Multi-tenant** — Isolation des projets et connexions par organisation
5. **RBAC avancé** — Contrôle d'accès granulaire par projet et par connexion

**État global : Phases 0 → 10 supposées stables ✅ — Phase 11 à implémenter intégralement 🔴**

---

## Ce qui est déjà en place (Phases 0 → 10)

### ✅ Infrastructure & Foundation
- Structure repo complète : `cmd/`, `internal/`, `pkg/`, `api/`, `web/`, `deploy/`, `migrations/`
- Config multi-env YAML, Docker Compose, Makefile, Logger zerolog, Auth JWT, migrations SQL
- Observabilité complète : traces OTel par bloc, métriques Prometheus, dashboard Grafana (Phase 10)

### ✅ Moteur DAG & Contracts
- `contracts/block.go` : `DataType`, `ColumnDef`, `Schema`, `DataRow`, `Port`, `BlockContext`, `BlockMetrics`
- `engine/dag.go` : `BuildDAG()` avec tri topologique, gestion edges `disabled`
- `engine/executor.go` : `ExecuteWithOptions()`, spans OTel par bloc, `ExecutionReport`

### ✅ Catalogue complet de blocs (Phases 4, 5)
Tous les blocs sources, transforms et targets MVP sont enregistrés et testés unitairement.

### ✅ Connexions multi-env (Phase 6)
CRUD XML, résolution `ACTIVE_ENV`, secrets via env vars, ping réel, persistance `.env-state.json`.

### ✅ API, UI et Scheduling (Phases 7, 8, 9)
API OpenAPI complète, React Flow canvas, suivi WebSocket, cron par projet, worker asynchrone avec retry.

### ✅ Qualité et observabilité (Phase 10)
Tests e2e, profiling mémoire, traces distribuées, versionnement XML + rollback.

---

## Axe 1 — Blocs DAG Multi-branches : Fork et Merge

### Objectif

Permettre à un pipeline de **diviser un flux en plusieurs branches parallèles** (`fork`) et de
**fusionner plusieurs flux en un seul** (`merge`) en conservant l'ordre et la cohérence des données.
C'est le prérequis indispensable pour les pipelines complexes (enrichissement multi-source, dédoublonnage croisé, etc.).

### Distinction Fork vs Split vs Join vs Union

| Bloc | Comportement | Phase |
|---|---|---|
| `transform.split` | 1 flux → N flux **conditionnels** (chaque ligne part dans 1 seul port) | Phase 5 ✅ |
| `transform.join` | 2 flux → 1 flux (jointure par clé) | Phase 5 ✅ |
| `transform.fork` | 1 flux → N flux **identiques** (chaque ligne est dupliquée vers tous les ports) | **Phase 11** 🔴 |
| `transform.merge` | N flux → 1 flux (concaténation sans logique de clé) | **Phase 11** 🔴 |

### `internal/etl/blocks/transforms/fork.go` — NOUVEAU

```go
package transforms

import (
    "sync"

    "github.com/ThibLOREK/GoLoadIt/internal/etl/contracts"
    "github.com/ThibLOREK/GoLoadIt/internal/etl/blocks"
)

func init() {
    blocks.Register("transform.fork", func() contracts.Block { return &Fork{} })
}

// Fork duplique chaque DataRow vers tous ses ports de sortie en parallèle.
// Param : aucun
// Ports : 1 entrée "in", N sorties "out_1".."out_N" (N configuré à la construction du DAG)
type Fork struct{}

func (f *Fork) Run(bctx *contracts.BlockContext) error {
    in := bctx.Input("in")
    if in == nil {
        return contracts.ErrMissingPort("in")
    }

    var wg sync.WaitGroup
    for row := range in.Ch {
        row := row // capture
        bctx.Metrics.RowsIn++
        for _, out := range bctx.Outputs {
            wg.Add(1)
            go func(ch chan<- contracts.DataRow, r contracts.DataRow) {
                defer wg.Done()
                ch <- r
                bctx.Metrics.RowsOut++
            }(out.Ch, row)
        }
        wg.Wait()
    }

    // Fermer tous les ports de sortie
    for _, out := range bctx.Outputs {
        close(out.Ch)
    }
    return nil
}
```

### `internal/etl/blocks/transforms/merge.go` — NOUVEAU

```go
package transforms

import (
    "sync"

    "github.com/ThibLOREK/GoLoadIt/internal/etl/contracts"
    "github.com/ThibLOREK/GoLoadIt/internal/etl/blocks"
)

func init() {
    blocks.Register("transform.merge", func() contracts.Block { return &Merge{} })
}

// Merge fusionne N ports d'entrée en un seul port de sortie.
// L'ordre d'arrivée est non-déterministe (fusion concurrente).
// Param : aucun
// Ports : N entrées "in_1".."in_N", 1 sortie "out"
type Merge struct{}

func (m *Merge) Run(bctx *contracts.BlockContext) error {
    out := bctx.Output("out")
    if out == nil {
        return contracts.ErrMissingPort("out")
    }
    defer close(out.Ch)

    var wg sync.WaitGroup
    for _, in := range bctx.Inputs {
        in := in
        wg.Add(1)
        go func() {
            defer wg.Done()
            for row := range in.Ch {
                bctx.Metrics.RowsIn++
                out.Ch <- row
                bctx.Metrics.RowsOut++
            }
        }()
    }
    wg.Wait()
    return nil
}
```

### Impacts sur le moteur DAG — `engine/dag.go`

Le moteur doit détecter les blocs à N entrées ou N sorties et câbler les ports dynamiquement
en fonction du nombre d'edges entrants/sortants déclaré dans le XML du projet.

```go
// buildPorts construit les ports d'un nœud en fonction de ses edges dans le DAG.
// Pour transform.fork  : 1 port "in" fixe + N ports "out_1".."out_N"
// Pour transform.merge : N ports "in_1".."in_N" + 1 port "out" fixe
func buildPorts(node *contracts.Node, edges []contracts.Edge, bufferSize int) ([]*contracts.Port, []*contracts.Port) {
    var inputs, outputs []*contracts.Port

    inEdges  := filterEdgesByTarget(edges, node.ID)
    outEdges := filterEdgesBySource(edges, node.ID)

    switch node.Type {
    case "transform.fork":
        inputs = []*contracts.Port{{ID: "in", Ch: make(chan contracts.DataRow, bufferSize)}}
        for i, e := range outEdges {
            outputs = append(outputs, &contracts.Port{
                ID: fmt.Sprintf("out_%d", i+1),
                Ch: make(chan contracts.DataRow, bufferSize),
                // L'ID de l'edge cible permet au câblage de connecter le bon port
                EdgeID: e.ID,
            })
        }
    case "transform.merge":
        for i, e := range inEdges {
            inputs = append(inputs, &contracts.Port{
                ID:     fmt.Sprintf("in_%d", i+1),
                Ch:     make(chan contracts.DataRow, bufferSize),
                EdgeID: e.ID,
            })
        }
        outputs = []*contracts.Port{{ID: "out", Ch: make(chan contracts.DataRow, bufferSize)}}
    default:
        // Comportement standard : 1 entrée "in", 1 sortie "out"
        if len(inEdges) > 0 {
            inputs = []*contracts.Port{{ID: "in", Ch: make(chan contracts.DataRow, bufferSize)}}
        }
        if len(outEdges) > 0 {
            outputs = []*contracts.Port{{ID: "out", Ch: make(chan contracts.DataRow, bufferSize)}}
        }
    }
    return inputs, outputs
}
```

### Catalogue UI — `internal/etl/blocks/catalogue.go`

```go
meta("transform.fork",  "transform", "Fork",  "Duplique le flux vers N branches parallèles", 1, 1, 1, 4),
meta("transform.merge", "transform", "Merge", "Fusionne N flux en un seul",                  4, 1, 1, 1),
```

### XML de projet — exemple de pipeline Fork/Merge

```xml
<project id="proj-fork-demo" name="Fork + Merge Demo">
  <nodes>
    <node id="src"   type="source.csv"      label="Source" />
    <node id="fork1" type="transform.fork"  label="Fork" />
    <node id="fil1"  type="transform.filter" label="Filtre montant > 100">
      <params><param key="condition" value="amount > 100" /></params>
    </node>
    <node id="fil2"  type="transform.filter" label="Filtre montant ≤ 100">
      <params><param key="condition" value="amount <= 100" /></params>
    </node>
    <node id="merge" type="transform.merge" label="Merge" />
    <node id="tgt"   type="target.csv"      label="Cible" />
  </nodes>
  <edges>
    <edge id="e1" source="src"   target="fork1" sourcePort="out"   targetPort="in"   />
    <edge id="e2" source="fork1" target="fil1"  sourcePort="out_1" targetPort="in"   />
    <edge id="e3" source="fork1" target="fil2"  sourcePort="out_2" targetPort="in"   />
    <edge id="e4" source="fil1"  target="merge" sourcePort="out"   targetPort="in_1" />
    <edge id="e5" source="fil2"  target="merge" sourcePort="out"   targetPort="in_2" />
    <edge id="e6" source="merge" target="tgt"   sourcePort="out"   targetPort="in"   />
  </edges>
</project>
```

### `contracts/project.go` — ajout des champs `sourcePort` et `targetPort` sur `Edge`

```go
// Edge représente un lien entre deux nœuds du DAG.
// Phase 11 : ajout de SourcePort et TargetPort pour les blocs multi-ports.
type Edge struct {
    ID         string `xml:"id,attr"         json:"id"`
    Source     string `xml:"source,attr"     json:"source"`
    Target     string `xml:"target,attr"     json:"target"`
    SourcePort string `xml:"sourcePort,attr" json:"sourcePort"` // ← NOUVEAU Phase 11
    TargetPort string `xml:"targetPort,attr" json:"targetPort"` // ← NOUVEAU Phase 11
    Disabled   bool   `xml:"disabled,attr"   json:"disabled"`
}
```

### Tests unitaires — `tests/unit/transforms/fork_test.go`

```go
package transforms_test

import (
    "context"
    "sync"
    "testing"

    "github.com/ThibLOREK/GoLoadIt/internal/etl/contracts"
    _ "github.com/ThibLOREK/GoLoadIt/internal/etl/blocks/transforms"
    "github.com/stretchr/testify/assert"
)

func TestFork_DuplicatesAllRows(t *testing.T) {
    in  := make(chan contracts.DataRow, 3)
    out1 := make(chan contracts.DataRow, 3)
    out2 := make(chan contracts.DataRow, 3)

    in <- contracts.DataRow{"id": 1}
    in <- contracts.DataRow{"id": 2}
    in <- contracts.DataRow{"id": 3}
    close(in)

    bctx := &contracts.BlockContext{
        Ctx:     context.Background(),
        Params:  map[string]string{},
        Inputs:  []*contracts.Port{{ID: "in",    Ch: in}},
        Outputs: []*contracts.Port{
            {ID: "out_1", Ch: out1},
            {ID: "out_2", Ch: out2},
        },
        Metrics: &contracts.BlockMetrics{},
    }

    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        // Fork ferme out1 et out2 à la fin — on draine
    }()

    err := (&Fork{}).Run(bctx)
    assert.NoError(t, err)
    assert.Len(t, drain(out1), 3)
    assert.Len(t, drain(out2), 3)
    assert.EqualValues(t, 3, bctx.Metrics.RowsIn)
}

func TestMerge_CombinesAllInputs(t *testing.T) {
    in1 := make(chan contracts.DataRow, 3)
    in2 := make(chan contracts.DataRow, 3)
    out := make(chan contracts.DataRow, 6)

    in1 <- contracts.DataRow{"src": "A", "val": 1}
    in1 <- contracts.DataRow{"src": "A", "val": 2}
    close(in1)
    in2 <- contracts.DataRow{"src": "B", "val": 10}
    in2 <- contracts.DataRow{"src": "B", "val": 20}
    in2 <- contracts.DataRow{"src": "B", "val": 30}
    close(in2)

    bctx := &contracts.BlockContext{
        Ctx:     context.Background(),
        Params:  map[string]string{},
        Inputs:  []*contracts.Port{{ID: "in_1", Ch: in1}, {ID: "in_2", Ch: in2}},
        Outputs: []*contracts.Port{{ID: "out", Ch: out}},
        Metrics: &contracts.BlockMetrics{},
    }

    err := (&Merge{}).Run(bctx)
    assert.NoError(t, err)
    rows := drain(out)
    assert.Len(t, rows, 5)
    assert.EqualValues(t, 5, bctx.Metrics.RowsIn)
}
```

---

## Axe 2 — CDC (Change Data Capture)

### Objectif

Permettre à GoLoadIt d'**ingérer les changements de données en temps réel** depuis les bases sources
(PostgreSQL via `pgoutput` / `wal2json`, MySQL via binlog) sans scanner la table entière à chaque exécution.
Le CDC est la brique fondamentale pour les pipelines near-realtime (latence < 1s).

### Architecture CDC
```
Base source PostgreSQL
│ WAL (Write-Ahead Log) — slot de réplication logique
│
▼
source.postgres_cdc (nouveau bloc)
│ connexion pgx replication stream
│ décode pgoutput → DataRow{_op: "INSERT"|"UPDATE"|"DELETE", ...}
│
▼
Moteur DAG GoLoadIt
│ blocs de transformation standards (filter, cast, add_column...)
│
▼
target.postgres / target.rest / target.kafka (futur)
```

### `internal/etl/blocks/sources/postgres_cdc.go` — NOUVEAU

```go
package sources

import (
    "context"
    "fmt"
    "strings"

    "github.com/jackc/pglogrepl"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/jackc/pgx/v5/pgproto3"
    "github.com/ThibLOREK/GoLoadIt/internal/etl/blocks"
    "github.com/ThibLOREK/GoLoadIt/internal/etl/contracts"
)

func init() {
    blocks.Register("source.postgres_cdc", func() contracts.Block { return &PostgresCDC{} })
}

// PostgresCDC lit le WAL PostgreSQL via un slot de réplication logique (pgoutput).
// Params :
//   connectionRef  : ID de la connexion PostgreSQL (doit avoir REPLICATION privilege)
//   publication    : nom de la publication PostgreSQL (ex: "goloadit_pub")
//   slotName       : nom du slot de réplication (ex: "goloadit_slot")
//   tables         : liste de tables à surveiller (ex: "public.orders,public.customers")
//   batchSize      : nombre de messages à accumuler avant de flush (défaut: 100)
type PostgresCDC struct{}

func (p *PostgresCDC) Run(bctx *contracts.BlockContext) error {
    connStr := bctx.Connection.DSN
    publication := bctx.Params["publication"]
    slotName    := bctx.Params["slotName"]
    if publication == "" || slotName == "" {
        return fmt.Errorf("postgres_cdc: paramètres 'publication' et 'slotName' obligatoires")
    }

    // Connexion en mode réplication
    conn, err := pgconn.Connect(bctx.Ctx, connStr+" replication=database")
    if err != nil {
        return fmt.Errorf("postgres_cdc: connexion réplication: %w", err)
    }
    defer conn.Close(bctx.Ctx)

    // Créer le slot si inexistant
    _ = pglogrepl.CreateReplicationSlot(bctx.Ctx, conn, slotName, "pgoutput",
        pglogrepl.CreateReplicationSlotOptions{Temporary: false})

    // Démarrer le stream
    opts := pglogrepl.StartReplicationOptions{
        PluginArgs: []string{
            "proto_version '1'",
            fmt.Sprintf("publication_names '%s'", publication),
        },
    }
    if err := pglogrepl.StartReplication(bctx.Ctx, conn, slotName, pglogrepl.LSN(0), opts); err != nil {
        return fmt.Errorf("postgres_cdc: start replication: %w", err)
    }

    out := bctx.Output("out")
    if out == nil {
        return contracts.ErrMissingPort("out")
    }
    defer close(out.Ch)

    relations := map[uint32]*pglogrepl.RelationMessage{}

    for {
        select {
        case <-bctx.Ctx.Done():
            return nil
        default:
        }

        msg, err := conn.ReceiveMessage(bctx.Ctx)
        if err != nil {
            if strings.Contains(err.Error(), "context canceled") {
                return nil
            }
            return fmt.Errorf("postgres_cdc: receive: %w", err)
        }

        // Ignorer les messages non-data
        cd, ok := msg.(*pgproto3.CopyData)
        if !ok {
            continue
        }
        if cd.Data != pglogrepl.XLogDataByteID {
            continue
        }

        xld, _ := pglogrepl.ParseXLogData(cd.Data[1:])
        logicalMsg, err := pglogrepl.Parse(xld.WALData)
        if err != nil {
            continue
        }

        switch v := logicalMsg.(type) {
        case *pglogrepl.RelationMessage:
            relations[v.RelationID] = v

        case *pglogrepl.InsertMessage:
            row := decodeRow(v.Tuple, relations[v.RelationID])
            row["_op"] = "INSERT"
            out.Ch <- row
            bctx.Metrics.RowsOut++

        case *pglogrepl.UpdateMessage:
            row := decodeRow(v.NewTuple, relations[v.RelationID])
            row["_op"] = "UPDATE"
            out.Ch <- row
            bctx.Metrics.RowsOut++

        case *pglogrepl.DeleteMessage:
            row := decodeRow(v.OldTuple, relations[v.RelationID])
            row["_op"] = "DELETE"
            out.Ch <- row
            bctx.Metrics.RowsOut++
        }

        // Envoyer le feedback de progression au serveur (toutes les 100 lignes)
        if bctx.Metrics.RowsOut%100 == 0 {
            _ = pglogrepl.SendStandbyStatusUpdate(bctx.Ctx, conn,
                pglogrepl.StandbyStatusUpdate{WALWritePosition: xld.WALStart})
        }
    }
}

// decodeRow convertit un TupleData pglogrepl en DataRow GoLoadIt.
func decodeRow(tuple *pglogrepl.TupleData, rel *pglogrepl.RelationMessage) contracts.DataRow {
    if tuple == nil || rel == nil {
        return contracts.DataRow{}
    }
    row := make(contracts.DataRow, len(rel.Columns))
    for i, col := range rel.Columns {
        if i < len(tuple.Columns) {
            row[col.Name] = string(tuple.Columns[i].Data)
        }
    }
    return row
}
```

### Catalogue UI

```go
meta("source.postgres_cdc", "source", "PostgreSQL CDC",
    "Ingère les changements WAL en temps réel (INSERT/UPDATE/DELETE)",
    0, 0, 1, 1),
```

### Paramètres UI — `NodeConfigPanel.tsx` additions

| Param Go | Champ UI | Type |
|---|---|---|
| `connectionRef` | Sélecteur connexion PostgreSQL | `<select>` |
| `publication` | Nom de publication | `text` |
| `slotName` | Nom du slot de réplication | `text` |
| `tables` | Tables surveillées (séparées par virgule) | `text` |
| `batchSize` | Taille du batch avant flush | `number` (défaut: 100) |

### Prérequis PostgreSQL (documentation opérationnelle)

```sql
-- Sur la base source : activer le WAL logique
ALTER SYSTEM SET wal_level = logical;
SELECT pg_reload_conf();

-- Créer la publication pour les tables souhaitées
CREATE PUBLICATION goloadit_pub FOR TABLE public.orders, public.customers;

-- Accorder les droits de réplication à l'utilisateur GoLoadIt
ALTER ROLE goloadit_user REPLICATION;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO goloadit_user;
```

---

## Axe 3 — Templates de Projets Réutilisables

### Objectif

Permettre de **sauvegarder un pipeline existant comme template**, de **lister les templates disponibles**
dans une bibliothèque, et d'**instancier un nouveau projet depuis un template** avec substitution de paramètres.
Un template est un fichier XML de projet avec des variables `{{PARAM_NAME}}` dans ses champs.

### Modèle de données — `contracts/project.go`

```go
// ProjectTemplate représente un pipeline réutilisable paramétrable.
type ProjectTemplate struct {
    ID          string            `xml:"id,attr"          json:"id"`
    Name        string            `xml:"name,attr"        json:"name"`
    Description string            `xml:"description,attr" json:"description"`
    Category    string            `xml:"category,attr"    json:"category"` // ex: "etl", "reporting", "cdc"
    Params      []TemplateParam   `xml:"params>param"     json:"params"`
    Nodes       []Node            `xml:"nodes>node"       json:"nodes"`
    Edges       []Edge            `xml:"edges>edge"       json:"edges"`
}

// TemplateParam décrit un paramètre substituable dans le template.
type TemplateParam struct {
    Key         string `xml:"key,attr"         json:"key"`
    Label       string `xml:"label,attr"       json:"label"`
    Description string `xml:"description,attr" json:"description"`
    Default     string `xml:"default,attr"     json:"default"`
    Required    bool   `xml:"required,attr"    json:"required"`
}
```

### Structure de stockage
```
templates/
├── etl/
│ ├── csv-to-postgres.xml # source.csv → filter → cast → target.postgres
│ ├── postgres-to-csv.xml # source.postgres → select → target.csv
│ └── dedup-pipeline.xml # source → dedup → sort → target
├── reporting/
│ ├── daily-aggregate.xml # source → filter (today) → aggregate → target
│ └── pivot-report.xml # source → pivot → target.csv
└── cdc/
└── postgres-cdc-sync.xml # source.postgres_cdc → filter → target.postgres
```


### Exemple de template XML paramétrable

```xml
<?xml version="1.0" encoding="UTF-8"?>
<template id="csv-to-postgres" name="CSV vers PostgreSQL" category="etl"
          description="Charge un fichier CSV dans une table PostgreSQL avec filtre optionnel">
  <params>
    <param key="SOURCE_PATH"     label="Chemin du fichier CSV" required="true"  default="" />
    <param key="SOURCE_DELIM"    label="Délimiteur"            required="false" default="," />
    <param key="FILTER_COND"     label="Condition de filtre"   required="false" default="1=1" />
    <param key="TARGET_CONN"     label="Connexion PostgreSQL"  required="true"  default="" />
    <param key="TARGET_TABLE"    label="Table cible"           required="true"  default="" />
    <param key="TARGET_MODE"     label="Mode d'écriture"       required="false" default="insert" />
  </params>
  <nodes>
    <node id="src"    type="source.csv"       label="Lecture CSV">
      <params>
        <param key="path"      value="{{SOURCE_PATH}}" />
        <param key="delimiter" value="{{SOURCE_DELIM}}" />
        <param key="hasHeader" value="true" />
      </params>
    </node>
    <node id="filter" type="transform.filter" label="Filtre">
      <params><param key="condition" value="{{FILTER_COND}}" /></params>
    </node>
    <node id="tgt"    type="target.postgres"  label="Écriture PostgreSQL">
      <params>
        <param key="connectionRef" value="{{TARGET_CONN}}" />
        <param key="table"         value="{{TARGET_TABLE}}" />
        <param key="mode"          value="{{TARGET_MODE}}" />
      </params>
    </node>
  </nodes>
  <edges>
    <edge id="e1" source="src"    target="filter" sourcePort="out" targetPort="in" />
    <edge id="e2" source="filter" target="tgt"    sourcePort="out" targetPort="in" />
  </edges>
</template>
```

### `internal/etl/template/service.go` — NOUVEAU

```go
package template

import (
    "encoding/xml"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/ThibLOREK/GoLoadIt/internal/etl/contracts"
    "github.com/google/uuid"
)

type Service struct {
    TemplatesDir string
    ProjectsDir  string
}

func NewService(templatesDir, projectsDir string) *Service {
    return &Service{TemplatesDir: templatesDir, ProjectsDir: projectsDir}
}

// List retourne tous les templates disponibles (toutes catégories).
func (s *Service) List() ([]contracts.ProjectTemplate, error) {
    var templates []contracts.ProjectTemplate
    err := filepath.Walk(s.TemplatesDir, func(path string, info os.FileInfo, err error) error {
        if err != nil || info.IsDir() || !strings.HasSuffix(path, ".xml") {
            return err
        }
        t, err := s.load(path)
        if err != nil {
            return nil // ignorer les fichiers invalides
        }
        templates = append(templates, *t)
        return nil
    })
    return templates, err
}

// Instantiate crée un nouveau projet à partir d'un template avec substitution des paramètres.
// values : map[string]string de la forme {"SOURCE_PATH": "/data/sales.csv", ...}
func (s *Service) Instantiate(templateID string, projectName string, values map[string]string) (*contracts.Project, error) {
    // Trouver le template
    var tmpl *contracts.ProjectTemplate
    _ = filepath.Walk(s.TemplatesDir, func(path string, info os.FileInfo, err error) error {
        if err != nil || info.IsDir() {
            return err
        }
        t, err := s.load(path)
        if err == nil && t.ID == templateID {
            tmpl = t
        }
        return nil
    })
    if tmpl == nil {
        return nil, fmt.Errorf("template '%s' introuvable", templateID)
    }

    // Vérifier les paramètres obligatoires
    for _, p := range tmpl.Params {
        if p.Required {
            if v, ok := values[p.Key]; !ok || v == "" {
                return nil, fmt.Errorf("paramètre obligatoire manquant: %s (%s)", p.Key, p.Label)
            }
        }
    }

    // Construire les valeurs finales (defaults + overrides)
    resolved := make(map[string]string)
    for _, p := range tmpl.Params {
        resolved[p.Key] = p.Default
    }
    for k, v := range values {
        resolved[k] = v
    }

    // Sérialiser le template en XML, substituer les variables {{KEY}} → valeur
    raw, err := xml.MarshalIndent(tmpl, "", "  ")
    if err != nil {
        return nil, err
    }
    xml_str := string(raw)
    for k, v := range resolved {
        xml_str = strings.ReplaceAll(xml_str, "{{"+k+"}}", v)
    }

    // Parser le XML résultant en Project
    var project contracts.Project
    if err := xml.Unmarshal([]byte(xml_str), &project); err != nil {
        return nil, fmt.Errorf("instantiation XML: %w", err)
    }

    // Assigner un nouvel ID et nom
    project.ID   = uuid.New().String()
    project.Name = projectName

    return &project, nil
}

// SaveAsTemplate sauvegarde un projet existant comme template.
func (s *Service) SaveAsTemplate(project *contracts.Project, templateID, description, category string) error {
    tmpl := contracts.ProjectTemplate{
        ID:          templateID,
        Name:        project.Name,
        Description: description,
        Category:    category,
        Nodes:       project.Nodes,
        Edges:       project.Edges,
    }
    data, err := xml.MarshalIndent(tmpl, "", "  ")
    if err != nil {
        return err
    }
    dir := filepath.Join(s.TemplatesDir, category)
    _ = os.MkdirAll(dir, 0o755)
    path := filepath.Join(dir, templateID+".xml")
    return os.WriteFile(path, append([]byte(xml.Header), data...), 0o644)
}

func (s *Service) load(path string) (*contracts.ProjectTemplate, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var t contracts.ProjectTemplate
    return &t, xml.Unmarshal(data, &t)
}
```

### Nouveaux endpoints API — `api/handlers/template_handler.go` — NOUVEAU

```go
package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/ThibLOREK/GoLoadIt/internal/etl/template"
)

type TemplateHandler struct {
    svc     *template.Service
    xmlStore interface{ Save(p interface{}) error }
}

// GET /api/v1/templates
// → retourne la liste de tous les templates avec leurs paramètres
func (h *TemplateHandler) List(w http.ResponseWriter, r *http.Request) {
    templates, err := h.svc.List()
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, templates)
}

// POST /api/v1/templates/{templateID}/instantiate
// Body: { "projectName": "Mon Pipeline", "params": {"SOURCE_PATH": "/data/sales.csv", ...} }
// → crée un nouveau projet et le persiste dans projects/
func (h *TemplateHandler) Instantiate(w http.ResponseWriter, r *http.Request) {
    templateID := chi.URLParam(r, "templateID")
    var body struct {
        ProjectName string            `json:"projectName"`
        Params      map[string]string `json:"params"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        writeError(w, http.StatusBadRequest, "corps JSON invalide")
        return
    }
    project, err := h.svc.Instantiate(templateID, body.ProjectName, body.Params)
    if err != nil {
        writeError(w, http.StatusUnprocessableEntity, err.Error())
        return
    }
    // Persister le nouveau projet
    if err := h.xmlStore.Save(project); err != nil {
        writeError(w, http.StatusInternalServerError, "sauvegarde projet: "+err.Error())
        return
    }
    w.WriteHeader(http.StatusCreated)
    writeJSON(w, project)
}

// POST /api/v1/projects/{id}/save-as-template
// Body: { "templateID": "my-template", "description": "...", "category": "etl" }
func (h *TemplateHandler) SaveAsTemplate(w http.ResponseWriter, r *http.Request) {
    projectID := chi.URLParam(r, "id")
    var body struct {
        TemplateID  string `json:"templateID"`
        Description string `json:"description"`
        Category    string `json:"category"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        writeError(w, http.StatusBadRequest, "corps JSON invalide")
        return
    }
    // Charger le projet source
    project, err := h.xmlStore.Load(projectID)
    if err != nil {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }
    if err := h.svc.SaveAsTemplate(project, body.TemplateID, body.Description, body.Category); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    w.WriteHeader(http.StatusCreated)
    writeJSON(w, map[string]string{"status": "saved", "templateID": body.TemplateID})
}
```

### Enregistrement des routes

```go
// api/handlers/router.go
r.Route("/api/v1/templates", func(r chi.Router) {
    r.Get("/", templateHandler.List)
    r.Post("/{templateID}/instantiate", templateHandler.Instantiate)
})

r.Post("/api/v1/projects/{id}/save-as-template", templateHandler.SaveAsTemplate)
```

---

## Axe 4 — Multi-Tenant

### Objectif

Isoler complètement les ressources (projets, connexions, runs, templates) par **organisation**,
afin de supporter plusieurs clients ou équipes sur la même instance GoLoadIt.

### Modèle de données — Ajout de `OrgID` sur les entités clés

```go
// pkg/models/tenant.go — NOUVEAU
package models

// Organization représente un tenant (organisation cliente).
type Organization struct {
    ID        string    `db:"id"         json:"id"`
    Name      string    `db:"name"       json:"name"`
    Slug      string    `db:"slug"       json:"slug"`        // ex: "acme-corp"
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    Active    bool      `db:"active"     json:"active"`
}
```

### Migration SQL — `migrations/005_multitenancy.sql` — NOUVEAU

```sql
-- 005_multitenancy.sql
CREATE TABLE organizations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    slug       VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    active     BOOLEAN NOT NULL DEFAULT TRUE
);

-- Ajouter org_id sur toutes les tables existantes
ALTER TABLE projects ADD COLUMN org_id UUID REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE runs     ADD COLUMN org_id UUID REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE users    ADD COLUMN org_id UUID REFERENCES organizations(id) ON DELETE CASCADE;

-- Index pour toutes les requêtes filtrées par org
CREATE INDEX idx_projects_org ON projects(org_id);
CREATE INDEX idx_runs_org     ON runs(org_id);
CREATE INDEX idx_users_org    ON users(org_id);

-- Organisation par défaut pour les données existantes
INSERT INTO organizations (id, name, slug) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Default', 'default');
UPDATE projects SET org_id = '00000000-0000-0000-0000-000000000001' WHERE org_id IS NULL;
UPDATE runs     SET org_id = '00000000-0000-0000-0000-000000000001' WHERE org_id IS NULL;
UPDATE users    SET org_id = '00000000-0000-0000-0000-000000000001' WHERE org_id IS NULL;

ALTER TABLE projects ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE runs     ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE users    ALTER COLUMN org_id SET NOT NULL;
```

### Middleware d'isolation tenant — `api/middleware/tenant.go` — NOUVEAU

```go
package middleware

import (
    "context"
    "net/http"
    "strings"

    "github.com/ThibLOREK/GoLoadIt/internal/security"
)

type contextKey string

const TenantKey contextKey = "orgID"

// TenantMiddleware extrait l'orgID du token JWT et l'injecte dans le contexte.
// Toutes les requêtes API sont ensuite filtrées par orgID.
func TenantMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        claims, ok := r.Context().Value(security.ClaimsKey).(*security.JWTClaims)
        if !ok || claims.OrgID == "" {
            http.Error(w, "tenant non identifié", http.StatusUnauthorized)
            return
        }
        ctx := context.WithValue(r.Context(), TenantKey, claims.OrgID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// OrgIDFromContext extrait l'orgID du contexte HTTP.
func OrgIDFromContext(ctx context.Context) string {
    orgID, _ := ctx.Value(TenantKey).(string)
    return orgID
}
```

### Isolation des fichiers XML par tenant
```
projects/
└── {org_slug}/
└── {project_id}/
├── project.xml
└── history/

connections/
└── {org_slug}/
├── conn-crm.xml
└── conn-erp.xml
```

### Modification de `XMLStore` pour le multi-tenant

```go
// internal/xml/store/store.go — modification de New() pour supporter orgSlug
type XMLStore struct {
    BaseDir string
    OrgSlug string // ← NOUVEAU : chaque tenant a son sous-dossier
}

func NewForOrg(baseDir, orgSlug string) *XMLStore {
    return &XMLStore{
        BaseDir: filepath.Join(baseDir, orgSlug),
        OrgSlug: orgSlug,
    }
}

// Le reste des méthodes (Save, Load, List, etc.) est inchangé —
// elles opèrent toutes sur s.BaseDir qui est déjà scopé par org.
```

### Adaptation des handlers

```go
// Exemple dans project_handler.go — toutes les requêtes filtrées par orgID
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
    orgID := middleware.OrgIDFromContext(r.Context())
    // Instancier le store scopé par org
    store := xmlstore.NewForOrg(h.baseDir, h.orgSlugFor(orgID))
    projects, err := store.List()
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, projects)
}
```

---

## Axe 5 — RBAC Avancé

### Objectif

Définir des **rôles par organisation** et des **permissions granulaires par ressource** (projet, connexion),
permettant à un admin de contrôler précisément qui peut lire, exécuter ou éditer chaque pipeline.

### Modèle de permissions

```go
// internal/security/rbac.go — NOUVEAU
package security

// Role définit un rôle au sein d'une organisation.
type Role string

const (
    RoleAdmin      Role = "admin"       // tout faire dans l'org
    RoleEditor     Role = "editor"      // créer/modifier projets et connexions
    RoleRunner     Role = "runner"      // exécuter des projets, lire les runs
    RoleViewer     Role = "viewer"      // lecture seule projets et runs
    RoleConnAdmin  Role = "conn_admin"  // gérer les connexions uniquement
)

// Permission est une action granulaire sur une ressource.
type Permission string

const (
    PermProjectRead     Permission = "project:read"
    PermProjectWrite    Permission = "project:write"
    PermProjectDelete   Permission = "project:delete"
    PermProjectRun      Permission = "project:run"
    PermProjectRollback Permission = "project:rollback"
    PermConnRead        Permission = "connection:read"
    PermConnWrite       Permission = "connection:write"
    PermConnDelete      Permission = "connection:delete"
    PermConnTest        Permission = "connection:test"
    PermEnvSwitch       Permission = "environment:switch"
    PermOrgManage       Permission = "org:manage"       // gestion membres/rôles
)

// rolePermissions mappe chaque rôle vers ses permissions.
var rolePermissions = map[Role][]Permission{
    RoleAdmin: {
        PermProjectRead, PermProjectWrite, PermProjectDelete, PermProjectRun, PermProjectRollback,
        PermConnRead, PermConnWrite, PermConnDelete, PermConnTest,
        PermEnvSwitch, PermOrgManage,
    },
    RoleEditor: {
        PermProjectRead, PermProjectWrite, PermProjectRun,
        PermConnRead, PermConnWrite, PermConnTest,
    },
    RoleRunner: {
        PermProjectRead, PermProjectRun,
        PermConnRead, PermConnTest,
    },
    RoleViewer: {
        PermProjectRead,
        PermConnRead,
    },
    RoleConnAdmin: {
        PermConnRead, PermConnWrite, PermConnDelete, PermConnTest,
        PermEnvSwitch,
    },
}

// HasPermission vérifie si un rôle possède une permission.
func HasPermission(role Role, perm Permission) bool {
    for _, p := range rolePermissions[role] {
        if p == perm {
            return true
        }
    }
    return false
}
```

### Migration SQL — `migrations/006_rbac.sql` — NOUVEAU

```sql
-- 006_rbac.sql
CREATE TYPE user_role AS ENUM ('admin', 'editor', 'runner', 'viewer', 'conn_admin');

ALTER TABLE users ADD COLUMN role user_role NOT NULL DEFAULT 'viewer';

-- Permissions granulaires par ressource (optionnel — niveau avancé)
CREATE TABLE resource_permissions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    resource    VARCHAR(50)  NOT NULL,   -- 'project' | 'connection'
    resource_id UUID,                    -- NULL = toutes les ressources du type
    permission  VARCHAR(50)  NOT NULL,
    granted_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, org_id, resource, resource_id, permission)
);

CREATE INDEX idx_resource_perms_user ON resource_permissions(user_id, org_id);
```

### Middleware RBAC — `api/middleware/rbac.go` — NOUVEAU

```go
package middleware

import (
    "net/http"

    "github.com/ThibLOREK/GoLoadIt/internal/security"
)

// RequirePermission retourne un middleware qui vérifie qu'un utilisateur a la permission requise.
func RequirePermission(perm security.Permission) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims, ok := r.Context().Value(security.ClaimsKey).(*security.JWTClaims)
            if !ok {
                http.Error(w, "non authentifié", http.StatusUnauthorized)
                return
            }
            role := security.Role(claims.Role)
            if !security.HasPermission(role, perm) {
                http.Error(w, "permission refusée : "+string(perm), http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### Application sur les routes — `api/handlers/router.go`

```go
import mw "github.com/ThibLOREK/GoLoadIt/api/middleware"
import sec "github.com/ThibLOREK/GoLoadIt/internal/security"

r.Route("/api/v1/projects", func(r chi.Router) {
    r.Use(mw.TenantMiddleware)
    r.With(mw.RequirePermission(sec.PermProjectRead)).Get("/", projectHandler.List)
    r.With(mw.RequirePermission(sec.PermProjectWrite)).Post("/", projectHandler.Create)
    r.Route("/{id}", func(r chi.Router) {
        r.With(mw.RequirePermission(sec.PermProjectRead)).Get("/", projectHandler.Get)
        r.With(mw.RequirePermission(sec.PermProjectWrite)).Put("/", projectHandler.Save)
        r.With(mw.RequirePermission(sec.PermProjectDelete)).Delete("/", projectHandler.Delete)
        r.With(mw.RequirePermission(sec.PermProjectRun)).Post("/run", runsHandler.Run)
        r.With(mw.RequirePermission(sec.PermProjectRollback)).Post("/rollback", projectHandler.Rollback)
        r.With(mw.RequirePermission(sec.PermProjectRead)).Get("/versions", projectHandler.ListVersions)
    })
})

r.Route("/api/v1/connections", func(r chi.Router) {
    r.Use(mw.TenantMiddleware)
    r.With(mw.RequirePermission(sec.PermConnRead)).Get("/", connHandler.List)
    r.With(mw.RequirePermission(sec.PermConnWrite)).Post("/", connHandler.Create)
    r.Route("/{connID}", func(r chi.Router) {
        r.With(mw.RequirePermission(sec.PermConnRead)).Get("/", connHandler.Get)
        r.With(mw.RequirePermission(sec.PermConnWrite)).Put("/", connHandler.Update)
        r.With(mw.RequirePermission(sec.PermConnDelete)).Delete("/", connHandler.Delete)
        r.With(mw.RequirePermission(sec.PermConnTest)).Post("/test", connHandler.Test)
    })
})

r.With(mw.RequirePermission(sec.PermEnvSwitch)).Put("/api/v1/environment", connHandler.SwitchEnv)
```

### `internal/security/claims.go` — Ajout de `OrgID` et `Role` dans le JWT

```go
// JWTClaims — modification pour inclure OrgID et Role (Phase 11)
type JWTClaims struct {
    UserID string `json:"sub"`
    Email  string `json:"email"`
    OrgID  string `json:"org_id"` // ← NOUVEAU Phase 11
    Role   string `json:"role"`   // ← NOUVEAU Phase 11 (ex: "admin", "editor")
    jwt.RegisteredClaims
}
```

---

## Plan d'action Phase 11

### Sprint A — Fork & Merge (1.5 jours)
- [ ] Créer `internal/etl/blocks/transforms/fork.go` avec `init()` + `Run()`
- [ ] Créer `internal/etl/blocks/transforms/merge.go` avec `init()` + `Run()`
- [ ] Modifier `contracts/project.go` : ajouter `SourcePort`, `TargetPort` sur `Edge`
- [ ] Modifier `engine/dag.go` : `buildPorts()` avec gestion multi-ports Fork/Merge
- [ ] Ajouter `transform.fork` et `transform.merge` dans `catalogue.go`
- [ ] Écrire `tests/unit/transforms/fork_test.go` et `merge_test.go`
- [ ] Vérifier le pipeline e2e `source → fork → (2× filter) → merge → target`

### Sprint B — CDC PostgreSQL (2 jours)
- [ ] Créer `internal/etl/blocks/sources/postgres_cdc.go`
- [ ] Ajouter `source.postgres_cdc` dans `catalogue.go`
- [ ] Ajouter les champs UI dans `NodeConfigPanel.tsx` (connectionRef, publication, slotName, tables)
- [ ] Documenter les prérequis PostgreSQL dans `docs/runbooks/05-cdc-postgres.md`
- [ ] Ajouter `testcontainers-go` avec PostgreSQL en mode WAL pour le test d'intégration CDC
- [ ] Test d'intégration : INSERT en base → le bloc CDC émet la DataRow en < 500ms

### Sprint C — Templates (1.5 jours)
- [ ] Créer `contracts/project.go` : struct `ProjectTemplate`, `TemplateParam`
- [ ] Créer `internal/etl/template/service.go` : `List`, `Instantiate`, `SaveAsTemplate`
- [ ] Créer `templates/etl/`, `templates/reporting/`, `templates/cdc/` avec fichiers XML d'exemple
- [ ] Créer `api/handlers/template_handler.go` : `List`, `Instantiate`, `SaveAsTemplate`
- [ ] Enregistrer les routes templates dans `router.go`
- [ ] UI : ajouter une page "Templates" avec grille de cartes et formulaire d'instanciation

### Sprint D — Multi-tenant (2 jours)
- [ ] Créer `migrations/005_multitenancy.sql` + appliquer
- [ ] Créer `pkg/models/tenant.go` : struct `Organization`
- [ ] Créer `api/middleware/tenant.go` : `TenantMiddleware`, `OrgIDFromContext`
- [ ] Modifier `internal/xml/store/store.go` : `NewForOrg(baseDir, orgSlug)`
- [ ] Adapter tous les handlers pour instancier le store scopé par org
- [ ] Restructurer le dossier `projects/` et `connections/` en `projects/{org_slug}/`
- [ ] Ajouter des endpoints admin : CRUD organisations, ajout/retrait membres

### Sprint E — RBAC (1.5 jours)
- [ ] Créer `internal/security/rbac.go` : types `Role`, `Permission`, `rolePermissions`, `HasPermission`
- [ ] Créer `migrations/006_rbac.sql` + appliquer
- [ ] Créer `api/middleware/rbac.go` : `RequirePermission(perm)`
- [ ] Modifier `internal/security/claims.go` : ajouter `OrgID`, `Role` dans `JWTClaims`
- [ ] Appliquer `RequirePermission` sur toutes les routes (voir code ci-dessus)
- [ ] UI : masquer les actions non autorisées selon le rôle JWT de l'utilisateur connecté
- [ ] Écrire `tests/unit/security/rbac_test.go` : matrice rôle × permission

---

## Impacts sur les interfaces existantes

| Interface / fichier | Modification Phase 11 | Breaking ? |
|---|---|---|
| `contracts/project.go` | Ajout `SourcePort`, `TargetPort` sur `Edge` + structs `ProjectTemplate`, `TemplateParam` | ⚠️ XML Breaking (champs optionnels — rétrocompat si `omitempty`) |
| `engine/dag.go` | `buildPorts()` généralise le câblage multi-ports | ⚠️ Interne uniquement — API publique inchangée |
| `internal/security/claims.go` | Ajout `OrgID`, `Role` dans `JWTClaims` | ⚠️ Tokens existants invalides → re-login requis |
| `internal/xml/store/store.go` | Ajout `NewForOrg()` + champ `OrgSlug` | ❌ Non breaking (nouvelle surcharge) |
| `api/handlers/router.go` | Ajout middlewares `TenantMiddleware` + `RequirePermission` sur toutes les routes | ⚠️ Breaking sur routes non authentifiées (dev local) |
| `migrations/` | Nouvelles migrations `005` et `006` | ⚠️ Requiert `make migrate` |
| `go.mod` | Ajout `github.com/jackc/pglogrepl` pour CDC | ❌ Additive |

---

## Checklist finale Phase 11 — "Definition of Done"

### Blocs Fork & Merge
- [ ] `go test ./tests/unit/transforms/... -run TestFork` et `TestMerge` verts
- [ ] Test e2e `source → fork → (2× filter) → merge → target.csv` produit le résultat attendu
- [ ] `NodeConfigPanel.tsx` affiche N ports configurables pour Fork/Merge
- [ ] Les edges `sourcePort` / `targetPort` sont sauvegardés dans le XML du projet

### CDC PostgreSQL
- [ ] `source.postgres_cdc` est visible dans la palette de blocs
- [ ] Test d'intégration : 1 INSERT → 1 DataRow reçue en < 500ms
- [ ] `_op: "INSERT"|"UPDATE"|"DELETE"` est bien présent sur chaque DataRow
- [ ] La déconnexion / reconnexion du slot de réplication est gérée proprement
- [ ] Runbook `docs/runbooks/05-cdc-postgres.md` décrit les prérequis PostgreSQL

### Templates
- [ ] `GET /api/v1/templates` retourne ≥ 3 templates d'exemple
- [ ] `POST /api/v1/templates/{id}/instantiate` crée un projet avec substitution correcte
- [ ] Les paramètres obligatoires non fournis retournent HTTP 422 avec message explicite
- [ ] `POST /api/v1/projects/{id}/save-as-template` crée bien un fichier XML dans `templates/`

### Multi-tenant
- [ ] Migration `005` appliquée sans erreur sur une base de données de test
- [ ] Un utilisateur de l'org A ne peut pas accéder aux projets de l'org B (HTTP 404 ou 403)
- [ ] `GET /api/v1/projects` d'un utilisateur `org-acme` ne retourne que les projets d'`org-acme`
- [ ] Les fichiers XML sont bien stockés sous `projects/acme-corp/` et non à la racine

### RBAC
- [ ] `go test ./tests/unit/security/rbac_test.go` — matrice complète rôle × permission verte
- [ ] Un `viewer` qui tente `POST /api/v1/runs` reçoit HTTP 403
- [ ] Un `runner` peut lancer un run mais pas modifier un projet (HTTP 403 sur PUT)
- [ ] L'UI masque les boutons "Éditer", "Supprimer", "Basculer env" pour les viewers
- [ ] Migration `006` appliquée sans erreur

---

## Architecture rappel — Flux Phase 11 complet
```
[Multi-tenant]
JWT Token { sub, email, org_id: "org-acme", role: "editor" }
│
▼
TenantMiddleware → OrgIDFromContext = "org-acme"
RequirePermission(PermProjectWrite) → HasPermission(RoleEditor, "project:write") = ✅
│
▼
XMLStore.NewForOrg(baseDir, "acme-corp")
→ projects/acme-corp/{project_id}/project.xml [isolation tenant]

[Templates]
POST /templates/csv-to-postgres/instantiate
→ template.Instantiate("csv-to-postgres", "Mon Pipeline", {"SOURCE_PATH": "/data/sales.csv"})
→ substitution {{SOURCE_PATH}} → "/data/sales.csv"
→ XMLStore.Save(newProject)
→ projects/acme-corp/{new_uuid}/project.xml

[Fork & Merge dans le DAG]
engine.BuildDAG(project)
→ node "fork1" type="transform.fork"
→ buildPorts: 1× "in" + N× "out_1".."out_N" (selon edges sortants)
→ node "merge" type="transform.merge"
→ buildPorts: N× "in_1".."in_N" + 1× "out"
→ exécution parallèle via goroutines + channels

[CDC]
source.postgres_cdc
→ connexion pgconn replication=database
→ pglogrepl.StartReplication(slotName, publication)
→ streaming WAL → DataRow{_op: "INSERT", id: 42, amount: 150.0}
→ out.Ch ← row
→ blocs transforms standards (filter, cast, etc.)
→ target.postgres.upsert()

---

*Document généré automatiquement par analyse du code source — à mettre à jour à chaque sprint.*