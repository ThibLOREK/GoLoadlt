import { useEffect, useState } from 'react'
import { Plus, Trash2, Zap } from 'lucide-react'
import { listConnections, createConnection, deleteConnection, testConnection, switchEnv, getEnv } from '@/api/client'
import type { Connection } from '@/types/api'
import Button from '@/components/ui/Button'
import Modal from '@/components/ui/Modal'

const DB_TYPES = ['postgres', 'mysql', 'mssql', 'rest']
const ENVS = ['dev', 'preprod', 'prod']

export default function ConnectionsPage() {
  const [connections, setConnections] = useState<Connection[]>([])
  const [activeEnv, setActiveEnvState] = useState('dev')
  const [showCreate, setShowCreate] = useState(false)
  const [testMsg, setTestMsg] = useState<string | null>(null)

  const [form, setForm] = useState({ name: '', type: 'postgres' })

  const load = () => {
    listConnections().then(setConnections).catch(console.error)
    getEnv().then(r => { setActiveEnvState(r.activeEnv); localStorage.setItem('activeEnv', r.activeEnv) })
  }
  useEffect(() => { load() }, [])

  const handleSwitchEnv = async (env: string) => {
    await switchEnv(env)
    setActiveEnvState(env)
    localStorage.setItem('activeEnv', env)
  }

  const handleCreate = async () => {
    if (!form.name.trim()) return
    await createConnection({ name: form.name, type: form.type, envs: {} })
    setShowCreate(false)
    setForm({ name: '', type: 'postgres' })
    load()
  }

  const handleTest = async (id: string) => {
    try {
      const r = await testConnection(id) as any
      setTestMsg(`✅ ${r.type} — ${r.host}/${r.db} [${r.env}]`)
    } catch (e: any) {
      setTestMsg(`❌ ${e.response?.data?.error ?? e.message}`)
    }
    setTimeout(() => setTestMsg(null), 5000)
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
        {connections.map(c => (
          <div key={c.id} className="bg-gray-900 border border-gray-800 rounded-xl p-5">
            <div className="flex items-center justify-between mb-2">
              <h2 className="font-semibold text-gray-100">{c.name}</h2>
              <span className="text-xs font-mono text-gray-500 bg-gray-800 px-2 py-0.5 rounded">{c.type}</span>
            </div>
            <div className="text-xs text-gray-600 mb-4">
              Profils : {Object.keys(c.envs ?? {}).join(', ') || 'aucun'}
            </div>
            <div className="flex gap-2">
              <Button size="sm" variant="ghost" onClick={() => handleTest(c.id)}>
                <Zap size={13} /> Tester
              </Button>
              <Button size="sm" variant="danger" onClick={async () => { await deleteConnection(c.id); load() }} className="ml-auto">
                <Trash2 size={13} />
              </Button>
            </div>
          </div>
        ))}
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
