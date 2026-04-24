# Étape 6 — Gestionnaire de connexions multi-env : État détaillé et tâches restantes

> Généré le 2026-04-24 · Basé sur un scan complet du code source

---

## Résumé de la Phase 6

La Phase 6 a pour objectif de **compléter et solidifier le gestionnaire de connexions multi-environnements** :
CRUD XML complet, résolution sécurisée des secrets, test de connexion réelle (ping DB),
switch global d'environnement persisté, et interface UI de gestion des profils par connexion.

**État global : squelette fonctionnel ✅ — intégration complète à finaliser ⚠️**

---

## Ce qui est déjà en place (Phases 0 → 5)

### ✅ Modèle de données `internal/connections/model.go`

```go
type Connection struct {
    ID      string             `xml:"id,attr" json:"id"`
    Name    string             `xml:"name,attr" json:"name"`
    Type    string             `xml:"type,attr" json:"type"` // postgres, mysql, mssql, rest
    Envs    map[string]ConnEnv `xml:"-" json:"envs"`
    EnvList []ConnEnv          `xml:"environments>env" json:"-"`
}

type ConnEnv struct {
    Name      string `xml:"name,attr" json:"name"`
    Host      string `xml:"host,attr" json:"host"`
    Port      int    `xml:"port,attr" json:"port"`
    Database  string `xml:"db,attr" json:"database"`
    User      string `xml:"user,attr" json:"user"`
    SecretRef string `xml:"secretRef,attr" json:"secretRef"` // ${ENV_VAR} ou vault:secret/path
}
```

Le modèle est **complet** — il couvre tous les champs nécessaires pour les connexions SQL.
La méthode `DSN(password string) string` est présente sur `ConnEnv`.

---

### ✅ Manager `internal/connections/manager/manager.go` (3283b)

| Méthode | État |
|---|---|
| `New(connsDir, activeEnv)` | ✅ Crée le manager, crée le répertoire, charge tout |
| `loadAll()` | ✅ Parse tous les `.xml` du répertoire de connexions |
| `Get(id)` | ✅ Retourne une connexion par ID avec verrou lecture |
| `List()` | ✅ Retourne toutes les connexions |
| `Save(conn)` | ✅ Sérialise en XML + écrit sur disque |
| `Delete(id)` | ✅ Supprime la map + le fichier XML |
| `SwitchEnv(env)` | ✅ Bascule `ActiveEnv` en mémoire uniquement |

**Problème identifié :** `SwitchEnv` ne persiste **pas** l'environnement actif sur disque — si le serveur
redémarre, il revient à la valeur par défaut du fichier de config YAML.

---

### ✅ Resolver `internal/connections/resolver/resolver.go`

```go
func Resolve(mgr *manager.Manager, connID string) (*ResolvedConn, error)
func ResolveWithEnv(conn *connections.Connection, env string) (*ResolvedConn, error)
```

Fonctionnel : résout l'env actif → appelle `secrets.Resolve()` → retourne un `ResolvedConn`
avec `DSN` prêt à l'emploi.

---

### ✅ Secrets `internal/connections/secrets/resolver.go`

```go
func Resolve(ref string) (string, error)
// Formats supportés :
//   ${ENV_VAR}     → os.Getenv()
//   vault:...      → ❌ non implémenté — retourne une erreur
//   texte brut     → retourné tel quel (dev local uniquement)
```

---

### ✅ Handler HTTP `api/handlers/connection_handler.go`

| Route | Handler | État |
|---|---|---|
| `GET /api/v1/connections` | `List` | ✅ Fonctionnel |
| `POST /api/v1/connections` | `Create` | ✅ Fonctionnel (UUID auto) |
| `GET /api/v1/connections/{connID}` | `Get` | ✅ Fonctionnel |
| `PUT /api/v1/connections/{connID}` | `Update` | ✅ Fonctionnel |
| `DELETE /api/v1/connections/{connID}` | `Delete` | ✅ Fonctionnel |
| `POST /api/v1/connections/{connID}/test` | `Test` | ⚠️ Résout seulement — pas de ping réel |
| `GET /api/v1/environment` | `GetEnv` | ✅ Fonctionnel |
| `PUT /api/v1/environment` | `SwitchEnv` | ⚠️ En mémoire uniquement — pas persisté |

---

### ✅ Frontend `web/ui/src/pages/ConnectionsPage.tsx` (5357b)

La page est présente et fonctionnelle sur le plan visuel. Elle gère :
- Listage des connexions avec type affiché
- Switch d'environnement global (boutons DEV / PREPROD / PROD)
- Création d'une connexion (nom + type) via une modale
- Suppression d'une connexion
- Test d'une connexion (affiche host/db/env)

**Ce qui manque dans l'UI :**
- Aucun formulaire d'édition des **profils d'environnement** (host, port, db, user, secretRef)
  par connexion → impossible de configurer une connexion utile depuis l'UI
- Pas de page de détail / édition complète d'une connexion
- `handleCreate` crée une connexion avec `envs: {}` — aucun profil — donc `Test` échoue
  systématiquement

---

### ✅ Client API `web/ui/src/api/client.ts`

Toutes les fonctions sont présentes :
`listConnections`, `getConnection`, `createConnection`, `updateConnection`,
`deleteConnection`, `testConnection`, `getEnvironment`, `switchEnvironment`.

---

### ✅ Types TypeScript `web/ui/src/types/api.ts`

```typescript
export interface ConnEnv {
  name: string; host: string; port: number;
  database: string; user: string; secretRef: string;
}

export interface Connection {
  id: string; name: string; type: string;
  envs: Record<string, ConnEnv>;
}
```

Types **complets et corrects** — correspondent exactement aux structs Go.

---

## Problèmes bloquants identifiés

### 🔴 BLOQUANT 1 — `Test` ne fait pas de vrai ping base de données

`connection_handler.go` → méthode `Test` :

```go
// ❌ ACTUEL — résout juste les paramètres, ne tente AUCUNE connexion réelle
func (h *ConnectionHandler) Test(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "connID")
    rc, err := resolver.Resolve(h.mgr, id)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, map[string]string{
        "status": "ok",
        "type":   rc.Type,
        "host":   rc.Host,
        "db":     rc.Database,
        "env":    h.mgr.ActiveEnv,
    })
}
```

**Ce handler répond toujours `"status": "ok"` dès que les paramètres sont résolvables**,
même si la base de données est inaccessible ou que les credentials sont faux.

**Fix obligatoire :**

```go
// ✅ CORRECT — ping réel selon le type de connexion
func (h *ConnectionHandler) Test(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "connID")
    rc, err := resolver.Resolve(h.mgr, id)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    if err := pingConnection(ctx, rc); err != nil {
        writeError(w, http.StatusBadGateway, fmt.Sprintf("ping échoué: %v", err))
        return
    }

    writeJSON(w, http.StatusOK, map[string]string{
        "status": "ok",
        "type":   rc.Type,
        "host":   rc.Host,
        "db":     rc.Database,
        "env":    h.mgr.ActiveEnv,
        "latency": "< 5s",
    })
}

// pingConnection ouvre et ferme une connexion DB pour vérifier l'accessibilité.
func pingConnection(ctx context.Context, rc *resolver.ResolvedConn) error {
    switch rc.Type {
    case "postgres":
        db, err := sql.Open("pgx", rc.DSN)
        if err != nil {
            return err
        }
        defer db.Close()
        return db.PingContext(ctx)
    case "mysql":
        dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", rc.User, rc.Password, rc.Host, rc.Port, rc.Database)
        db, err := sql.Open("mysql", dsn)
        if err != nil {
            return err
        }
        defer db.Close()
        return db.PingContext(ctx)
    case "mssql":
        u := &url.URL{
            Scheme: "sqlserver", User: url.UserPassword(rc.User, rc.Password),
            Host: fmt.Sprintf("%s:%d", rc.Host, rc.Port),
            RawQuery: url.Values{"database": {rc.Database}}.Encode(),
        }
        db, err := sql.Open("sqlserver", u.String())
        if err != nil {
            return err
        }
        defer db.Close()
        return db.PingContext(ctx)
    case "rest":
        req, _ := http.NewRequestWithContext(ctx, http.MethodGet, rc.Host, nil)
        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            return err
        }
        resp.Body.Close()
        return nil
    default:
        return fmt.Errorf("type de connexion inconnu : %s", rc.Type)
    }
}
```

**Fichier à modifier :** `api/handlers/connection_handler.go`
**Imports à ajouter :** `context`, `database/sql`, `fmt`, `net/http`, `net/url`, `time`

---

### 🔴 BLOQUANT 2 — `SwitchEnv` non persisté : perte au redémarrage

`manager.go` → `SwitchEnv` :

```go
// ❌ ACTUEL — purement en mémoire
func (m *Manager) SwitchEnv(env string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.ActiveEnv = env
}
```

À chaque redémarrage du serveur, `ActiveEnv` revient à la valeur lue dans `config.dev.yaml`
(`activeEnv: dev`). Un opérateur qui bascule en `prod` perd ce paramètre sans avertissement.

**Fix obligatoire — écriture dans un fichier `.env-state.json` dans `connsDir` :**

```go
// ✅ CORRECT — persiste l'env actif sur disque
func (m *Manager) SwitchEnv(env string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.ActiveEnv = env
    return m.persistEnvState()
}

func (m *Manager) persistEnvState() error {
    type state struct {
        ActiveEnv string `json:"activeEnv"`
    }
    data, err := json.Marshal(state{ActiveEnv: m.ActiveEnv})
    if err != nil {
        return err
    }
    return os.WriteFile(filepath.Join(m.connsDir, ".env-state.json"), data, 0o644)
}

// loadEnvState est appelé dans New() après loadAll()
func (m *Manager) loadEnvState() {
    type state struct {
        ActiveEnv string `json:"activeEnv"`
    }
    data, err := os.ReadFile(filepath.Join(m.connsDir, ".env-state.json"))
    if err != nil {
        return // fichier absent = première fois, on garde la valeur par défaut
    }
    var s state
    if json.Unmarshal(data, &s) == nil && s.ActiveEnv != "" {
        m.ActiveEnv = s.ActiveEnv
    }
}
```

**Modifier `New()` pour appeler `loadEnvState()` après `loadAll()` :**

```go
func New(connsDir string, activeEnv string) (*Manager, error) {
    // ...
    m := &Manager{connsDir: connsDir, connections: make(map[string]*connections.Connection), ActiveEnv: activeEnv}
    if err := m.loadAll(); err != nil {
        return nil, err
    }
    m.loadEnvState() // ← ajouter cette ligne
    return m, nil
}
```

**Adapter le handler `SwitchEnv` :**

```go
func (h *ConnectionHandler) SwitchEnv(w http.ResponseWriter, r *http.Request) {
    var body struct{ Env string `json:"env"` }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Env == "" {
        writeError(w, http.StatusBadRequest, "champ 'env' manquant (dev|preprod|prod)")
        return
    }
    if err := h.mgr.SwitchEnv(body.Env); err != nil {
        writeError(w, http.StatusInternalServerError, "impossible de persister l'env: "+err.Error())
        return
    }
    h.log.Info().Str("env", body.Env).Msg("environnement basculé et persisté")
    writeJSON(w, http.StatusOK, map[string]string{"activeEnv": body.Env})
}
```

**Fichiers à modifier :**
- `internal/connections/manager/manager.go`
- `api/handlers/connection_handler.go`

---

### 🔴 BLOQUANT 3 — UI : impossible de configurer les profils d'environnement

`ConnectionsPage.tsx` crée des connexions avec `envs: {}`. Le formulaire de création
ne propose que `name` et `type`. Il n'existe aucune modale / page pour éditer
les profils `dev`, `preprod`, `prod` d'une connexion.

**Résultat concret :** Toute connexion créée depuis l'UI a zéro profil.
`Test` échoue avec `"resolver: profil 'dev' introuvable"`. La feature est inutilisable.

**Composant `ConnectionEnvForm` à créer :**
`web/ui/src/components/connections/ConnectionEnvForm.tsx`

```tsx
import { useState } from 'react'
import type { ConnEnv } from '@/types/api'

interface Props {
  envName: string
  initial?: ConnEnv
  onSave: (env: ConnEnv) => void
  onCancel: () => void
}

const DEFAULT: ConnEnv = { name: '', host: '', port: 5432, database: '', user: '', secretRef: '' }

export default function ConnectionEnvForm({ envName, initial, onSave, onCancel }: Props) {
  const [form, setForm] = useState<ConnEnv>({ ...DEFAULT, ...initial, name: envName })
  const set = (k: keyof ConnEnv, v: string | number) => setForm(f => ({ ...f, [k]: v }))

  return (
    <div className="space-y-3">
      <h3 className="text-sm font-semibold text-gray-300 uppercase tracking-wider">{envName}</h3>
      {(['host', 'database', 'user'] as (keyof ConnEnv)[]).map(field => (
        <div key={field}>
          <label className="block text-xs text-gray-400 mb-1 capitalize">{field}</label>
          <input
            className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-gray-100"
            value={form[field] as string}
            onChange={e => set(field, e.target.value)}
            placeholder={field === 'host' ? 'db.example.com' : ''}
          />
        </div>
      ))}
      <div>
        <label className="block text-xs text-gray-400 mb-1">Port</label>
        <input
          type="number"
          className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-gray-100"
          value={form.port}
          onChange={e => set('port', parseInt(e.target.value) || 5432)}
        />
      </div>
      <div>
        <label className="block text-xs text-gray-400 mb-1">
          Secret Ref <span className="text-gray-600">(ex: ${'{'}DB_PASSWORD{'}'})</span>
        </label>
        <input
          className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm font-mono text-gray-100"
          value={form.secretRef}
          onChange={e => set('secretRef', e.target.value)}
          placeholder="${DB_PASSWORD}"
        />
      </div>
      <div className="flex justify-end gap-2 pt-1">
        <button onClick={onCancel} className="px-3 py-1.5 text-sm text-gray-400 hover:text-gray-200">
          Annuler
        </button>
        <button
          onClick={() => onSave(form)}
          disabled={!form.host || !form.database || !form.user}
          className="px-4 py-1.5 text-sm bg-brand-600 hover:bg-brand-500 disabled:opacity-40 text-white rounded-lg"
        >
          Enregistrer
        </button>
      </div>
    </div>
  )
}
```

**Page de détail `ConnectionDetailPage.tsx` à créer :**
`web/ui/src/pages/ConnectionDetailPage.tsx`

```tsx
// Route : /connections/:connID
// Affiche les 3 onglets DEV / PREPROD / PROD
// Chaque onglet ouvre ConnectionEnvForm pré-rempli
// PUT /api/v1/connections/:connID à la sauvegarde
```

**Ajouter la route dans `App.tsx` :**

```tsx
// ❌ ACTUEL — aucune route /connections/:connID
// ✅ À AJOUTER
<Route path="/connections/:connID" element={<ConnectionDetailPage />} />
```

---

### 🟡 IMPORTANT 4 — Vault non implémenté dans `secrets/resolver.go`

```go
// ❌ ACTUEL — retourne une erreur
if strings.HasPrefix(ref, "vault:") {
    return "", fmt.Errorf("secrets: intégration Vault non encore implémentée (ref: %s)", ref)
}
```

Pour le MVP, il suffit d'une **interface `SecretProvider`** extensible sans tout réécrire :

```go
// internal/connections/secrets/provider.go — NOUVEAU FICHIER

package secrets

// Provider résout un secret à partir d'une référence.
type Provider interface {
    Resolve(ref string) (string, error)
}

// EnvProvider lit depuis os.Getenv (défaut).
type EnvProvider struct{}

func (EnvProvider) Resolve(ref string) (string, error) {
    return Resolve(ref) // délègue au resolver existant
}

// VaultProvider — implémentation future.
type VaultProvider struct {
    Address string
    Token   string
}

func (v VaultProvider) Resolve(ref string) (string, error) {
    // À implémenter : appel HTTP à l'API Vault
    // GET {Address}/v1/{path} avec header X-Vault-Token: {Token}
    return "", fmt.Errorf("vault: non encore implémenté")
}
```

---

### 🟡 IMPORTANT 5 — Validation des valeurs d'`env` dans `SwitchEnv`

Le handler accepte n'importe quelle chaîne comme valeur d'env, y compris des valeurs invalides.

**Fix — whitelist dans le handler :**

```go
var validEnvs = map[string]bool{"dev": true, "preprod": true, "prod": true}

func (h *ConnectionHandler) SwitchEnv(w http.ResponseWriter, r *http.Request) {
    var body struct{ Env string `json:"env"` }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Env == "" {
        writeError(w, http.StatusBadRequest, "champ 'env' manquant")
        return
    }
    if !validEnvs[body.Env] {
        writeError(w, http.StatusBadRequest, fmt.Sprintf("env invalide '%s': valeurs acceptées dev|preprod|prod", body.Env))
        return
    }
    // ...
}
```

---

### 🟡 IMPORTANT 6 — `inject_connections.go` du moteur ETL à connecter au Manager

`internal/etl/engine/inject_connections.go` injecte les connexions dans le `BlockContext`.
Il doit utiliser le `resolver` pour passer un DSN résolu (et non les paramètres bruts)
au moment de l'exécution.

**Vérifier que `InjectConnections()` appelle bien `resolver.Resolve(mgr, connRef)` :**

```go
// internal/etl/engine/inject_connections.go — vérification obligatoire
func InjectConnections(ctx *contracts.BlockContext, mgr *manager.Manager) error {
    if ctx.ConnectionRef == "" {
        return nil
    }
    rc, err := resolver.Resolve(mgr, ctx.ConnectionRef)
    if err != nil {
        return fmt.Errorf("inject_connections: %w", err)
    }
    ctx.Connection = &contracts.ResolvedConnection{
        Type: rc.Type,
        DSN:  rc.DSN,
    }
    return nil
}
```

Si `contracts.ResolvedConnection` n'existe pas encore, l'ajouter dans `contracts/block.go`.

---

### 🟡 IMPORTANT 7 — Connexions absentes du formulaire de blocs sources dans l'UI

Les blocs `source.postgres`, `source.mysql`, `source.mssql` ont un param `connectionRef`
dans leurs structs Go. Dans `NodeConfigPanel.tsx`, ce champ doit afficher un **sélecteur**
avec la liste des connexions disponibles (via `GET /api/v1/connections`).

**Table de correspondance param Go ↔ champ UI :**

| Bloc Go | Param Go | Champ UI attendu |
|---|---|---|
| `source.postgres` | `connectionRef` | `<select>` liste des connexions de type `postgres` |
| `source.mysql` | `connectionRef` | `<select>` liste des connexions de type `mysql` |
| `source.mssql` | `connectionRef` | `<select>` liste des connexions de type `mssql` |
| `target.postgres` | `connectionRef` | `<select>` liste des connexions de type `postgres` |

**Hook à créer `useConnections.ts` :**

```typescript
// web/ui/src/hooks/useConnections.ts
import { useEffect, useState } from 'react'
import { listConnections } from '@/api/client'
import type { Connection } from '@/types/api'

export function useConnections(filterType?: string) {
  const [connections, setConnections] = useState<Connection[]>([])
  useEffect(() => {
    listConnections().then(all =>
      setConnections(filterType ? all.filter(c => c.type === filterType) : all)
    )
  }, [filterType])
  return connections
}
```

**Patch dans `NodeConfigPanel.tsx` pour les blocs avec `connectionRef` :**

```tsx
// Ajouter dans NodeConfigPanel.tsx
import { useConnections } from '@/hooks/useConnections'

// Dans le rendu du paramètre connectionRef :
{paramKey === 'connectionRef' && (
  <ConnectionRefSelect
    blockType={node.type}  // ex: "source.postgres"
    value={params.connectionRef ?? ''}
    onChange={v => updateParam('connectionRef', v)}
  />
)}
```

---

### 🟡 IMPORTANT 8 — Test unitaire du resolver absent

Il n'existe pas de test pour `resolver.go` ni pour `secrets/resolver.go`.

**Template de test `tests/unit/connections/resolver_test.go` :**

```go
package connections_test

import (
    "os"
    "testing"

    "github.com/rinjold/go-etl-studio/internal/connections"
    "github.com/rinjold/go-etl-studio/internal/connections/resolver"
    "github.com/rinjold/go-etl-studio/internal/connections/manager"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestResolve_EnvVar(t *testing.T) {
    t.Setenv("TEST_DB_PASS", "supersecret")

    dir := t.TempDir()
    mgr, err := manager.New(dir, "dev")
    require.NoError(t, err)

    conn := &connections.Connection{
        ID:   "test-conn",
        Name: "Test",
        Type: "postgres",
        Envs: map[string]connections.ConnEnv{
            "dev": {
                Name:      "dev",
                Host:      "localhost",
                Port:      5432,
                Database:  "testdb",
                User:      "testuser",
                SecretRef: "${TEST_DB_PASS}",
            },
        },
    }
    require.NoError(t, mgr.Save(conn))

    rc, err := resolver.Resolve(mgr, "test-conn")
    require.NoError(t, err)
    assert.Equal(t, "postgres", rc.Type)
    assert.Equal(t, "localhost", rc.Host)
    assert.Equal(t, "supersecret", rc.Password)
    assert.Contains(t, rc.DSN, "password=supersecret")
}

func TestResolve_MissingEnvVar(t *testing.T) {
    os.Unsetenv("MISSING_VAR")
    dir := t.TempDir()
    mgr, _ := manager.New(dir, "dev")

    conn := &connections.Connection{
        ID: "bad-conn", Type: "postgres",
        Envs: map[string]connections.ConnEnv{
            "dev": {Name: "dev", Host: "h", Port: 5432, SecretRef: "${MISSING_VAR}"},
        },
    }
    _ = mgr.Save(conn)

    _, err := resolver.Resolve(mgr, "bad-conn")
    assert.ErrorContains(t, err, "MISSING_VAR")
}

func TestResolve_MissingProfile(t *testing.T) {
    dir := t.TempDir()
    mgr, _ := manager.New(dir, "prod") // env = prod

    conn := &connections.Connection{
        ID: "partial-conn", Type: "postgres",
        Envs: map[string]connections.ConnEnv{
            "dev": {Name: "dev", Host: "h", Port: 5432, SecretRef: "password"},
            // profil prod absent
        },
    }
    _ = mgr.Save(conn)

    _, err := resolver.Resolve(mgr, "partial-conn")
    assert.ErrorContains(t, err, "profil 'prod' introuvable")
}

func TestSwitchEnv_Persisted(t *testing.T) {
    dir := t.TempDir()
    mgr, err := manager.New(dir, "dev")
    require.NoError(t, err)
    require.Equal(t, "dev", mgr.ActiveEnv)

    require.NoError(t, mgr.SwitchEnv("prod"))
    assert.Equal(t, "prod", mgr.ActiveEnv)

    // Simuler un redémarrage
    mgr2, err := manager.New(dir, "dev")
    require.NoError(t, err)
    assert.Equal(t, "prod", mgr2.ActiveEnv) // doit survivre au redémarrage
}
```

---

## Plan d'action pour finaliser la Phase 6

### Sprint A — Backend : Test réel + persistance env (1 jour)

- [ ] **Implémenter `pingConnection()`** dans `connection_handler.go` avec support postgres/mysql/mssql/rest
- [ ] **Modifier `SwitchEnv()`** dans `manager.go` pour persister dans `.env-state.json`
- [ ] **Ajouter `loadEnvState()`** appelé dans `New()` pour restaurer l'env au démarrage
- [ ] **Adapter le handler `SwitchEnv`** pour propager l'erreur de persistence
- [ ] **Ajouter la validation** de la whitelist `dev|preprod|prod`
- [ ] **Vérifier `go build ./...`** sans erreur après les modifications

### Sprint B — Tests unitaires connexions (0.5 jour)

- [ ] Créer `tests/unit/connections/resolver_test.go` avec les 4 cas de test ci-dessus
- [ ] Ajouter `tests/unit/connections/manager_test.go` : Save/Load/Delete + persistance XML
- [ ] Ajouter `tests/unit/connections/secrets_test.go` : `${ENV_VAR}`, valeur brute, vault (erreur attendue)
- [ ] Vérifier `go test ./tests/unit/connections/...` vert

### Sprint C — Secrets extensibles (0.5 jour)

- [ ] Créer `internal/connections/secrets/provider.go` avec l'interface `Provider`
- [ ] Créer `EnvProvider` (wrapping le `Resolve` existant)
- [ ] Créer le stub `VaultProvider` avec TODO documenté
- [ ] Modifier `resolver.go` pour accepter un `Provider` optionnel (via injection ou paramètre)

### Sprint D — UI : page de détail + formulaire profils (1.5 jours)

- [ ] Créer `web/ui/src/components/connections/ConnectionEnvForm.tsx` (voir code ci-dessus)
- [ ] Créer `web/ui/src/pages/ConnectionDetailPage.tsx` :
  - Charge la connexion via `GET /api/v1/connections/:connID`
  - Affiche 3 onglets : DEV / PREPROD / PROD
  - Chaque onglet contient un `ConnectionEnvForm` pré-rempli
  - Sauvegarde via `PUT /api/v1/connections/:connID`
- [ ] Ajouter la route `/connections/:connID` dans `App.tsx`
- [ ] Modifier `ConnectionsPage.tsx` :
  - Rendre les cartes connexion cliquables → navigation vers `ConnectionDetailPage`
  - Afficher le nombre de profils configurés sur chaque carte
  - Désactiver le bouton "Tester" si aucun profil configuré pour l'env actif

### Sprint E — UI : sélecteur de connexion dans NodeConfigPanel (0.5 jour)

- [ ] Créer `web/ui/src/hooks/useConnections.ts` (voir code ci-dessus)
- [ ] Modifier `NodeConfigPanel.tsx` : afficher un `<select>` pour `connectionRef` sur les blocs sources/targets DB
- [ ] Le `<select>` filtre les connexions par `type` correspondant au bloc
- [ ] Vérifier que le `connectionRef` sélectionné est bien propagé dans les `params` du nœud → sauvegardé en XML

### Sprint F — Connecter le moteur ETL au manager (0.5 jour)

- [ ] Vérifier `internal/etl/engine/inject_connections.go` — appelle-t-il `resolver.Resolve()` ?
- [ ] Si non : implémenter l'injection correcte (voir code IMPORTANT 6 ci-dessus)
- [ ] Ajouter `contracts.ResolvedConnection` dans `contracts/block.go` si absent
- [ ] Test end-to-end : `source.postgres (connectionRef="test-conn") → target.csv` avec `${DB_PASS}` résolu

---

## Checklist finale Phase 6 — "Definition of Done"

### Backend Go

- [ ] `go build ./...` passe sans erreur
- [ ] `go vet ./...` passe proprement
- [ ] `POST /api/v1/connections/{connID}/test` retourne une **vraie** erreur si la DB est inaccessible
- [ ] `PUT /api/v1/environment` persiste l'env actif — survit à un redémarrage du serveur
- [ ] `manager.New()` restaure l'env actif depuis `.env-state.json` si présent
- [ ] `${ENV_VAR}` est correctement résolu dans les blocs sources/targets à l'exécution
- [ ] `go test ./tests/unit/connections/...` — tous les tests verts

### Pipelines de validation end-to-end connexions

- [ ] Créer une connexion `postgres` via l'UI avec un profil `dev` → sauvegarder → tester → `✅ ok`
- [ ] Changer `secretRef` pour une variable d'env manquante → tester → `❌ MISSING_VAR non définie`
- [ ] Créer un pipeline `source.postgres (connectionRef) → target.csv` → exécuter → résultat correct
- [ ] Basculer env `dev → prod` → relancer → le resolver utilise bien le profil `prod`
- [ ] Redémarrer le serveur → `GET /api/v1/environment` retourne bien `prod` (pas `dev`)

### Frontend React

- [ ] `npm run build` passe sans erreur
- [ ] `ConnectionsPage` affiche le nombre de profils configurés par connexion
- [ ] `ConnectionDetailPage` permet de configurer les 3 profils (dev/preprod/prod) et de sauvegarder
- [ ] Clic sur "Tester" fait un vrai ping — feedback visuel ✅/❌ avec message d'erreur lisible
- [ ] `NodeConfigPanel` affiche un `<select>` pour `connectionRef` sur `source.postgres` / `target.postgres`
- [ ] Switch d'env depuis `ConnectionsPage` est bien réfléchi dans l'env actif affiché

### Déploiement

- [ ] `docker-compose up` → env actif persisté dans le volume Docker (monter `connections/` comme volume)
- [ ] Les variables `${DB_PASSWORD}` sont injectées via `environment:` dans `docker-compose.yml`

---

## Architecture rappel — Flux de résolution d'une connexion

```
Bloc source.postgres
    │  connectionRef = "conn-crm"
    │
    ▼
engine.InjectConnections(blockCtx, manager)
    │
    ├── manager.Get("conn-crm") → Connection{Envs: {dev: {...}, prod: {...}}}
    │
    ├── manager.ActiveEnv = "prod"   ← lu depuis .env-state.json au démarrage
    │
    ├── resolver.ResolveWithEnv(conn, "prod")
    │       │
    │       ├── conn.Envs["prod"] → ConnEnv{Host: "prod-db.internal", SecretRef: "${DB_PROD_PASS}"}
    │       │
    │       └── secrets.Resolve("${DB_PROD_PASS}") → os.Getenv("DB_PROD_PASS") = "s3cr3t"
    │
    └── ResolvedConn{DSN: "host=prod-db.internal port=5432 dbname=crm user=app password=s3cr3t ..."}
            │
            ▼
    blockCtx.Connection.DSN → utilisé par source.postgres pour sql.Open()
```

---

## Fichiers impactés — récapitulatif

| Fichier | Action | Priorité |
|---|---|---|
| `internal/connections/manager/manager.go` | Ajouter `persistEnvState()`, `loadEnvState()`, modifier `SwitchEnv()` | 🔴 BLOQUANT |
| `api/handlers/connection_handler.go` | Implémenter `pingConnection()` + adapter `SwitchEnv` + validation whitelist | 🔴 BLOQUANT |
| `web/ui/src/pages/ConnectionsPage.tsx` | Rendre cartes cliquables, afficher nb profils, désactiver Test si 0 profil | 🔴 BLOQUANT |
| `web/ui/src/pages/ConnectionDetailPage.tsx` | **CRÉER** — page détail avec onglets DEV/PREPROD/PROD | 🔴 BLOQUANT |
| `web/ui/src/components/connections/ConnectionEnvForm.tsx` | **CRÉER** — formulaire de profil d'env | 🔴 BLOQUANT |
| `web/ui/src/App.tsx` | Ajouter route `/connections/:connID` | 🔴 BLOQUANT |
| `internal/connections/secrets/provider.go` | **CRÉER** — interface `Provider` + `EnvProvider` + `VaultProvider` stub | 🟡 IMPORTANT |
| `internal/etl/engine/inject_connections.go` | Vérifier/compléter l'injection via `resolver.Resolve()` | 🟡 IMPORTANT |
| `internal/etl/contracts/block.go` | Ajouter `ResolvedConnection` si absent | 🟡 IMPORTANT |
| `web/ui/src/hooks/useConnections.ts` | **CRÉER** — hook `useConnections(filterType?)` | 🟡 IMPORTANT |
| `web/ui/src/components/editor/NodeConfigPanel.tsx` | Ajouter sélecteur connexion pour blocs avec `connectionRef` | 🟡 IMPORTANT |
| `tests/unit/connections/resolver_test.go` | **CRÉER** — 4 cas de test | 🟡 IMPORTANT |
| `tests/unit/connections/manager_test.go` | **CRÉER** — Save/Load/Delete/persistEnv | 🟡 IMPORTANT |
| `tests/unit/connections/secrets_test.go` | **CRÉER** — EnvVar, plain, vault | 🟡 IMPORTANT |

---

*Document généré automatiquement par analyse du code source — à mettre à jour à chaque sprint.*
