import { useMemo } from 'react'
import type { Node } from '@xyflow/react'

// Champs obligatoires par type de bloc.
const REQUIRED_PARAMS: Record<string, string[]> = {
  'source.csv':         ['path'],
  'source.postgres':    ['query'],
  'source.mysql':       ['query'],
  'source.mssql':       ['query'],
  'target.csv':         ['path'],
  'target.postgres':    ['table'],
  'transform.filter':   ['condition'],
  'transform.select':   ['columns'],
  'transform.cast':     ['column', 'targetType'],
  'transform.add_column': ['name', 'expression'],
  'transform.join':     ['leftKey', 'rightKey'],
  'transform.split':    ['conditions'],
  'transform.aggregate': ['groupBy', 'aggregations'],
  'transform.sort':     ['columns'],
  'transform.dedup':    ['keys'],
  'transform.pivot':    ['groupBy', 'pivotColumn', 'valueColumn'],
  'transform.unpivot':  ['columns', 'keyName', 'valueName'],
}

export interface NodeValidation {
  valid: boolean
  missing: string[]
}

/**
 * useNodeValidation retourne un Map<nodeId, NodeValidation> pour tous les nodes.
 * Un node source/target qui nécessite une connexion (connRef) est aussi vérifié.
 */
export function useNodeValidation(nodes: Node[]): Map<string, NodeValidation> {
  return useMemo(() => {
    const map = new Map<string, NodeValidation>()
    for (const node of nodes) {
      const blockType = node.data.blockType as string
      const params = (node.data.params as Record<string, string>) ?? {}
      const required = REQUIRED_PARAMS[blockType] ?? []
      const missing: string[] = []

      for (const key of required) {
        if (!params[key]?.trim()) missing.push(key)
      }

      // Les sources DB et targets DB nécessitent une connexion référencée.
      const needsConn = [
        'source.postgres', 'source.mysql', 'source.mssql',
        'target.postgres',
      ].includes(blockType)
      if (needsConn && !(node.data.connRef as string)?.trim()) {
        missing.push('connRef')
      }

      map.set(node.id, { valid: missing.length === 0, missing })
    }
    return map
  }, [nodes])
}
