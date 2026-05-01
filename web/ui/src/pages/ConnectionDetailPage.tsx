import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft, CheckCircle, AlertCircle, Loader2 } from 'lucide-react'
import { getConnection, updateConnection } from '@/api/client'
import type { Connection, ConnEnv } from '@/types/api'
import ConnectionEnvForm from '@/components/connections/ConnectionEnvForm'
import Button from '@/components/ui/Button'

const ENVS = ['dev', 'preprod', 'prod'] as const
type EnvName = typeof ENVS[number]

type Status = { type: 'success' | 'error'; msg: string } | null

export default function ConnectionDetailPage() {
  const { connID } = useParams<{ connID: string }>()
  const navigate = useNavigate()

  const [conn, setConn] = useState<Connection | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<EnvName>('dev')
  const [status, setStatus] = useState<Status>(null)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!connID) return
    setLoading(true)
    getConnection(connID)
      .then(setConn)
      .catch(() => setError('Impossible de charger la connexion.'))
      .finally(() => setLoading(false))
  }, [connID])

  const handleSave = async (envName: EnvName, envData: ConnEnv) => {
    if (!conn || !connID) return
    setSaving(true)
    setStatus(null)
    const updated: Connection = {
      ...conn,
      envs: { ...conn.envs, [envName]: { ...envData, name: envName } },
    }
    try {
      const saved = await updateConnection(connID, updated)
      setConn(saved)
      setStatus({ type: 'success', msg: `Profil ${envName.toUpperCase()} enregistré.` })
    } catch {
      setStatus({ type: 'error', msg: 'Erreur lors de la sauvegarde.' })
    } finally {
      setSaving(false)
      setTimeout(() => setStatus(null), 4000)
    }
  }

  if (loading) {
    return (
      <div className="p-8 flex items-center gap-3 text-gray-400">
        <Loader2 size={18} className="animate-spin" />
        Chargement…
      </div>
    )
  }

  if (error || !conn) {
    return (
      <div className="p-8">
        <button onClick={() => navigate('/connections')} className="flex items-center gap-2 text-sm text-gray-400 hover:text-gray-200 mb-4 transition-colors">
          <ArrowLeft size={16} /> Retour
        </button>
        <p className="text-red-400">{error ?? 'Connexion introuvable.'}</p>
      </div>
    )
  }

  return (
    <div className="p-8 max-w-2xl">
      {/* Header */}
      <div className="flex items-center gap-3 mb-6">
        <button
          onClick={() => navigate('/connections')}
          className="flex items-center gap-1.5 text-sm text-gray-400 hover:text-gray-200 transition-colors"
        >
          <ArrowLeft size={16} />
          Connexions
        </button>
        <span className="text-gray-700">/</span>
        <h1 className="text-xl font-bold text-gray-100">{conn.name}</h1>
        <span className="text-xs font-mono text-gray-500 bg-gray-800 px-2 py-0.5 rounded">
          {conn.type}
        </span>
      </div>

      {/* Feedback */}
      {status && (
        <div
          className={`flex items-center gap-2 mb-4 px-4 py-2 rounded-lg text-sm ${
            status.type === 'success'
              ? 'bg-green-900/40 text-green-300 border border-green-800'
              : 'bg-red-900/40 text-red-300 border border-red-800'
          }`}
        >
          {status.type === 'success'
            ? <CheckCircle size={15} />
            : <AlertCircle size={15} />}
          {status.msg}
        </div>
      )}

      {/* Onglets */}
      <div className="flex gap-1 mb-6 bg-gray-900 p-1 rounded-lg w-fit">
        {ENVS.map(env => {
          const hasProfile = Boolean(conn.envs?.[env]?.host)
          return (
            <button
              key={env}
              onClick={() => setActiveTab(env)}
              className={`px-4 py-1.5 rounded-md text-xs font-bold uppercase tracking-wider transition-colors ${
                activeTab === env
                  ? 'bg-brand-600 text-white'
                  : 'text-gray-400 hover:text-gray-200'
              }`}
            >
              {env}
              {hasProfile && (
                <span className="ml-1.5 inline-block w-1.5 h-1.5 rounded-full bg-green-500 align-middle" />
              )}
            </button>
          )
        })}
      </div>

      {/* Formulaire de l'onglet actif */}
      <div className="bg-gray-900 border border-gray-800 rounded-xl p-5">
        <ConnectionEnvForm
          key={activeTab}
          envName={activeTab}
          initial={conn.envs?.[activeTab]}
          onSave={envData => handleSave(activeTab, envData)}
          onCancel={() => navigate('/connections')}
        />
        {saving && (
          <div className="flex items-center gap-2 text-xs text-gray-500 mt-2">
            <Loader2 size={12} className="animate-spin" />
            Enregistrement…
          </div>
        )}
      </div>
    </div>
  )
}
