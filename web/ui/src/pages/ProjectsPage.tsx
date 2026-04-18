import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Plus, Play, Pencil, Trash2, FileCode2 } from 'lucide-react'
import { listProjects, createProject, deleteProject, runProject } from '@/api/client'
import type { Project } from '@/types/api'
import Button from '@/components/ui/Button'
import Modal from '@/components/ui/Modal'

export default function ProjectsPage() {
  const [projects, setProjects] = useState<Project[]>([])
  const [showCreate, setShowCreate] = useState(false)
  const [newName, setNewName] = useState('')
  const [newDesc, setNewDesc] = useState('')
  const [runResult, setRunResult] = useState<string | null>(null)
  const navigate = useNavigate()

  const load = () => listProjects().then(setProjects).catch(console.error)
  useEffect(() => { load() }, [])

  const handleCreate = async () => {
    if (!newName.trim()) return
    const p = await createProject({ name: newName, description: newDesc, nodes: [], edges: [] })
    setShowCreate(false)
    setNewName('')
    setNewDesc('')
    navigate(`/projects/${p.id}/edit`)
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Supprimer ce projet ?')) return
    await deleteProject(id)
    load()
  }

  const handleRun = async (id: string) => {
    try {
      const report = await runProject(id)
      setRunResult(report.success ? `✅ Succès en ${report.duration}` : `❌ Erreur lors de l'exécution`)
    } catch (e: any) {
      setRunResult(`❌ ${e.message}`)
    }
    setTimeout(() => setRunResult(null), 4000)
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-100">Projets ETL</h1>
          <p className="text-gray-500 text-sm mt-1">{projects.length} projet(s)</p>
        </div>
        <Button onClick={() => setShowCreate(true)}>
          <Plus size={16} /> Nouveau projet
        </Button>
      </div>

      {runResult && (
        <div className="mb-4 px-4 py-2 bg-gray-800 rounded-lg text-sm">{runResult}</div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        {projects.map(p => (
          <div key={p.id} className="bg-gray-900 border border-gray-800 rounded-xl p-5 flex flex-col gap-3 hover:border-brand-600 transition-colors">
            <div className="flex items-start justify-between">
              <div>
                <h2 className="font-semibold text-gray-100">{p.name}</h2>
                {p.description && <p className="text-gray-500 text-xs mt-1">{p.description}</p>}
              </div>
              <span className="text-xs text-gray-600 font-mono">v{p.version}</span>
            </div>
            <div className="text-xs text-gray-600">
              {p.nodes?.length ?? 0} bloc(s) • {p.edges?.length ?? 0} lien(s)
            </div>
            <div className="flex gap-2 mt-auto pt-2 border-t border-gray-800">
              <Button size="sm" variant="ghost" onClick={() => navigate(`/projects/${p.id}/edit`)}>
                <Pencil size={13} /> Éditer
              </Button>
              <Button size="sm" variant="primary" onClick={() => handleRun(p.id)}>
                <Play size={13} /> Run
              </Button>
              <Button size="sm" variant="ghost" onClick={() => window.open(`/api/v1/projects/${p.id}/xml`, '_blank')}>
                <FileCode2 size={13} /> XML
              </Button>
              <Button size="sm" variant="danger" onClick={() => handleDelete(p.id)} className="ml-auto">
                <Trash2 size={13} />
              </Button>
            </div>
          </div>
        ))}
      </div>

      {showCreate && (
        <Modal title="Nouveau projet" onClose={() => setShowCreate(false)}>
          <div className="space-y-4">
            <div>
              <label className="block text-xs text-gray-400 mb-1">Nom *</label>
              <input
                autoFocus
                className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-gray-100 focus:outline-none focus:border-brand-500"
                value={newName}
                onChange={e => setNewName(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && handleCreate()}
                placeholder="Mon projet ETL"
              />
            </div>
            <div>
              <label className="block text-xs text-gray-400 mb-1">Description</label>
              <input
                className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-gray-100 focus:outline-none focus:border-brand-500"
                value={newDesc}
                onChange={e => setNewDesc(e.target.value)}
                placeholder="Description optionnelle"
              />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <Button variant="ghost" onClick={() => setShowCreate(false)}>Annuler</Button>
              <Button onClick={handleCreate} disabled={!newName.trim()}>Créer</Button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}
