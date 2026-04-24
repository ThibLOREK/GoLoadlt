# Étape 8 — Interface Visuelle MVP : État détaillé et tâches restantes

> Généré le 2026-04-24 · Basé sur un scan complet du code source

---

## Résumé de la Phase 8

La Phase 8 a pour objectif de livrer une interface visuelle fonctionnelle et connectée au backend Go :
authentification, canvas React Flow avec palette de blocs, configuration au clic, gestion des connexions
multi-env, exécution temps réel avec suivi WebSocket bloc par bloc, et historique des runs.

**État global : pages et composants présents ✅ — mais intégration complète à finaliser ⚠️**

---

## Ce qui est déjà en place (Phases 0 → 7)

### ✅ Infrastructure & Foundation
- Structure de repo complète (`cmd/`, `internal/`, `pkg/`, `api/`, `web/`, `deploy/`, `migrations/`)
- Config multi-env YAML (`config.dev.yaml`, `config.preprod.yaml`, `config.prod.yaml`)
- Docker Compose, Makefile fonctionnel
- Migrations SQL : `001_init.sql`, `002_runs.sql`, `002_connections_env.sql`, `003_schedules.sql`, `004_users.sql`
- Logger (`zerolog`), auth service JWT, middleware

### ✅ Backend Go — Moteur & API
- `contracts/block.go` : `DataType`, `ColumnDef`, `Schema`, `DataRow`, `Port`, `BlockContext`, `Block`, `BlockFactory` → **complets**
- `contracts/project.go` : `Project`, `Node`, `Edge`, `Param` avec tags XML + JSON → **complet**
- `engine/dag.go` + `engine/executor.go` : moteur DAG complet (tri topologique, câblage ports, `ExecutionReport`)
- `orchestrator/service.go` : `RunProject()`, `CancelRun()` → **fonctionnel**
- `xml/store`, `xml/parser`, `xml/serializer` → persistance XML opérationnelle
- Blocs sources : `source.csv`, `source.postgres`, `source.mssql`, `source.mysql`, bonus datetime/directory/text_input
- Blocs transforms : `filter`, `select`, `cast`, `add_column`, `split`, `pivot`, `unpivot`, `join`, `dedup`, `sort`, `aggregate` + 9 blocs bonus
- Blocs targets : `target.csv`, `target.postgres`, `target.browse`
- API complète : CRUD projets, connexions, runs, switch d'environnement, documentation OpenAPI

### ✅ Frontend React — Stack installée
- Vite + React 18 + TypeScript + `@xyflow/react` 12.3.6 + Tailwind 3 + Zustand + Axios
- Pages existantes : `Login.tsx`, `Dashboard.tsx`, `ProjectsPage.tsx`, `PipelineDesigner.tsx`,
  `PipelineList.tsx`, `EditorPage.tsx`, `ConnectionsPage.tsx`, `RunHistory.tsx`
- Composants editor présents : `BlockPalette`, `ETLBlockNode`, `NodeConfigPanel` (24KB), `DataPreviewPanel`
- Nodes React Flow : `SourceNode`, `TargetNode`, `TransformNode`
- Store Zustand : `web/ui/src/store/`
- Hooks : `web/ui/src/hooks/`
- Client API : `web/ui/src/api/` + `api.js`

---

## État détaillé — Composants UI Phase 8

### Pages

| Page | Fichier | Existe | Auth Guard | Connectée API | État |
|---|---|---|---|---|---|
| Login | `Login.tsx` | ✅ | — | ⚠️ à vérifier | ⚠️ |
| Dashboard | `Dashboard.tsx` | ✅ | ⚠️ | ⚠️ | ⚠️ |
| Liste des projets | `ProjectsPage.tsx` | ✅ | ⚠️ | ⚠️ | ⚠️ |
| Designer canvas | `EditorPage.tsx` | ✅ | ⚠️ | ⚠️ | ⚠️ |
| Connexions | `ConnectionsPage.tsx` | ✅ | ⚠️ | ⚠️ | ⚠️ |
| Historique runs | `RunHistory.tsx` | ✅ | ⚠️ | ⚠️ | ⚠️ |
| Pipeline Designer | `PipelineDesigner.tsx` | ✅ | ⚠️ | ⚠️ | ⚠️ |

### Composants Editor

| Composant | Fichier | Existe | Complet | Blocs catalogue | WebSocket |
|---|---|---|---|---|---|
| Palette de blocs | `BlockPalette` | ✅ | ⚠️ | ⚠️ catalogue partiel | ❌ N/A |
| Nœud ETL | `ETLBlockNode` | ✅ | ⚠️ | ⚠️ | ❌ N/A |
| Panneau config | `NodeConfigPanel` | ✅ (24KB) | ⚠️ | ⚠️ champs à vérifier | ❌ N/A |
| Aperçu données | `DataPreviewPanel` | ✅ | ⚠️ | ❌ N/A | ❌ non branché |
| Suivi exécution | `RunProgressPanel` | ❌ **manquant** | ❌ | ❌ N/A | ❌ à créer |

---

## Problèmes bloquants identifiés

### 🔴 BLOQUANT 1 — Authentification non gardée (routes publiques)

`App.tsx` déclare les routes mais il n'existe pas de `PrivateRoute` / `AuthGuard` vérifiant la présence
du token JWT avant d'accéder au canvas ou aux connexions.

**Fix obligatoire :**

```tsx
// web/ui/src/components/auth/AuthGuard.tsx
import { Navigate } from "react-router-dom";
import { useAuthStore } from "@/store/authStore";

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((s) => s.token);
  if (!token) return <Navigate to="/login" replace />;
  return <>{children}</>;
}
```

**Intégration dans `App.tsx` :**

```tsx
<Route path="/editor/:id" element={
  <AuthGuard><EditorPage /></AuthGuard>
} />
<Route path="/connections" element={
  <AuthGuard><ConnectionsPage /></AuthGuard>
} />
```

**Fichiers à créer / modifier :**
- `web/ui/src/components/auth/AuthGuard.tsx` → **à créer**
- `web/ui/src/store/authStore.ts` → vérifier qu'il persiste le token JWT
- `web/ui/src/App.tsx` → wrapper toutes les routes privées

---

### 🔴 BLOQUANT 2 — `PipelineDesigner.tsx` : import ReactFlow v10 non migré

`PipelineDesigner.tsx` (5658b) importe encore depuis `reactflow` (API v10) alors que
`package.json` déclare `@xyflow/react: ^12.3.6` (breaking change v12).

**Fix obligatoire :**

```ts
// ❌ ACTUEL
import ReactFlow, { addEdge, Background, Controls, MiniMap,
  useEdgesState, useNodesState, Connection, Node } from "reactflow";
import "reactflow/dist/style.css";

// ✅ CORRECT v12
import { ReactFlow, addEdge, Background, Controls, MiniMap,
  useEdgesState, useNodesState, type Connection, type Node } from "@xyflow/react";
import "@xyflow/react/dist/style.css";
```

**Vérification complète :**
```bash
grep -r "from \"reactflow\"" web/ui/src/
# Tout résultat doit être remplacé par "@xyflow/react"
```

**Fichiers concernés :**
- `web/ui/src/pages/PipelineDesigner.tsx`
- `web/ui/src/nodes/SourceNode.tsx`
- `web/ui/src/nodes/TargetNode.tsx`
- `web/ui/src/nodes/TransformNode.tsx`
- `web/ui/src/components/editor/ETLBlockNode.tsx`

---

### 🔴 BLOQUANT 3 — WebSocket suivi temps réel absent côté UI

Le backend doit diffuser les événements d'exécution bloc par bloc. Côté frontend,
aucun hook WebSocket n'est branché sur le canvas.

**Contrat de message WebSocket côté Go à implémenter :**

```go
// internal/orchestrator/ws.go
package orchestrator

type BlockEvent struct {
    RunID     string `json:"runId"`
    BlockID   string `json:"blockId"`
    Status    string `json:"status"`   // "running" | "succeeded" | "failed" | "skipped"
    RowsIn    int64  `json:"rowsIn"`
    RowsOut   int64  `json:"rowsOut"`
    DurationMs int64 `json:"durationMs"`
    Error     string `json:"error,omitempty"`
}
```

**Hook React à créer :**

```ts
// web/ui/src/hooks/useRunWebSocket.ts
import { useEffect, useRef } from "react";
import { useRunStore } from "@/store/runStore";

export function useRunWebSocket(runId: string | null) {
  const wsRef = useRef<WebSocket | null>(null);
  const setBlockStatus = useRunStore((s) => s.setBlockStatus);

  useEffect(() => {
    if (!runId) return;
    const ws = new WebSocket(`${import.meta.env.VITE_WS_URL}/api/v1/runs/${runId}/ws`);
    wsRef.current = ws;

    ws.onmessage = (e) => {
      const event = JSON.parse(e.data) as BlockEvent;
      setBlockStatus(event.blockId, event.status, event.rowsOut);
    };

    return () => ws.close();
  }, [runId]);
}
```

**Store Zustand à créer :**

```ts
// web/ui/src/store/runStore.ts
import { create } from "zustand";

interface BlockStatus {
  status: "idle" | "running" | "succeeded" | "failed" | "skipped";
  rowsOut: number;
}

interface RunStore {
  blockStatuses: Record<string, BlockStatus>;
  setBlockStatus: (blockId: string, status: BlockStatus["status"], rowsOut: number) => void;
  reset: () => void;
}

export const useRunStore = create<RunStore>((set) => ({
  blockStatuses: {},
  setBlockStatus: (blockId, status, rowsOut) =>
    set((s) => ({ blockStatuses: { ...s.blockStatuses, [blockId]: { status, rowsOut } } })),
  reset: () => set({ blockStatuses: {} }),
}));
```

---

### 🔴 BLOQUANT 4 — `RunProgressPanel` absent du canvas

Le canvas `EditorPage.tsx` n'a pas de composant affichant l'état d'avancement bloc par bloc
pendant un run. C'est le pendant visuel du `DataPreviewPanel`.

**Composant à créer :**

```tsx
// web/ui/src/components/editor/RunProgressPanel.tsx
import { useRunStore } from "@/store/runStore";

const statusColor: Record<string, string> = {
  idle:      "bg-gray-200 text-gray-600",
  running:   "bg-blue-100 text-blue-700 animate-pulse",
  succeeded: "bg-green-100 text-green-700",
  failed:    "bg-red-100 text-red-700",
  skipped:   "bg-yellow-100 text-yellow-700",
};

export function RunProgressPanel({ nodeIds }: { nodeIds: string[] }) {
  const blockStatuses = useRunStore((s) => s.blockStatuses);

  return (
    <div className="flex flex-col gap-1 p-2 bg-white border rounded shadow text-xs">
      <p className="font-semibold text-gray-700 mb-1">Exécution en cours</p>
      {nodeIds.map((id) => {
        const s = blockStatuses[id] ?? { status: "idle", rowsOut: 0 };
        return (
          <div key={id} className={`flex justify-between px-2 py-1 rounded ${statusColor[s.status]}`}>
            <span className="font-mono">{id}</span>
            <span>{s.status} · {s.rowsOut} lignes</span>
          </div>
        );
      })}
    </div>
  );
}
```

---

### 🟡 IMPORTANT 5 — `BlockPalette` : catalogue non synchronisé avec le backend

La palette affiche des blocs en dur côté frontend. Elle doit consommer
`GET /api/v1/blocks/catalogue` pour être toujours en phase avec le registre Go.

**Endpoint Go à exposer :**

```go
// api/handlers/catalogue_handler.go
func (h *Handler) GetCatalogue(w http.ResponseWriter, r *http.Request) {
    catalogue := blocks.Registry().All() // retourne []blocks.BlockMeta
    render.JSON(w, r, catalogue)
}
```

**Structure `BlockMeta` :**

```go
// internal/etl/blocks/catalogue.go
type BlockMeta struct {
    Type        string `json:"type"`        // "transform.filter"
    Category    string `json:"category"`    // "transform" | "source" | "target"
    Label       string `json:"label"`       // "Filtre"
    Description string `json:"description"` // "Filtre les lignes selon une condition"
    MinInputs   int    `json:"minInputs"`
    MaxInputs   int    `json:"maxInputs"`
    MinOutputs  int    `json:"minOutputs"`
    MaxOutputs  int    `json:"maxOutputs"`
}
```

**Hook React à créer :**

```ts
// web/ui/src/hooks/useCatalogue.ts
import { useEffect, useState } from "react";
import { apiClient } from "@/api/client";
import type { BlockMeta } from "@/types/catalogue";

export function useCatalogue() {
  const [catalogue, setCatalogue] = useState<BlockMeta[]>([]);
  useEffect(() => {
    apiClient.get<BlockMeta[]>("/api/v1/blocks/catalogue")
      .then((r) => setCatalogue(r.data));
  }, []);
  return catalogue;
}
```

---

### 🟡 IMPORTANT 6 — `NodeConfigPanel.tsx` : cohérence paramètres UI ↔ Go

`NodeConfigPanel.tsx` (24KB) doit exposer des champs dont les **clés correspondent exactement**
aux `Params` attendus par chaque bloc Go.

**Référence de correspondance obligatoire :**

| Bloc Go | Param Go (`bctx.Params["..."]`) | Champ UI attendu |
|---|---|---|
| `source.csv` | `path`, `delimiter`, `encoding`, `hasHeader` | Champs texte + checkbox |
| `source.postgres` | `connectionRef`, `query` | Sélecteur connexion + textarea |
| `target.postgres` | `connectionRef`, `table`, `mode` | Sélecteur + texte + radio insert/upsert/truncate |
| `transform.filter` | `condition` | Champ texte "Condition" |
| `transform.filter_advanced` | `condition_true`, `condition_false` | 2 champs texte |
| `transform.select` | `columns` | Multi-select colonnes |
| `transform.cast` | `column`, `targetType` | Sélecteur colonne + type |
| `transform.add_column` | `name`, `expression` | Nom + expression |
| `transform.split` | `conditions` | Textarea (conditions, une par ligne) |
| `transform.pivot` | `groupBy`, `pivotColumn`, `valueColumn` | 3 sélecteurs |
| `transform.unpivot` | `columns`, `keyName`, `valueName` | Multi-select + 2 champs |
| `transform.join` | `leftKey`, `rightKey`, `type` | 2 champs + radio inner/left/right/full |
| `transform.dedup` | `keys` | Multi-select colonnes clés |
| `transform.sort` | `columns`, `order` | Multi-select + asc/desc |
| `transform.aggregate` | `groupBy`, `aggregations` | Multi-select + liste agg |

---

### 🟡 IMPORTANT 7 — `EditorPage.tsx` : sauvegarde XML non déclenchée

L'éditeur doit envoyer `PUT /api/v1/projects/{id}` avec le DAG sérialisé à chaque
modification du canvas (ajout de nœud, connexion d'arête, changement de paramètre).

**Intégration dans `EditorPage.tsx` :**

```ts
const saveProject = useCallback(
  debounce(async (nodes: Node[], edges: Edge[]) => {
    const project = serializeDAG(nodes, edges); // nodes/edges → contracts.Project
    await apiClient.put(`/api/v1/projects/${projectId}`, project);
  }, 800),
  [projectId]
);

// Déclencher à chaque changement
useEffect(() => {
  saveProject(nodes, edges);
}, [nodes, edges]);
```

**Fonction `serializeDAG` à créer :**

```ts
// web/ui/src/utils/serializeDAG.ts
import type { Node, Edge } from "@xyflow/react";
import type { Project } from "@/types/project";

export function serializeDAG(nodes: Node[], edges: Edge[]): Project {
  return {
    nodes: nodes.map((n) => ({
      id: n.id,
      type: n.data.blockType as string,
      label: n.data.label as string,
      params: n.data.params as Record<string, string>,
      position: { x: n.position.x, y: n.position.y },
    })),
    edges: edges.map((e) => ({
      id: e.id,
      source: e.source,
      sourceHandle: e.sourceHandle ?? "out",
      target: e.target,
      targetHandle: e.targetHandle ?? "in",
    })),
  };
}
```

---

### 🟡 IMPORTANT 8 — `ConnectionsPage.tsx` : switch d'environnement global absent

La page connexions affiche les connexions mais n'expose pas le sélecteur
`Dev / Préprod / Prod` qui appelle `PUT /api/v1/environment`.

**Composant à ajouter dans `ConnectionsPage.tsx` :**

```tsx
// web/ui/src/components/connections/EnvSwitcher.tsx
import { useState } from "react";
import { apiClient } from "@/api/client";

const ENVS = ["dev", "preprod", "prod"] as const;
type Env = typeof ENVS[number];

export function EnvSwitcher() {
  const [active, setActive] = useState<Env>("dev");

  const switchEnv = async (env: Env) => {
    await apiClient.put("/api/v1/environment", { env });
    setActive(env);
  };

  return (
    <div className="flex gap-2 items-center">
      <span className="text-sm font-medium text-gray-600">Environnement actif :</span>
      {ENVS.map((env) => (
        <button
          key={env}
          onClick={() => switchEnv(env)}
          className={`px-3 py-1 rounded text-sm font-semibold transition
            ${active === env
              ? "bg-blue-600 text-white"
              : "bg-gray-100 text-gray-700 hover:bg-gray-200"
            }`}
        >
          {env.charAt(0).toUpperCase() + env.slice(1)}
        </button>
      ))}
    </div>
  );
}
```

---

### 🟡 IMPORTANT 9 — `RunHistory.tsx` : logs par bloc non affichés

`RunHistory.tsx` (2117b) liste les runs mais n'affiche pas le détail par bloc
(`rowsIn`, `rowsOut`, `durationMs`, erreurs) retourné par `ExecutionReport`.

**Structure `ExecutionReport` côté Go (rappel) :**

```go
// internal/etl/engine/executor.go
type BlockResult struct {
    BlockID    string        `json:"blockId"`
    Status     string        `json:"status"`
    RowsIn     int64         `json:"rowsIn"`
    RowsOut    int64         `json:"rowsOut"`
    Duration   time.Duration `json:"durationMs"`
    Error      string        `json:"error,omitempty"`
}

type ExecutionReport struct {
    RunID     string        `json:"runId"`
    Status    string        `json:"status"`
    StartedAt time.Time     `json:"startedAt"`
    EndedAt   time.Time     `json:"endedAt"`
    Results   []BlockResult `json:"results"`
}
```

**Composant detail à ajouter dans `RunHistory.tsx` :**

```tsx
function RunDetail({ report }: { report: ExecutionReport }) {
  return (
    <table className="w-full text-xs mt-2 border rounded">
      <thead className="bg-gray-100">
        <tr>
          <th className="p-2 text-left">Bloc</th>
          <th className="p-2">Statut</th>
          <th className="p-2">Lignes entrée</th>
          <th className="p-2">Lignes sortie</th>
          <th className="p-2">Durée (ms)</th>
          <th className="p-2">Erreur</th>
        </tr>
      </thead>
      <tbody>
        {report.results.map((r) => (
          <tr key={r.blockId} className="border-t">
            <td className="p-2 font-mono">{r.blockId}</td>
            <td className="p-2 text-center">{r.status}</td>
            <td className="p-2 text-right">{r.rowsIn}</td>
            <td className="p-2 text-right">{r.rowsOut}</td>
            <td className="p-2 text-right">{r.durationMs}</td>
            <td className="p-2 text-red-600">{r.error ?? "—"}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
```

---

## Plan d'action pour finaliser la Phase 8

### Sprint A — Auth & Navigation (1 jour)

- [ ] Créer `web/ui/src/components/auth/AuthGuard.tsx`
- [ ] Vérifier `web/ui/src/store/authStore.ts` : persistance JWT (`localStorage`)
- [ ] Wrapper toutes les routes privées dans `App.tsx` avec `<AuthGuard>`
- [ ] Vérifier que `Login.tsx` appelle `POST /api/v1/auth/login` et stocke le token
- [ ] Tester : accès à `/editor/xxx` sans token → redirection `/login`

### Sprint B — Canvas & Palette (2 jours)

- [ ] **Fix imports ReactFlow** : remplacer `from "reactflow"` → `from "@xyflow/react"` dans tous les `.tsx`
- [ ] **Fix style import** : `"reactflow/dist/style.css"` → `"@xyflow/react/dist/style.css"`
- [ ] Exposer `GET /api/v1/blocks/catalogue` côté Go (handler + `BlockMeta`)
- [ ] Créer `web/ui/src/hooks/useCatalogue.ts` et brancher sur `BlockPalette`
- [ ] Vérifier drag-and-drop depuis palette → canvas : nœud créé avec `data.blockType` correct
- [ ] Vérifier que les arêtes entre nœuds respectent les contraintes de ports (1→N pour `split`)

### Sprint C — Sauvegarde & Chargement XML (1 jour)

- [ ] Créer `web/ui/src/utils/serializeDAG.ts`
- [ ] Brancher `debounce(saveProject)` dans `EditorPage.tsx` sur `onNodesChange` + `onEdgesChange`
- [ ] Implémenter `deserializeDAG` pour charger `GET /api/v1/projects/{id}` → nodes/edges React Flow
- [ ] Tester : créer un projet, ajouter 3 blocs, rafraîchir → le canvas est restauré à l'identique

### Sprint D — Configuration des blocs (1 jour)

- [ ] Vérifier `NodeConfigPanel.tsx` sur les 15 types de blocs (table de correspondance §IMPORTANT 6)
- [ ] S'assurer que les changements de paramètres déclenchent bien `onNodesChange` (pour la sauvegarde auto)
- [ ] Ajouter les champs manquants pour `source.csv`, `source.postgres`, `target.postgres`

### Sprint E — Exécution temps réel WebSocket (2 jours)

- [ ] Implémenter côté Go : `internal/orchestrator/ws.go` — diffusion des `BlockEvent` via WebSocket
  - Enregistrer le handler `GET /api/v1/runs/{id}/ws` dans le router `chi`
  - Émettre un `BlockEvent` avant et après chaque `block.Run()` dans `executor.go`
- [ ] Créer `web/ui/src/store/runStore.ts`
- [ ] Créer `web/ui/src/hooks/useRunWebSocket.ts`
- [ ] Créer `web/ui/src/components/editor/RunProgressPanel.tsx`
- [ ] Intégrer `RunProgressPanel` dans `EditorPage.tsx` (panneau latéral droit pendant le run)
- [ ] Tester : lancer un run depuis l'UI → chaque nœud du canvas passe successivement de `idle` → `running` → `succeeded`

### Sprint F — Connexions & Switch Env (1 jour)

- [ ] Créer `web/ui/src/components/connections/EnvSwitcher.tsx`
- [ ] Intégrer `EnvSwitcher` dans `ConnectionsPage.tsx` (header de la page)
- [ ] Vérifier que `ConnectionsPage.tsx` appelle bien `GET /api/v1/connections` pour lister
- [ ] Vérifier le formulaire de création / édition d'une connexion avec ses profils Dev/Préprod/Prod
- [ ] Tester : créer une connexion PostgreSQL, switcher en `prod`, relancer un run → les paramètres de prod sont utilisés

### Sprint G — Historique des runs (0.5 jour)

- [ ] Brancher `RunHistory.tsx` sur `GET /api/v1/projects/{id}/runs`
- [ ] Ajouter le composant `RunDetail` pour afficher l'`ExecutionReport` au clic sur un run
- [ ] Afficher statut global, durée totale, lignes traitées par bloc

---

## Nouveaux fichiers à créer

| Fichier | Rôle |
|---|---|
| `web/ui/src/components/auth/AuthGuard.tsx` | Garde de route JWT |
| `web/ui/src/store/authStore.ts` | Store Zustand token JWT |
| `web/ui/src/store/runStore.ts` | Store Zustand statuts blocs pendant run |
| `web/ui/src/hooks/useCatalogue.ts` | Fetch catalogue blocs depuis API |
| `web/ui/src/hooks/useRunWebSocket.ts` | Connexion WebSocket run temps réel |
| `web/ui/src/components/editor/RunProgressPanel.tsx` | Panneau suivi run dans canvas |
| `web/ui/src/components/connections/EnvSwitcher.tsx` | Switch Dev/Préprod/Prod |
| `web/ui/src/utils/serializeDAG.ts` | nodes/edges → `contracts.Project` |
| `web/ui/src/utils/deserializeDAG.ts` | `contracts.Project` → nodes/edges |
| `web/ui/src/types/catalogue.ts` | Types TypeScript `BlockMeta` |
| `internal/orchestrator/ws.go` | Diffusion WebSocket `BlockEvent` |
| `api/handlers/catalogue_handler.go` | `GET /api/v1/blocks/catalogue` |

---

## Interfaces Go impactées

| Interface / struct | Fichier | Modification |
|---|---|---|
| `engine.Executor` | `internal/etl/engine/executor.go` | Émettre `BlockEvent` via channel WebSocket après chaque bloc |
| `orchestrator.Service` | `internal/orchestrator/service.go` | Accepter un `chan BlockEvent` en paramètre de `RunProject()` |
| Router `chi` | `api/handlers/` ou `internal/app/` | Enregistrer `GET /api/v1/runs/{id}/ws` + `GET /api/v1/blocks/catalogue` |
| `blocks.Registry` | `internal/etl/blocks/catalogue.go` | Exposer `All() []BlockMeta` sur le registre |

---

## Checklist finale Phase 8 — "Definition of Done"

### Backend Go
- [ ] `go build ./...` passe sans erreur ni warning
- [ ] `GET /api/v1/blocks/catalogue` retourne tous les blocs du registre en JSON
- [ ] `GET /api/v1/runs/{id}/ws` diffuse les `BlockEvent` en temps réel pendant l'exécution
- [ ] `PUT /api/v1/environment` bascule l'env actif et est pris en compte au prochain run
- [ ] `POST /api/v1/auth/login` retourne un JWT valide

### Frontend React
- [ ] `npm run build` passe sans erreur
- [ ] Aucun import `from "reactflow"` restant (tous migrés vers `@xyflow/react`)
- [ ] Accès aux routes privées sans JWT → redirection `/login`
- [ ] Palette chargée dynamiquement depuis `GET /api/v1/blocks/catalogue`
- [ ] Drag-and-drop bloc palette → canvas fonctionne pour les 3 catégories (source/transform/target)
- [ ] Configuration d'un bloc au clic → `NodeConfigPanel` affiche les bons champs
- [ ] Sauvegarde automatique au canvas (debounce 800ms) → `projects/{id}/project.xml` mis à jour
- [ ] Lancement d'un run → `RunProgressPanel` affiche l'avancement bloc par bloc en temps réel
- [ ] `ConnectionsPage` : création connexion multi-env + switch env global fonctionnels
- [ ] `RunHistory` : liste des runs + détail `ExecutionReport` par bloc au clic

### Pipelines de validation end-to-end
- [ ] Login → Dashboard → ouvrir un projet → canvas restauré avec tous les blocs et arêtes
- [ ] Ajouter `source.csv → transform.filter → target.csv` → sauvegarder → relancer → fichier de sortie correct
- [ ] Switcher env `dev` → `prod` → run utilise les params de connexion production
- [ ] WebSocket : fermer l'onglet pendant un run → reconnexion → statuts cohérents

### Déploiement
- [ ] `docker-compose up` démarre sans erreur (server + postgres)
- [ ] Frontend servi par le backend Go (assets statiques compilés dans `web/assets/`)
- [ ] Variables d'environnement `VITE_API_URL` et `VITE_WS_URL` correctement configurées

---

## Architecture rappel — Flux UI complet Phase 8
```
Utilisateur
│
▼
Login.tsx ──POST /api/v1/auth/login──▶ JWT stocké dans authStore
│
▼
ProjectsPage.tsx ──GET /api/v1/projects──▶ liste des projets
│ clic sur un projet
▼
EditorPage.tsx
├── BlockPalette ◀── GET /api/v1/blocks/catalogue
├── ReactFlow canvas
│ ├── SourceNode / TransformNode / TargetNode
│ └── NodeConfigPanel (au clic sur un nœud)
│
│ onNodesChange / onEdgesChange (debounce 800ms)
▼
PUT /api/v1/projects/{id} ──▶ project_handler.go ──▶ xml/serializer ──▶ projects/{id}/project.xml
│
│ clic "Exécuter"
▼
POST /api/v1/runs ──▶ orchestrator.RunProject()
│ │ xml/store.Load()
│ │ engine.Executor.Execute()
│ │ └─ BlockEvent ──▶ chan ──▶ WebSocket hub
│ ▼
│ jobs.Repository.SetStatus("succeeded")
│
▼ WebSocket ws://…/api/v1/runs/{id}/ws
useRunWebSocket.ts ──▶ runStore.setBlockStatus()
│
▼
RunProgressPanel.tsx ──▶ statut coloré par bloc dans le canvas

ConnectionsPage.tsx
├── GET /api/v1/connections ──▶ liste des connexions XML
├── EnvSwitcher ──PUT /api/v1/environment──▶ switch global Dev/Préprod/Prod
└── Formulaire créer/éditer connexion ──POST|PUT /api/v1/connections──▶ connections/*.xml

RunHistory.tsx ──GET /api/v1/projects/{id}/runs──▶ liste runs
└── clic sur un run ──GET /api/v1/runs/{id}──▶ ExecutionReport détail par bloc
```

---

*Document généré automatiquement par analyse du code source — à mettre à jour à chaque sprint.*