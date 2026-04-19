import { useEffect, useCallback, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  addEdge,
  useNodesState,
  useEdgesState,
  type Connection as RFConnection,
  type Edge as RFEdge,
  type Node,
  BackgroundVariant,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'

import { getProject, updateProject, getCatalogue, runProject } from '@/api/client'
import { useEditorStore } from '@/store/editorStore'
import { useAutoSave } from '@/hooks/useAutoSave'
import { useNodeValidation } from '@/hooks/useNodeValidation'
import BlockPalette from '@/components/editor/BlockPalette'
import NodeConfigPanel from '@/components/editor/NodeConfigPanel'
import ETLBlockNode from '@/components/editor/ETLBlockNode'
import DataPreviewPanel from '@/components/editor/DataPreviewPanel'
import Button from '@/components/ui/Button'
import { Save, Play, ArrowLeft, AlertTriangle, Eye } from 'lucide-react'
import type { ETLNode, ETLEdge, Project } from '@/types/api'

const nodeTypes = { etlBlock: ETLBlockNode }

// Style d'un edge normal
const EDGE_STYLE_ACTIVE   = { stroke: '#4f7bff', strokeWidth: 2 }
// Style d'un edge désactivé (pointillé gris)
const EDGE_STYLE_DISABLED = { stroke: '#4b5563', strokeWidth: 2, strokeDasharray: '6 3', opacity: 0.45 }

function toRFNodes(etlNodes: ETLNode[]) {
  return etlNodes.map(n => ({
    id: n.id,
    type: 'etlBlock' as const,
    position: { x: n.posX ?? 100, y: n.posY ?? 100 },
    data: {
      label: n.label ?? n.type,
      blockType: n.type,
      connRef: n.connectionRef ?? '',
      params: Object.fromEntries((n.params ?? []).map(p => [p.name, p.value])),
    },
  }))
}

function toRFEdges(etlEdges: ETLEdge[]) {
  return etlEdges.map((e, i) => ({
    id: `e-${i}-${e.from}-${e.to}`,
    source: e.from,
    target: e.to,
    sourceHandle: e.fromPort ?? null,
    targetHandle: e.toPort ?? null,
    data: { disabled: e.disabled ?? false },
    style: e.disabled ? EDGE_STYLE_DISABLED : EDGE_STYLE_ACTIVE,
    label: e.disabled ? '🚫' : undefined,
    labelStyle: { fill: '#6b7280', fontSize: 10 },
  }))
}

function toETLNodes(rfNodes: Node[]): ETLNode[] {
  return rfNodes.map(n => ({
    id: n.id,
    type: n.data.blockType as string,
    label: n.data.label as string,
    connectionRef: n.data.connRef as string,
    posX: Math.round(n.position.x),
    posY: Math.round(n.position.y),
    params: Object.entries((n.data.params as Record<string, string>) ?? {}).map(([k, v]) => ({ name: k, value: v })),
  }))
}

function toETLEdges(rfEdges: RFEdge[]): ETLEdge[] {
  return rfEdges.map(e => ({
    from: e.source,
    to: e.target,
    fromPort: e.sourceHandle ?? '',
    toPort: e.targetHandle ?? '',
    disabled: (e.data as any)?.disabled ?? false,
  }))
}

export default function EditorPage() {
  const { projectId } = useParams<{ projectId: string }>()
  const navigate = useNavigate()
  const { project, setProject, setCatalogue, setNodes: storeSetNodes, selectedNodeId, selectNode, isDirty, markDirty, markClean } = useEditorStore()

  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])

  // Preview
  const [preview, setPreview] = useState<Record<string, Record<string, any>[]> | null>(null)
  const [showPreview, setShowPreview] = useState(false)

  // Validation globale
  const validationMap = useNodeValidation(nodes)
  const invalidCount = [...validationMap.values()].filter(v => !v.valid).length

  useEffect(() => { storeSetNodes(nodes) }, [nodes])
  useAutoSave(projectId, nodes, edges)

  useEffect(() => {
    if (!projectId) return
    Promise.all([getProject(projectId), getCatalogue()]).then(([p, cat]) => {
      setProject(p)
      setCatalogue(cat)
      setNodes(toRFNodes(p.nodes ?? []))
      setEdges(toRFEdges(p.edges ?? []))
    })
  }, [projectId])

  const onConnect = useCallback(
    (conn: RFConnection) => {
      setEdges(eds => addEdge({ ...conn, style: EDGE_STYLE_ACTIVE, data: { disabled: false } }, eds))
      markDirty()
    },
    [setEdges]
  )

  // Clic droit sur un edge : bascule activé / désactivé
  const onEdgeContextMenu = useCallback(
    (event: React.MouseEvent, edge: RFEdge) => {
      event.preventDefault()
      setEdges(eds =>
        eds.map(e => {
          if (e.id !== edge.id) return e
          const disabled = !((e.data as any)?.disabled ?? false)
          return {
            ...e,
            data: { ...e.data, disabled },
            style: disabled ? EDGE_STYLE_DISABLED : EDGE_STYLE_ACTIVE,
            label: disabled ? '🚫' : undefined,
          }
        })
      )
      markDirty()
    },
    [setEdges]
  )

  const handleSave = async () => {
    if (!project || !projectId) return
    const updated: Project = {
      ...project,
      nodes: toETLNodes(nodes),
      edges: toETLEdges(edges),
    }
    await updateProject(projectId, updated)
    setProject(updated)
    markClean()
  }

  const handleRun = async () => {
    if (invalidCount > 0) {
      const ok = confirm(`${invalidCount} bloc(s) ont des paramètres manquants. Exécuter quand même ?`)
      if (!ok) return
    }
    await handleSave()
    if (!projectId) return
    try {
      const report = await runProject(projectId) as any
      if (report.preview && Object.keys(report.preview).length > 0) {
        setPreview(report.preview)
        setShowPreview(true)
      }
      alert(report.success ? `✅ Succès en ${report.duration}` : `❌ Erreur d'exécution`)
    } catch (e: any) {
      alert(`❌ ${e.response?.data?.error ?? e.message}`)
    }
  }

  const onDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      const blockType = e.dataTransfer.getData('application/goloadit-block')
      if (!blockType) return
      const bounds = (e.currentTarget as HTMLElement).getBoundingClientRect()
      const position = { x: e.clientX - bounds.left - 75, y: e.clientY - bounds.top - 20 }
      const id = `node-${Date.now()}`
      setNodes(nds => [
        ...nds,
        { id, type: 'etlBlock', position, data: { label: blockType.split('.').pop() ?? blockType, blockType, connRef: '', params: {} } },
      ])
      markDirty()
    },
    [setNodes]
  )

  return (
    <div className="flex h-screen">
      <BlockPalette />

      <div className="flex-1 flex flex-col min-w-0">
        {/* Toolbar */}
        <div className="flex items-center gap-2 px-4 py-2 bg-gray-900 border-b border-gray-800 flex-shrink-0">
          <Button size="sm" variant="ghost" onClick={() => navigate('/projects')}>
            <ArrowLeft size={14} /> Projets
          </Button>
          <span className="text-sm font-semibold text-gray-300 ml-2">{project?.name}</span>

          {isDirty && (
            <span className="text-xs text-yellow-400 ml-1 animate-pulse">↻ sauvegarde...</span>
          )}

          {invalidCount > 0 && (
            <span className="flex items-center gap-1 text-xs text-red-400 ml-2">
              <AlertTriangle size={12} />
              {invalidCount} bloc{invalidCount > 1 ? 's' : ''} incomplet{invalidCount > 1 ? 's' : ''}
            </span>
          )}
          {invalidCount === 0 && nodes.length > 0 && (
            <span className="text-xs text-green-400 ml-2">✓ Projet valide</span>
          )}

          <div className="ml-auto flex gap-2 items-center">
            {/* Bouton aperçu (visible après un run) */}
            {preview && (
              <Button
                size="sm"
                variant="ghost"
                onClick={() => setShowPreview(v => !v)}
                title="Afficher / masquer l'aperçu des données"
              >
                <Eye size={14} />
                {showPreview ? 'Masquer aperçu' : 'Aperçu'}
              </Button>
            )}
            <Button size="sm" variant="ghost" onClick={handleSave}><Save size={14} /> Sauvegarder</Button>
            <Button size="sm" onClick={handleRun} disabled={nodes.length === 0}><Play size={14} /> Exécuter</Button>
          </div>
        </div>

        {/* Tooltip d'aide pour les edges */}
        <div className="text-center text-xs text-gray-700 py-0.5 bg-gray-900/50 border-b border-gray-800/50 select-none">
          Clic droit sur un lien pour l’activer / désactiver
        </div>

        {/* Canvas ReactFlow */}
        <div className="flex-1" onDrop={onDrop} onDragOver={e => e.preventDefault()}>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={(changes) => { onNodesChange(changes); markDirty() }}
            onEdgesChange={(changes) => { onEdgesChange(changes); markDirty() }}
            onConnect={onConnect}
            onNodeClick={(_, node) => selectNode(node.id)}
            onPaneClick={() => selectNode(null)}
            onEdgeContextMenu={onEdgeContextMenu}
            nodeTypes={nodeTypes}
            fitView
            deleteKeyCode="Delete"
            style={{ background: '#0a0d14' }}
          >
            <Background variant={BackgroundVariant.Dots} color="#1e2433" />
            <Controls />
            <MiniMap nodeColor="#4f7bff" maskColor="rgba(0,0,0,0.7)" style={{ background: '#0f1117' }} />
          </ReactFlow>
        </div>

        {/* Panel de prévisualisation des données (style Pentaho PDI) */}
        {showPreview && preview && (
          <DataPreviewPanel
            preview={preview}
            nodes={nodes as any}
            onClose={() => setShowPreview(false)}
          />
        )}
      </div>

      {selectedNodeId && <NodeConfigPanel nodeId={selectedNodeId} nodes={nodes} setNodes={setNodes} />}
    </div>
  )
}
