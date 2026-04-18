import { useEffect, useRef } from 'react'
import type { Node, Edge } from '@xyflow/react'
import { updateProject } from '@/api/client'
import type { Project, ETLNode, ETLEdge } from '@/types/api'
import { useEditorStore } from '@/store/editorStore'

const AUTO_SAVE_DELAY_MS = 2000

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

function toETLEdges(rfEdges: Edge[]): ETLEdge[] {
  return rfEdges.map(e => ({
    from: e.source,
    to: e.target,
    fromPort: e.sourceHandle ?? '',
    toPort: e.targetHandle ?? '',
  }))
}

/**
 * useAutoSave déclenche une sauvegarde silencieuse 2s après la dernière modification.
 * N'envoie la requête que si le projet est "dirty".
 */
export function useAutoSave(projectId: string | undefined, nodes: Node[], edges: Edge[]) {
  const { project, setProject, markClean, isDirty } = useEditorStore()
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    if (!isDirty || !projectId || !project) return

    if (timerRef.current) clearTimeout(timerRef.current)

    timerRef.current = setTimeout(async () => {
      try {
        const updated: Project = {
          ...project,
          nodes: toETLNodes(nodes),
          edges: toETLEdges(edges),
        }
        await updateProject(projectId, updated)
        setProject(updated)
        markClean()
      } catch (e) {
        console.error('Auto-save échoué :', e)
      }
    }, AUTO_SAVE_DELAY_MS)

    return () => {
      if (timerRef.current) clearTimeout(timerRef.current)
    }
  }, [isDirty, nodes, edges, projectId])
}
