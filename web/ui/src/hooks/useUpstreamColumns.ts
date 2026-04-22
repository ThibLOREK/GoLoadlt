import { useMemo } from 'react'
import type { Node } from '@xyflow/react'
import { useEditorStore } from '@/store/editorStore'

/**
 * Remonte la chaîne de blocs en amont de `nodeId` jusqu'à trouver
 * une source qui expose ses colonnes (params.headers ou params.columns).
 * Fonctionne avec n'importe quelle profondeur de pipeline.
 */
export function useUpstreamColumns(nodeId: string, propNodes?: Node[]): string[] {
  const { edges, nodes: storeNodes } = useEditorStore()

  return useMemo(() => {
    // On fusionne les nodes du store et les nodes passées en prop
    // (les props sont plus récentes côté React, le store peut être légèrement en retard)
    const nodesMap = new Map<string, Node>()
    storeNodes.forEach(n => nodesMap.set(n.id, n))
    if (propNodes) propNodes.forEach(n => nodesMap.set(n.id, n))

    const visited = new Set<string>()

    function walk(id: string): string[] {
      if (visited.has(id)) return []
      visited.add(id)

      const node = nodesMap.get(id)
      if (!node) return []

      const p = (node.data.params ?? {}) as Record<string, string>

      // source.csv → params.headers
      if (p.headers && p.headers.trim()) {
        return p.headers.split(',').map(s => s.trim()).filter(Boolean)
      }
      // source.data_grid → params.columns
      if (p.columns && p.columns.trim()) {
        return p.columns.split(',').map(s => s.trim()).filter(Boolean)
      }
      // source SQL → on ne connaît pas les colonnes statiquement, on s'arrête
      const bt = node.data.blockType as string
      if (['source.postgres', 'source.mysql', 'source.mssql'].includes(bt)) {
        return []
      }

      // Bloc transform ou autre : on remonte vers le nœud source connecté
      const parentEdge = edges.find(e => e.target === id)
      if (!parentEdge) return []
      return walk(parentEdge.source)
    }

    // On part du nœud courant et on remonte
    const parentEdge = edges.find(e => e.target === nodeId)
    if (!parentEdge) return []
    return walk(parentEdge.source)
  }, [edges, storeNodes, propNodes, nodeId])
}
