import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Plus, Trash2, Zap } from 'lucide-react'
import { listConnections, createConnection, deleteConnection, testConnection, switchEnvironment, getEnvironment } from '@/api/client'
import type { Connection } from '@/types/api'
import Button from '@/components/ui/Button'
import Modal from '@/components/ui/Modal'

const DB_TYPES = ['postgres', 'mysql', 'mssql', 'rest']
const ENVS = ['dev', 'preprod', 'prod']

/** Compte le nombre de profils d'env non vides (host présent) parmi les 3 standard. */
function countConfiguredProfiles(c: Connection): number {
  return ENVS.filter(env => Boolean(c.envs?.[env]?.host?.trim())).length
}

/** Retourne true si la connexion possède un profil configuré pour l'env donné. */
function hasProfileForEnv(c: Connection, env: string): boolean {
  return Boolean(c.envs?.[env]?.host?.trim())
}

export default function ConnectionsPage() {
  const navigate = useNavigate()
  const [connections, setConnections] = useState<Connection[]>([])
  const [activeEnv, setActiveEnvState] = useState('dev')
  const [showCreate, setShowCreate] = useState(false)
  const [testMsg, setTestMsg] = useState<string | null>(null)

  const [form, setForm] = useState({ name: '', type: 'postgres' })

  const load = () => {
    listConnections().then(setConnections).catch(console.error)
    getEnvironment().then((r: { activeEnv: string }) => { setActiveEnvState(r.activeEnv) })
  }
  useEffect(() => { load() }, [])

  const handleSwitchEnv = async (env: string) => {
    await switchEnvironment(env)
    setActiveEnvState(env)
  }

  const handleCreate = async () => {
    if (!form.name.trim()) return
    await createConnection({ name: form.name, type: form.type, envs: {} })
    setShowCreate(false)
    setForm({ name: '', type: 'postgres' })
    load()
  }

  const handleTest = async (e: React.MouseEvent, id: string) => {
    e.stopPropagation()
    try {
      const r = await testConnection(id) as { type: string; host: string; db: string; env: string }
      setTestMsg(`✅ ${r.type} — ${r.host}/${r.db} [${r.env}]`)
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { error?: string } }; message?: string }
      setTestMsg(`❌ ${axiosErr.response?.data?.error ?? axiosErr.message ?? 'Erreur inconnue'}`)
    }
    setTimeout(() => setTestMsg(null), 5000)
  }

  const handleDelete = async (e: React.MouseEvent, id: string) => {
    e.stopPropagation()
    await deleteConnection(id)
    load()
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-100">Connexions</h1>
          <p className="text-gray-500 text-sm mt-1">Réutilisables entre tous les projets</p>
        </div>
        <Button onClick={() => setShowCreate(true)}><Plus size={16} /> Nouvelle connexion</Button>
      </div>

      {/* Switch d'environnement global */}
      <div className="mb-6 flex items-center gap-3">
        <span className="text-sm text-gray-400">Environnement actif :</span>
        {ENVS.map(env => (
          <button
            key={env}
            onClick={() => handleSwitchEnv(env)}
            className={`px-3 py-1 rounded-lg text-xs font-bold transition-colors ${
              activeEnv === env ? 'bg-brand-600 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }`}
          >
            {env.toUpperCase()}
          </button>
        ))}
      </div>

      {testMsg && <div className="mb-4 px-4 py-2 bg-gray-800 rounded-lg text-sm">{testMsg}</div>}

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        {connections.map(c => {
          const profileCount = countConfiguredProfiles(c)
          const canTest = hasProfileForEnv(c, activeEnv)
          return (
            <div
              key={c.id}
              onClick={() => navigate(`/connections/${c.id}`)}
              className="bg-gray-900 border border-gray-800 rounded-xl p-5 cursor-pointer hover:border-gray-600 transition-colors"
            >
              <div className="flex items-center justify-between mb-2">
                <h2 className="font-semibold text-gray-100">{c.name}</h2>
                <span className="text-xs font-mono text-gray-500 bg-gray-800 px-2 py-0.5 rounded">{c.type}</span>
              </div>
              <div className="text-xs text-gray-600 mb-4">
                {profileCount === 0
                  ? 'Aucun profil configuré'
                  : `${profileCount} profil${profileCount > 1 ? 's' : ''} configuré${profileCount > 1 ? 's' : ''}`
                }
              </div>
              <div className="flex gap-2">
                <Button
                  size="sm"
                  variant="ghost"
                  disabled={!canTest}
                  onClick={e => handleTest(e, c.id)}
                  title={canTest ? undefined : `Aucun profil pour l'env ${activeEnv}`}
                >
                  <Zap size={13} /> Tester
                </Button>
                <Button
                  size="sm"
                  variant="danger"
                  onClick={e => handleDelete(e, c.id)}
                  className="ml-auto"
                >
                  <Trash2 size={13} />
                </Button>
              </div>
            </div>
          )
        })}
      </div>

      {showCreate && (
        <Modal title="Nouvelle connexion" onClose={() => setShowCreate(false)}>
          <div className="space-y-4">
            <div>
              <label className="block text-xs text-gray-400 mb-1">Nom *</label>
              <input
                autoFocus
                className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-gray-100 focus:outline-none focus:border-brand-500"
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                placeholder="CRM PostgreSQL"
              />
            </div>
            <div>
              <label className="block text-xs text-gray-400 mb-1">Type</label>
              <select
                className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-gray-100"
                value={form.type}
                onChange={e => setForm(f => ({ ...f, type: e.target.value }))}
              >
                {DB_TYPES.map(t => <option key={t} value={t}>{t}</option>)}
              </select>
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <Button variant="ghost" onClick={() => setShowCreate(false)}>Annuler</Button>
              <Button onClick={handleCreate} disabled={!form.name.trim()}>Créer</Button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}
