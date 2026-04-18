import { useEffect, useCallback } from 'react'
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
import Button from '@/components/ui/Button'
import { Save, Play, ArrowLeft, AlertTriangle } from 'lucide-react'
import type { ETLNode, ETLEdge, Project } from '@/types/api'

const nodeTypes = { etlBlock: ETLBlockNode }

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
    style: { stroke: '#4f7bff', strokeWidth: 2 },
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

export default function EditorPage() {
  const { projectId } = useParams<{ projectId: string }>()
  const navigate = useNavigate()
  const { project, setProject, setCatalogue, setNodes: storeSetNodes, selectedNodeId, selectNode, isDirty, markDirty, markClean } = useEditorStore()

  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])

  // Validation globale
  const validationMap = useNodeValidation(nodes)
  const invalidCount = [...validationMap.values()].filter(v => !v.valid).length

  // Sync nodes dans le store pour ETLBlockNode (qui y accède via useEditorStore)
  useEffect(() => { storeSetNodes(nodes) }, [nodes])

  // Auto-save
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
      setEdges(eds => addEdge({ ...conn, style: { stroke: '#4f7bff', strokeWidth: 2 } }, eds))
      markDirty()
    },
    [setEdges]
  )

  const handleSave = async () => {
    if (!project || !projectId) return
    const updated: Project = {
      ...project,
      nodes: toETLNodes(nodes),
      edges: edges.map(e => ({ from: e.source, to: e.target, fromPort: e.sourceHandle ?? '', toPort: e.targetHandle ?? '' })),
    }
    await updateProject(projectId, updated)
    setProject(updated)
    markClean()
  }

  const handleRun = async () => {
    if (invalidCount > 0) {
      const ok = confirm(`${invalidCount} bloc(s) ont des paramètres manquants. Exécuter quand même ?`)
      if (!ok) return
    }
    await handleSave()
    if (!projectId) return
    try {
      const report = await runProject(projectId) as any
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

      <div className="flex-1 flex flex-col">
        {/* Toolbar */}
        <div className="flex items-center gap-2 px-4 py-2 bg-gray-900 border-b border-gray-800">
          <Button size="sm" variant="ghost" onClick={() => navigate('/projects')}>
            <ArrowLeft size={14} /> Projets
          </Button>
          <span className="text-sm font-semibold text-gray-300 ml-2">{project?.name}</span>

          {/* Badge dirty */}
          {isDirty && (
            <span className="text-xs text-yellow-400 ml-1 animate-pulse">↻ sauvegarde...</span>
          )}

          {/* Badge validation globale */}
          {invalidCount > 0 && (
            <span className="flex items-center gap-1 text-xs text-red-400 ml-2">
              <AlertTriangle size={12} />
              {invalidCount} bloc{invalidCount > 1 ? 's' : ''} incomplet{invalidCount > 1 ? 's' : ''}
            </span>
          )}
          {invalidCount === 0 && nodes.length > 0 && (
            <span className="text-xs text-green-400 ml-2">✓ Projet valide</span>
          )}

          <div className="ml-auto flex gap-2">
            <Button size="sm" variant="ghost" onClick={handleSave}><Save size={14} /> Sauvegarder</Button>
            <Button size="sm" onClick={handleRun} disabled={nodes.length === 0}><Play size={14} /> Exécuter</Button>
          </div>
        </div>

        <div className="flex-1" onDrop={onDrop} onDragOver={e => e.preventDefault()}>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={(changes) => { onNodesChange(changes); markDirty() }}
            onEdgesChange={(changes) => { onEdgesChange(changes); markDirty() }}
            onConnect={onConnect}
            onNodeClick={(_, node) => selectNode(node.id)}
            onPaneClick={() => selectNode(null)}
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
      </div>

      {selectedNodeId && <NodeConfigPanel nodeId={selectedNodeId} nodes={nodes} setNodes={setNodes} />}
    </div>
  )
}
