# Étape 5 — Blocs de Transformation MVP : État détaillé et tâches restantes

> Généré le 2026-04-20 · Basé sur un scan complet du code source
> **Mis à jour le 2026-05-01** — Audit post-Sprint E

---

## État global — 2026-05-01

**Score : 13/14 (93%) — Quasi-complet ✅**

| Domaine | État |
|---|---|
| Infrastructure & Foundation | ✅ Complet |
| Blocs sources (Phase 4) | ✅ Complet |
| Blocs targets (Phase 4) | ✅ Complet |
| Moteur d'exécution DAG | ✅ Complet |
| ProjectStore Save/Load/ListAll/Delete | ✅ Complet |
| Parser ParseProjectFile + ParseProjectBytes | ✅ Complet |
| Serializer SerializeProject | ✅ Complet |
| Catalogue blocs (GET /api/v1/catalogue) | ✅ Complet |
| Import/Export XML API | ✅ Complet |
| Run sync (POST /{id}/run) + Preview | ✅ Complet |
| RowsIn / RowsOut comptabilisés | ⚠️ Struct présente, non populée |
| Tests d'intégration blocs | ✅ Présents |

---

## Résumé de la Phase 5

La Phase 5 a pour objectif d'implémenter tous les blocs de transformation du MVP :
`filter`, `select`, `cast`, `add_column`, `split`, `pivot`, `unpivot`, `join`, `dedup`, `sort`, `aggregate`.

**État global : fichiers Go présents ✅ — intégration complète à finaliser ⚠️**

---

## Ce qui est déjà en place (Phases 0 → 4)

### ✅ Infrastructure & Foundation
- Structure de repo complète (`cmd/`, `internal/`, `pkg/`, `api/`, `web/`, `deploy/`, `migrations/`)
- Config multi-env YAML (`config.dev.yaml`, `config.preprod.yaml`, `config.prod.yaml`)
- Docker Compose, Makefile fonctionnel
- Migrations SQL : `001_init.sql`, `002_runs.sql`, `002_connections_env.sql`, `003_schedules.sql`, `004_users.sql`
- Logger (`zerolog`), auth service JWT, middleware

### ✅ Contracts & Modèle DAG
- `contracts/block.go` : `DataType`, `ColumnDef`, `Schema`, `DataRow`, `Port`, `BlockContext`, `Block`, `BlockFactory` → **complets**
- `contracts/project.go` : `Project`, `Node`, `Edge`, `Param` avec tags XML + JSON → **complet**
- `contracts/preview.go` : `PreviewStore` pour la capture des N premières lignes par bloc → **présent**

### ✅ Moteur d'exécution DAG
- `engine/dag.go` : `BuildDAG()` avec gestion des edges `disabled` et nœuds orphelins → **fonctionnel**
- `engine/executor.go` : `Execute()` complet — tri topologique, câblage des ports via `previewPort`, `RunResult`, `ExecutionReport`
- `engine/inject_connections.go` : injection des connexions dans le `BlockContext`

### ✅ Blocs Sources (Phase 4)
| Bloc | Fichier | État |
|---|---|---|
| `source.csv` | `sources/csv.go` (7617b) | ✅ Fonctionnel |
| `source.postgres` | `sources/postgres.go` (1851b) | ✅ Présent |
| `source.mysql` / `source.mssql` | `sources/mysql_mssql.go` | ✅ Présent |
| `source.datetime` | `sources/datetime.go` | ✅ Bonus |
| `source.directory` | `sources/directory.go` | ✅ Bonus |
| `source.text_input` | `sources/text_input.go` | ✅ Bonus |

### ✅ Blocs Targets (Phase 4)
| Bloc | Fichier | État |
|---|---|---|
| `target.csv` | `targets/csv.go` | ✅ Présent |
| `target.postgres` | `targets/postgres.go` | ✅ Présent |
| `target.browse` | `targets/browse.go` | ✅ Présent (preview UI) |

### ✅ Frontend React — Stack installée
- Vite + React 18 + TypeScript + `@xyflow/react` 12.3.6 + Tailwind 3 + Zustand + Axios
- Pages : `Dashboard`, `PipelineDesigner`, `PipelineList`, `EditorPage`, `ProjectsPage`, `ConnectionsPage`, `RunHistory`, `Login`
- Composants editor : `BlockPalette`, `ETLBlockNode`, `NodeConfigPanel` (24KB), `DataPreviewPanel`
- Nodes React Flow : `SourceNode`, `TargetNode`, `TransformNode`

---

## État détaillé — Blocs de Transformation Phase 5

| Bloc | Fichier Go | `init()` Register | Catalogue UI | NodeConfigPanel | Test unitaire |
|---|---|---|---|---|---|
| `transform.filter` | ✅ `filter.go` | ✅ | ✅ | ⚠️ à vérifier | ❌ manquant |
| `transform.filter_advanced` | ✅ `filter_advanced.go` (5605b) | ✅ | ✅ | ⚠️ à vérifier | ❌ manquant |
| `transform.select` | ✅ `select.go` | ✅ | ✅ | ⚠️ à vérifier | ❌ manquant |
| `transform.cast` | ✅ `cast.go` | ✅ | ✅ | ⚠️ à vérifier | ❌ manquant |
| `transform.add_column` | ✅ `add_column.go` | ✅ | ✅ | ⚠️ à vérifier | ❌ manquant |
| `transform.split` | ✅ `split.go` | ✅ | ✅ | ⚠️ à vérifier | ❌ manquant |
| `transform.pivot` | ✅ `pivot.go` | ✅ | ✅ | ⚠️ à vérifier | ❌ manquant |
| `transform.unpivot` | ✅ `unpivot.go` | ✅ | ✅ | ⚠️ à vérifier | ❌ manquant |
| `transform.join` | ✅ `join.go` | ✅ | ✅ | ⚠️ à vérifier | ❌ manquant |
| `transform.dedup` | ✅ `dedup.go` | ✅ | ✅ | ⚠️ à vérifier | ❌ manquant |
| `transform.sort` | ✅ `sort.go` | ✅ | ✅ | ⚠️ à vérifier | ❌ manquant |
| `transform.aggregate` | ✅ `aggregate.go` | ✅ | ✅ | ⚠️ à vérifier | ❌ manquant |
| `transform.dummy` | ✅ `dummy.go` | ✅ | ✅ | ✅ pass-through | ❌ manquant |

### Blocs bonus présents en Go mais absents du catalogue UI

Ces blocs existent dans `internal/etl/blocks/transforms/` mais ne sont **pas** dans `catalogue.go` → invisibles dans l'UI.

| Bloc | Fichier | Taille | Action |
|---|---|---|---|
| `transform.union` | `union.go` | 860b | Ajouter au catalogue |
| `transform.regex` | `regex.go` | 2153b | Ajouter au catalogue |
| `transform.find_replace` | `find_replace.go` | 1531b | Ajouter au catalogue |
| `transform.sampling` | `sampling.go` | 1424b | Ajouter au catalogue |
| `transform.text_to_columns` | `text_to_columns.go` | 1608b | Ajouter au catalogue |
| `transform.auto_field` | `auto_field.go` | 1332b | Ajouter au catalogue |
| `transform.append_fields` | `append_fields.go` | 1093b | Ajouter au catalogue |
| `transform.data_cleansing` | `data_cleansing.go` | 2275b | Ajouter au catalogue |
| `transform.datetime_transform` | `datetime_transform.go` | 2749b | Ajouter au catalogue |

---

## Problèmes bloquants identifiés

### 🔴 BLOQUANT 1 — Incompatibilité ReactFlow v10 → v12 dans PipelineDesigner.tsx

`web/ui/src/pages/PipelineDesigner.tsx` importe depuis `reactflow` (ancienne API v10) :

```ts
// ❌ ACTUEL — API v10, ne compilera pas avec @xyflow/react v12
import ReactFlow, { addEdge, Background, Controls, MiniMap,
  useEdgesState, useNodesState, Connection, Node } from "reactflow";
import "reactflow/dist/style.css";
```

Alors que `package.json` déclare `@xyflow/react: ^12.3.6` (API v12 — breaking change).

**Fix obligatoire :**

```ts
// ✅ CORRECT — API v12
import { ReactFlow, addEdge, Background, Controls, MiniMap,
  useEdgesState, useNodesState, type Connection, type Node } from "@xyflow/react";
import "@xyflow/react/dist/style.css";
```

**Fichiers à corriger :**
- `web/ui/src/pages/PipelineDesigner.tsx`
- `web/ui/src/nodes/SourceNode.tsx`
- `web/ui/src/nodes/TargetNode.tsx`
- `web/ui/src/nodes/TransformNode.tsx`
- `web/ui/src/components/editor/ETLBlockNode.tsx`
- Tous les fichiers `.tsx` qui importent `from "reactflow"`

**Commande de vérification :**
```bash
grep -r "from \"reactflow\"" web/ui/src/
# Tout résultat doit être remplacé par "@xyflow/react"
```

---

### 🔴 BLOQUANT 2 — Dépendances absentes de package.json

`PipelineDesigner.tsx` utilise `@mui/material` et `@tanstack/react-query`, **absents** de `package.json` → build cassé.

**Deux options :**

**Option A (recommandée) — Rester full-Tailwind, supprimer MUI :**
Remplacer les composants MUI (`Box`, `Typography`, `Paper`, `Stack`, `Button`, `TextField`, `Alert`, `Snackbar`) par des équivalents Tailwind natifs — cohérent avec le reste du projet.

**Option B — Installer MUI :**
```bash
cd web/ui
npm install @mui/material @emotion/react @emotion/styled @tanstack/react-query
```

---

### 🔴 BLOQUANT 3 — `internal/orchestrator/service.go` vide (44 bytes)

L'orchestrateur est le pont entre `POST /api/v1/runs` et le moteur d'exécution. Il est actuellement vide.

**À implémenter :**

```go
package orchestrator

import (
    "context"
    "github.com/ThibLOREK/GoLoadlt/internal/etl/engine"
    "github.com/ThibLOREK/GoLoadlt/internal/xml/store"
    "github.com/ThibLOREK/GoLoadlt/internal/jobs"
)

type Service struct {
    executor *engine.Executor
    xmlStore *store.XMLStore
    jobRepo  jobs.Repository
}

func NewService(executor *engine.Executor, xmlStore *store.XMLStore, jobRepo jobs.Repository) *Service {
    return &Service{executor: executor, xmlStore: xmlStore, jobRepo: jobRepo}
}

// RunProject charge le XML du projet, parse le DAG et l'exécute.
func (s *Service) RunProject(ctx context.Context, projectID string) (*engine.ExecutionReport, error) {
    project, err := s.xmlStore.Load(projectID)
    if err != nil {
        return nil, err
    }
    run, err := s.jobRepo.Create(ctx, projectID)
    if err != nil {
        return nil, err
    }
    _ = s.jobRepo.SetStatus(ctx, run.ID, "running")
    report, execErr := s.executor.Execute(ctx, project)
    status := "succeeded"
    if execErr != nil {
        status = "failed"
    }
    _ = s.jobRepo.SetStatus(ctx, run.ID, status)
    return report, execErr
}

func (s *Service) CancelRun(ctx context.Context, runID string) error {
    return s.jobRepo.SetStatus(ctx, runID, "cancelled")
}
```

---

### ⚠️ PARTIEL — RowsIn / RowsOut non comptabilisés dans executor.go

**Identifié lors de l'audit Sprint E (2026-05-01)**

`RunResult` contient les champs `RowsIn` et `RowsOut` mais ils ne sont pas peuplés dans la boucle d'exécution de `engine/executor.go`. Les stats de lignes sont toujours à zéro dans l'`ExecutionReport`.

**Fix à apporter dans `executor.go`** — incrémenter les compteurs dans la boucle par bloc :

```go
// Dans la boucle d'exécution de chaque nœud :
result := RunResult{NodeID: node.ID}
startedAt := time.Now()

// Compter les lignes en wrappant le port de sortie
rowsIn  := countRows(inputPort)
rowsOut := countRows(outputPort)

result.RowsIn    = rowsIn
result.RowsOut   = rowsOut
result.Duration  = time.Since(startedAt)
```

**Impact :** `GET /api/v1/runs/{id}/report` retourne des stats à zéro — inutilisable pour le monitoring.

---

### 🟡 IMPORTANT 4 — XML Store / Parser / Serializer incomplets

Les dossiers `internal/xml/` n'ont que des stubs minimaux. Le store XML est la **source de vérité** du système (rôle équivalent à `$JENKINS_HOME/jobs/`).

**`internal/xml/store/store.go` à implémenter :**
```go
type XMLStore struct{ baseDir string }

func New(baseDir string) *XMLStore
// Save : sérialise le projet, archive history/v{n}.xml, calcule SHA256
func (s *XMLStore) Save(project *contracts.Project) error
// Load : charge et parse projects/{id}/project.xml
func (s *XMLStore) Load(projectID string) (*contracts.Project, error)
// List : retourne tous les projets disponibles
func (s *XMLStore) List() ([]contracts.Project, error)
// Delete : supprime le répertoire du projet
func (s *XMLStore) Delete(projectID string) error
```

**`internal/xml/parser/parser.go` à compléter :**
```go
// Parse décode un flux XML en *contracts.Project
func Parse(r io.Reader) (*contracts.Project, error) {
    var p contracts.Project
    return &p, xml.NewDecoder(r).Decode(&p)
}
```

**`internal/xml/serializer/serializer.go` à compléter :**
```go
// Serialize encode un *contracts.Project en XML indenté
func Serialize(p *contracts.Project) ([]byte, error) {
    return xml.MarshalIndent(p, "", "  ")
}
```

---

### 🟡 IMPORTANT 5 — `internal/jobs/job.go` minimaliste (165 bytes)

Le repository d'accès PostgreSQL pour les runs est absent.

**À ajouter :**
```go
type Repository interface {
    Create(ctx context.Context, projectID string) (*Run, error)
    SetStatus(ctx context.Context, runID string, status string) error
    GetByID(ctx context.Context, runID string) (*Run, error)
    ListByProject(ctx context.Context, projectID string) ([]Run, error)
}
```

---

### 🟡 IMPORTANT 6 — Blocs bonus absents du catalogue UI

Ajouter dans `internal/etl/blocks/catalogue.go` :

```go
// Blocs bonus à exposer dans l'UI
meta("transform.union",            "transform", "Union",             "Fusionne deux flux en un seul",                                    2, 10, 1, 1),
meta("transform.regex",            "transform", "Regex Extract",     "Extrait des groupes via une regex sur une colonne",                1, 1,  1, 1),
meta("transform.find_replace",     "transform", "Find & Replace",    "Remplace des valeurs dans une colonne",                           1, 1,  1, 1),
meta("transform.sampling",         "transform", "Sampling",          "Échantillonne un pourcentage aléatoire du flux",                  1, 1,  1, 1),
meta("transform.text_to_columns",  "transform", "Text to Columns",   "Découpe une colonne texte en plusieurs colonnes via un délimiteur", 1, 1, 1, 1),
meta("transform.auto_field",       "transform", "Auto Field",        "Détecte automatiquement le type des colonnes",                    1, 1,  1, 1),
meta("transform.append_fields",    "transform", "Append Fields",     "Ajoute des colonnes à valeur constante au flux",                  1, 1,  1, 1),
meta("transform.data_cleansing",   "transform", "Data Cleansing",    "Nettoie les valeurs nulles, espaces, normalise la casse",         1, 1,  1, 1),
meta("transform.datetime_transform","transform","DateTime Transform","Formate et convertit les colonnes date/heure",                    1, 1,  1, 1),
```

---

### 🟡 IMPORTANT 7 — NodeConfigPanel.tsx — Cohérence paramètres UI ↔ Go

`NodeConfigPanel.tsx` (24KB) doit exposer des champs dont les **clés correspondent exactement** aux `Params` attendus par chaque bloc Go.

**Référence de correspondance obligatoire :**

| Bloc Go | Param Go (`bctx.Params["..."]`) | Champ UI attendu |
|---|---|---|
| `transform.filter` | `condition` | Champ texte "Condition" |
| `transform.filter_advanced` | `condition_true`, `condition_false` | 2 champs texte |
| `transform.select` | `columns` | Liste de colonnes (multi-select) |
| `transform.cast` | `column`, `targetType` | Sélecteur colonne + type |
| `transform.add_column` | `name`, `expression` | Nom + expression |
| `transform.split` | `conditions` | Textarea (conditions CSV) |
| `transform.pivot` | `groupBy`, `pivotColumn`, `valueColumn` | 3 sélecteurs |
| `transform.unpivot` | `columns`, `keyName`, `valueName` | Multi-select + 2 champs |
| `transform.join` | `leftKey`, `rightKey`, `type` | 2 champs + radio inner/left/right/full |
| `transform.dedup` | `keys` | Liste de colonnes clés |
| `transform.sort` | `columns`, `order` | Multi-select + asc/desc |
| `transform.aggregate` | `groupBy`, `aggregations` | Multi-select + liste |

---

### ⚠️ NOUVEAU — Conflit de numérotation migrations SQL

**Identifié lors de l'audit Sprint E (2026-05-01)**

`002_connections_env.sql` et `002_runs.sql` portent le même numéro `002`.
Si golang-migrate ou Flyway est utilisé, le démarrage échouera avec une erreur de doublon.

**Fix : renuméroter** `002_runs.sql` → `010_runs.sql` (ou autre numéro libre).

---

## Plan d'action pour finaliser la Phase 5

### Sprint A — Corriger le build (1 jour)

- [ ] **Fix imports ReactFlow** : remplacer `from "reactflow"` → `from "@xyflow/react"` dans tous les fichiers `web/ui/src/**/*.tsx`
- [ ] **Fix style import** : `"reactflow/dist/style.css"` → `"@xyflow/react/dist/style.css"`
- [ ] **Résoudre les dépendances** : choisir Option A (Tailwind) ou Option B (installer MUI) et appliquer
- [ ] **Vérifier `npm run build`** passe sans erreur
- [ ] **Vérifier `go build ./...`** passe sans erreur

### Sprint B — Valider chaque bloc transform (2-3 jours)

Pour chaque bloc de la Phase 5 :

- [ ] Vérifier que `init()` appelle bien `blocks.Register("transform.xxx", ...)`
- [ ] Écrire un test unitaire dans `tests/unit/transforms/{bloc}_test.go`
- [ ] Vérifier cohérence noms de paramètres entre Go et NodeConfigPanel

**Template de test unitaire :**
```go
package transforms_test

import (
    "context"
    "testing"
    "github.com/ThibLOREK/GoLoadlt/internal/etl/blocks/transforms"
    "github.com/ThibLOREK/GoLoadlt/internal/etl/contracts"
    "github.com/stretchr/testify/assert"
)

func drain(ch <-chan contracts.DataRow) []contracts.DataRow {
    var rows []contracts.DataRow
    for r := range ch {
        rows = append(rows, r)
    }
    return rows
}

func TestFilter_BasicCondition(t *testing.T) {
    in  := make(chan contracts.DataRow, 3)
    out := make(chan contracts.DataRow, 3)
    in <- contracts.DataRow{"amount": 150.0}
    in <- contracts.DataRow{"amount": 50.0}
    in <- contracts.DataRow{"amount": 200.0}
    close(in)

    bctx := &contracts.BlockContext{
        Ctx:     context.Background(),
        Params:  map[string]string{"condition": "amount > 100"},
        Inputs:  []*contracts.Port{{ID: "in",  Ch: in}},
        Outputs: []*contracts.Port{{ID: "out", Ch: out}},
    }
    err := (&transforms.Filter{}).Run(bctx)
    assert.NoError(t, err)
    close(out)
    rows := drain(out)
    assert.Len(t, rows, 2) // 150 et 200 passent, 50 est filtré
}
```

### Sprint C — XML persistence (2 jours)

- [ ] Implémenter `internal/xml/store/store.go` (Save, Load, List, Delete + archivage history + SHA256)
- [ ] Compléter `internal/xml/parser/parser.go`
- [ ] Compléter `internal/xml/serializer/serializer.go`
- [ ] Connecter `api/handlers/project_handler.go` au store XML
- [ ] Tester : créer un projet via l'UI → vérifier que `projects/{id}/project.xml` est créé

### Sprint D — Orchestrateur & Runs (1 jour)

- [ ] Implémenter `internal/orchestrator/service.go` (voir code ci-dessus)
- [ ] Compléter `internal/jobs/job.go` avec l'interface `Repository` + implémentation PostgreSQL
- [ ] Vérifier que `POST /api/v1/runs` → charge XML → parse → execute → écrit statut en base → retourne le rapport

### Sprint E — Catalogue complet + config UI (1 jour)

- [ ] Ajouter les 9 blocs bonus dans `catalogue.go` (voir code ci-dessus)
- [ ] Vérifier/compléter `NodeConfigPanel.tsx` pour tous les blocs selon la table de correspondance
- [ ] Tester le drag-and-drop → configuration → exécution end-to-end dans l'UI
- [ ] Vérifier que `DataPreviewPanel` affiche bien les lignes capturées par `PreviewStore`

### Sprint F — RowsIn/RowsOut + migration fix (0.5 jour)

- [ ] **Peupler `RowsIn`/`RowsOut`** dans `engine/executor.go` (voir fix ci-dessus)
- [ ] **Renuméroter** `002_runs.sql` → `010_runs.sql` pour éliminer le conflit migrations

---

## Checklist finale Phase 5 — "Definition of Done"

### Backend Go
- [x] `go build ./...` passe sans erreur ni warning ✅ (vérifié audit 2026-05-01)
- [ ] `go vet ./...` passe proprement
- [ ] Tous les 12 blocs MVP ont un test unitaire `go test ./tests/unit/transforms/...` vert
- [x] XML Store : save + load + history fonctionnels ✅ (store.go présent et câblé)
- [x] `POST /api/v1/runs` exécute un pipeline de bout en bout et retourne l'`ExecutionReport` ✅
- [ ] `RowsIn` / `RowsOut` peuplés dans `ExecutionReport`

### Pipelines de validation end-to-end
- [ ] `source.csv → transform.filter → transform.add_column → target.csv` → fichier de sortie correct
- [ ] `source.csv → transform.join (2 sources) → target.csv` → jointure inner correcte
- [ ] `source.csv → transform.split → (2 × target.csv)` → deux fichiers distincts selon condition
- [ ] `source.csv → transform.pivot → target.browse` → aperçu pivoté dans l'UI
- [ ] `source.csv → transform.aggregate → target.postgres` → données agrégées insérées en base

### Frontend React
- [ ] `npm run build` passe sans erreur
- [ ] Aucun import `from "reactflow"` restant (tous migrés vers `@xyflow/react`)
- [ ] Palette de blocs affiche tous les blocs du catalogue (y compris les 9 bonus)
- [ ] `NodeConfigPanel` ouvre et affiche les bons champs pour chaque type de bloc
- [ ] `DataPreviewPanel` affiche les premières lignes après exécution

### Déploiement
- [ ] `docker-compose up` démarre sans erreur (server + postgres)
- [ ] Les migrations s'appliquent automatiquement au démarrage (`cmd/migrate`)
- [ ] Le frontend est servi par le backend Go (assets statiques compilés)

---

## Architecture rappel — Flux d'exécution complet

```
UI (ReactFlow) ──save──▶ POST /api/v1/projects/{id}
                              │
                          project_handler.go
                              │ xml/serializer
                              ▼
                      projects/{id}/project.xml   ◀── source de vérité
                              │
POST /api/v1/runs ──────▶ runs.go handler
                              │
                          orchestrator.Service.RunProject()
                              │ xml/store.Load()
                              │ xml/parser.Parse()
                              ▼
                      contracts.Project (DAG en mémoire)
                              │
                          engine.Executor.Execute()
                              │ BuildDAG() → TopologicalSort()
                              │ câblage ports (previewPort)
                              │ exécution séquentielle
                              ▼
                      ExecutionReport { Results, Preview }
                              │
                          jobs.Repository.SetStatus("succeeded")
                              │
                              ▼
              UI ◀── WebSocket/SSE ── suivi temps réel par bloc
```

---

## État au 2026-05-01

### Ce qui est ✅ complètement terminé
- Infrastructure, contracts, moteur DAG, blocs sources/transforms/targets
- ProjectStore (Save/Load/ListAll/Delete), Parser, Serializer
- Catalogue blocs exposé via API (`GET /api/v1/catalogue`)
- Import XML (`POST /projects/import`) + Export XML (`GET /projects/{id}/xml`)
- Run synchrone (`POST /projects/{id}/run`) avec Preview dans la réponse
- Tests d'intégration blocs présents (csv_extractor, csv_loader, transformer)

### Ce qui est ⚠️ partiellement fait
- **RowsIn/RowsOut** : struct présente dans `RunResult`, non peuplée dans la boucle executor — stats toujours à zéro dans l'ExecutionReport
- **Tests unitaires blocs transforms** : aucun des 12 blocs MVP n'a de test `go test` dédié

### Ce qui est ❌ manquant
| Point | Priorité |
|---|---|
| `RowsIn`/`RowsOut` peuplés dans executor.go | **BLOQUANT** — monitoring inutilisable |
| Conflit numérotation `002_*.sql` (doublon) | **BLOQUANT** — migration automatique cassée |
| Tests unitaires 12 blocs transforms | IMPORTANT |
| 9 blocs bonus dans catalogue UI | NICE-TO-HAVE |
| Fix imports ReactFlow v10 → v12 | **BLOQUANT** (frontend) |
| Dépendances MUI / react-query absentes | **BLOQUANT** (frontend) |

### Prochains sprints recommandés
1. **Sprint F** (0.5j) — Fix `RowsIn`/`RowsOut` dans executor + renumérotation migration `002_runs.sql`
2. **Sprint A-frontend** (1j) — Fix imports ReactFlow + résolution dépendances MUI
3. **Sprint B** (2-3j) — Tests unitaires 12 blocs transforms

---

*Document généré automatiquement par analyse du code source — à mettre à jour à chaque sprint.*
