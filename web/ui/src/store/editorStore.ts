import { create } from 'zustand'
import type { Node, Edge } from '@xyflow/react'
import type { Project, BlockMeta } from '@/types/api'

interface EditorStore {
  project: Project | null
  nodes: Node[]
  edges: Edge[]
  catalogue: BlockMeta[]
  selectedNodeId: string | null
  isDirty: boolean

  setProject: (p: Project) => void
  setNodes: (nodes: Node[]) => void
  setEdges: (edges: Edge[]) => void
  setCatalogue: (c: BlockMeta[]) => void
  selectNode: (id: string | null) => void
  markDirty: () => void
  markClean: () => void
  updateNodeParam: (nodeId: string, key: string, value: string) => void
}

export const useEditorStore = create<EditorStore>((set, get) => ({
  project: null,
  nodes: [],
  edges: [],
  catalogue: [],
  selectedNodeId: null,
  isDirty: false,

  setProject: (p) => set({ project: p }),
  setNodes:   (nodes) => set({ nodes }),
  setEdges:   (edges) => set({ edges }),
  setCatalogue: (c) => set({ catalogue: c }),
  selectNode: (id) => set({ selectedNodeId: id }),
  markDirty:  () => set({ isDirty: true }),
  markClean:  () => set({ isDirty: false }),

  updateNodeParam: (nodeId, key, value) => {
    const nodes = get().nodes.map(n => {
      if (n.id !== nodeId) return n
      const params: Record<string, string> = { ...(n.data.params as Record<string, string> || {}) }
      params[key] = value
      return { ...n, data: { ...n.data, params } }
    })
    set({ nodes, isDirty: true })
  },
}))
