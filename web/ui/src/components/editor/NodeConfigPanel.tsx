import { useEffect, useMemo, useRef, useState } from 'react'
import type { Node, Dispatch, SetStateAction } from 'react'
import { useEditorStore } from '@/store/editorStore'
import { useNodeValidation } from '@/hooks/useNodeValidation'
import { X, Trash2, AlertTriangle, CheckCircle2, FolderOpen, RefreshCcw, Eye, Sparkles, Plus, Minus, GripVertical } from 'lucide-react'
import Button from '@/components/ui/Button'
import Badge from '@/components/ui/Badge'
import ConnectionRefSelect from '@/components/connections/ConnectionRefSelect'

interface Props {
  nodeId: string
  nodes: Node[]
  setNodes: Dispatch<SetStateAction<Node[]>>
}

function updateParam(nodeId: string, key: string, value: string, setNodes: Dispatch<SetStateAction<Node[]>>) {
  setNodes(nds => nds.map(n => {
    if (n.id !== nodeId) return n
    const params = { ...(n.data.params as Record<string, string> ?? {}), [key]: value }
    return { ...n, data: { ...n.data, params } }
  }))
}

function mergeParams(nodeId: string, patch: Record<string, string>, setNodes: Dispatch<SetStateAction<Node[]>>) {
  setNodes(nds => nds.map(n => {
    if (n.id !== nodeId) return n
    const current = (n.data.params as Record<string, string> ?? {})
    const params = { ...current, ...patch }
    return { ...n, data: { ...n.data, params } }
  }))
}

// ─── Types pour Data Grid ────────────────────────────────────────────────────

const DATA_TYPES = ['string', 'int', 'float', 'bool', 'date', 'datetime'] as const
type DataType = typeof DATA_TYPES[number]

interface GridColumn {
  name: string
  type: DataType
}

interface GridState {
  columns: GridColumn[]
  rows: string[][]
}

function parseGridParams(params: Record<string, string>): GridState {
  let columns: GridColumn[] = []
  let rows: string[][] = []
  try {
    const names = params.columns ? params.columns.split(',').map(s => s.trim()).filter(Boolean) : []
    const types = params.types ? params.types.split(',').map(s => s.trim()) : []
    columns = names.map((name, i) => ({ name, type: (types[i] as DataType) ?? 'string' }))
  } catch { columns = [] }
  try {
    rows = params.rows ? JSON.parse(params.rows) : []
    if (!Array.isArray(rows)) rows = []
  } catch { rows = [] }
  return { columns, rows }
}

function serializeGrid(grid: GridState): Record<string, string> {
  return {
    columns: grid.columns.map(c => c.name).join(','),
    types: grid.columns.map(c => c.type).join(','),
    rows: JSON.stringify(grid.rows),
  }
}

// ─── Récupère les colonnes du nœud source connecté en amont ──────────────────

function useUpstreamColumns(nodeId: string): string[] {
  const { edges, nodes } = useEditorStore()
  return useMemo(() => {
    // Edge entrant dont la cible est nodeId
    const incoming = edges.find(e => e.target === nodeId)
    if (!incoming) return []
    const sourceNode = nodes.find(n => n.id === incoming.source)
    if (!sourceNode) return []
    // Colonnes détectées via params.headers (source.csv) ou params.columns (data_grid)
    const p = (sourceNode.data.params ?? {}) as Record<string, string>
    const raw = p.headers ?? p.columns ?? ''
    if (raw) return raw.split(',').map(s => s.trim()).filter(Boolean)
    return []
  }, [edges, nodes, nodeId])
}

// ─── Lit les clés indexées key_0, key_1… depuis params ───────────────────────

function readIndexedKeys(params: Record<string, string>): string[] {
  const keys: string[] = []
  for (let i = 0; ; i++) {
    const v = params[`key_${i}`]
    if (v === undefined || v === '') break
    keys.push(v)
  }
  return keys
}

// Écrit key_0…key_N dans les params, efface les anciens surplus et la clé legacy 'keys'
function writeIndexedKeys(
  nodeId: string,
  selected: string[],
  prevCount: number,
  setNodes: Dispatch<SetStateAction<Node[]>>
) {
  setNodes(nds => nds.map(n => {
    if (n.id !== nodeId) return n
    const params = { ...(n.data.params as Record<string, string> ?? {}) }
    // Supprimer anciens
    for (let i = 0; i < Math.max(prevCount, selected.length + 1); i++) delete params[`key_${i}`]
    delete params['keys'] // supprimer l'ancienne clé CSV si elle existait
    // Écrire nouveaux
    selected.forEach((v, i) => { params[`key_${i}`] = v })
    return { ...n, data: { ...n.data, params } }
  }))
}

// ─── Composant liste déroulante multi-sélection pour colonnes ────────────────

function ColumnMultiSelect({
  nodeId,
  params,
  setNodes,
  description,
}: {
  nodeId: string
  params: Record<string, string>
  setNodes: Dispatch<SetStateAction<Node[]>>
  description?: string
}) {
  const columns = useUpstreamColumns(nodeId)
  const selected = readIndexedKeys(params)
  const prevCount = selected.length

  const addKey = (col: string) => {
    if (selected.includes(col)) return
    writeIndexedKeys(nodeId, [...selected, col], prevCount, setNodes)
  }

  const removeKey = (col: string) => {
    writeIndexedKeys(nodeId, selected.filter(k => k !== col), prevCount, setNodes)
  }

  const moveKey = (from: number, to: number) => {
    const next = [...selected]
    const [item] = next.splice(from, 1)
    next.splice(to, 0, item)
    writeIndexedKeys(nodeId, next, prevCount, setNodes)
  }

  const available = columns.filter(c => !selected.includes(c))

  return (
    <div className="space-y-2">
      {/* Liste des clés sélectionnées */}
      {selected.length > 0 && (
        <div className="space-y-1">
          {selected.map((col, i) => (
            <div
              key={col}
              className="flex items-center gap-2 bg-gray-800 border border-gray-700 rounded-lg px-2 py-1.5 group"
            >
              <GripVertical size={13} className="text-gray-600 flex-shrink-0 cursor-grab" />
              <span className="flex-1 text-xs text-gray-100 font-mono">{col}</span>
              <span className="text-[10px] text-gray-600 mr-1">clé {i + 1}</span>
              {i > 0 && (
                <button
                  type="button"
                  onClick={() => moveKey(i, i - 1)}
                  className="text-gray-600 hover:text-gray-300 text-[10px] px-1"
                  title="Monter"
                >▲</button>
              )}
              {i < selected.length - 1 && (
                <button
                  type="button"
                  onClick={() => moveKey(i, i + 1)}
                  className="text-gray-600 hover:text-gray-300 text-[10px] px-1"
                  title="Descendre"
                >▼</button>
              )}
              <button
                type="button"
                onClick={() => removeKey(col)}
                className="text-gray-600 hover:text-red-400 transition-colors"
                title="Retirer"
              >
                <Minus size={13} />
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Liste déroulante pour ajouter une colonne */}
      {available.length > 0 ? (
        <select
          className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-gray-400 focus:outline-none focus:ring-1 focus:ring-brand-500 cursor-pointer"
          value=""
          onChange={e => { if (e.target.value) addKey(e.target.value) }}
        >
          <option value="">+ Ajouter une colonne clé…</option>
          {available.map(col => (
            <option key={col} value={col}>{col}</option>
          ))}
        </select>
      ) : columns.length === 0 ? (
        <div className="text-xs text-gray-500 border border-dashed border-gray-700 rounded-lg px-3 py-2">
          Connecte un bloc source pour voir les colonnes disponibles.
        </div>
      ) : (
        <div className="text-xs text-gray-500 border border-dashed border-gray-700 rounded-lg px-3 py-2">
          Toutes les colonnes sont sélectionnées.
        </div>
      )}

      {/* Champ texte libre de secours si aucune colonne détectée */}
      {columns.length === 0 && (
        <input
          className={inputCls()}
          value={selected.join(', ')}
          placeholder="id, email, date (saisie manuelle)"
          onChange={e => {
            const vals = e.target.value.split(',').map(s => s.trim()).filter(Boolean)
            writeIndexedKeys(nodeId, vals, prevCount, setNodes)
          }}
        />
      )}

      {description && <p className="text-[11px] text-gray-500">{description}</p>}
    </div>
  )
}

// ─── Composant éditeur de grille ─────────────────────────────────────────────

function DataGridEditor({
  params,
  onChange,
}: {
  params: Record<string, string>
  onChange: (patch: Record<string, string>) => void
}) {
  const [grid, setGrid] = useState<GridState>(() => parseGridParams(params))

  useEffect(() => {
    onChange(serializeGrid(grid))
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [grid])

  const updateGrid = (next: GridState) => setGrid(next)

  const addColumn = () => {
    const next: GridState = {
      columns: [...grid.columns, { name: `col_${grid.columns.length + 1}`, type: 'string' }],
      rows: grid.rows.map(r => [...r, '']),
    }
    updateGrid(next)
  }

  const removeColumn = (colIdx: number) => {
    if (grid.columns.length <= 1) return
    const next: GridState = {
      columns: grid.columns.filter((_, i) => i !== colIdx),
      rows: grid.rows.map(r => r.filter((_, i) => i !== colIdx)),
    }
    updateGrid(next)
  }

  const renameColumn = (colIdx: number, name: string) => {
    const columns = grid.columns.map((c, i) => i === colIdx ? { ...c, name } : c)
    updateGrid({ ...grid, columns })
  }

  const changeType = (colIdx: number, type: DataType) => {
    const columns = grid.columns.map((c, i) => i === colIdx ? { ...c, type } : c)
    updateGrid({ ...grid, columns })
  }

  const addRow = () => {
    const next: GridState = {
      ...grid,
      rows: [...grid.rows, Array(grid.columns.length).fill('')],
    }
    updateGrid(next)
  }

  const removeRow = (rowIdx: number) => {
    updateGrid({ ...grid, rows: grid.rows.filter((_, i) => i !== rowIdx) })
  }

  const setCell = (rowIdx: number, colIdx: number, value: string) => {
    const rows = grid.rows.map((r, ri) =>
      ri === rowIdx ? r.map((c, ci) => ci === colIdx ? value : c) : r
    )
    updateGrid({ ...grid, rows })
  }

  const cellCls = 'bg-gray-800 border border-gray-700 px-2 py-1 text-xs text-gray-100 focus:outline-none focus:ring-1 focus:ring-brand-500 w-full min-w-[80px]'
  const headerCellCls = 'bg-gray-900 border border-gray-700 px-1 py-1 text-xs text-gray-100 focus:outline-none focus:ring-1 focus:ring-brand-500 w-full min-w-[80px] font-semibold'

  return (
    <div className="rounded-xl border border-gray-700 bg-gray-950 overflow-hidden">
      <div className="px-3 py-2 border-b border-gray-800 flex items-center justify-between bg-gray-900">
        <span className="text-xs font-semibold text-gray-300">Éditeur de données</span>
        <div className="flex items-center gap-1">
          <button type="button" onClick={addColumn}
            className="flex items-center gap-1 text-[11px] px-2 py-1 rounded bg-gray-700 hover:bg-gray-600 text-gray-200 transition-colors"
            title="Ajouter une colonne">
            <Plus size={11} /> Colonne
          </button>
          <button type="button" onClick={addRow}
            className="flex items-center gap-1 text-[11px] px-2 py-1 rounded bg-gray-700 hover:bg-gray-600 text-gray-200 transition-colors"
            title="Ajouter une ligne">
            <Plus size={11} /> Ligne
          </button>
        </div>
      </div>

      {grid.columns.length === 0 ? (
        <div className="p-4 text-xs text-gray-500 text-center">
          Aucune colonne — clique sur <strong>+ Colonne</strong> pour commencer
        </div>
      ) : (
        <div className="overflow-auto max-h-96">
          <table className="border-collapse text-xs w-full">
            <thead className="sticky top-0 z-10 bg-gray-900">
              <tr>
                <th className="w-6 border border-gray-700 bg-gray-900" />
                {grid.columns.map((col, ci) => (
                  <th key={ci} className="border border-gray-700 bg-gray-900 p-0 min-w-[90px]">
                    <div className="flex items-center gap-0.5 px-1">
                      <input className={headerCellCls} value={col.name}
                        onChange={e => renameColumn(ci, e.target.value)}
                        placeholder={`col_${ci + 1}`} title="Nom de la colonne" />
                      <button type="button" onClick={() => removeColumn(ci)}
                        className="flex-shrink-0 text-gray-600 hover:text-red-400 transition-colors p-0.5"
                        title="Supprimer cette colonne">
                        <Minus size={11} />
                      </button>
                    </div>
                  </th>
                ))}
              </tr>
              <tr>
                <th className="w-6 border border-gray-700 bg-gray-800 text-[10px] text-gray-500 px-1">type</th>
                {grid.columns.map((col, ci) => (
                  <th key={ci} className="border border-gray-700 bg-gray-800 p-0">
                    <select
                      className="w-full bg-gray-800 text-[11px] text-brand-400 px-1 py-1 focus:outline-none focus:ring-1 focus:ring-brand-500 border-0 cursor-pointer"
                      value={col.type} onChange={e => changeType(ci, e.target.value as DataType)}>
                      {DATA_TYPES.map(t => <option key={t} value={t}>{t}</option>)}
                    </select>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {grid.rows.length === 0 ? (
                <tr>
                  <td colSpan={grid.columns.length + 1} className="text-center text-gray-600 py-3 text-xs border border-gray-800">
                    Aucune ligne — clique sur <strong>+ Ligne</strong>
                  </td>
                </tr>
              ) : grid.rows.map((row, ri) => (
                <tr key={ri} className="odd:bg-gray-950 even:bg-gray-900/40 group">
                  <td className="border border-gray-800 text-center text-[10px] text-gray-600 w-6 px-1 group-hover:text-red-400">
                    <button type="button" onClick={() => removeRow(ri)}
                      className="w-full text-center hover:text-red-400 transition-colors"
                      title="Supprimer cette ligne">{ri + 1}</button>
                  </td>
                  {row.map((cell, ci) => (
                    <td key={ci} className="border border-gray-800 p-0">
                      <input className={cellCls} value={cell}
                        onChange={e => setCell(ri, ci, e.target.value)} placeholder="" />
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <div className="px-3 py-1.5 border-t border-gray-800 bg-gray-900 text-[10px] text-gray-600">
        {grid.columns.length} colonne{grid.columns.length !== 1 ? 's' : ''} · {grid.rows.length} ligne{grid.rows.length !== 1 ? 's' : ''}
      </div>
    </div>
  )
}

// ─── Panel principal ──────────────────────────────────────────────────────────

export default function NodeConfigPanel({ nodeId, nodes, setNodes }: Props) {
  const { catalogue, selectNode } = useEditorStore()
  const node = nodes.find(n => n.id === nodeId)
  const [preview, setPreview] = useState<{ columns: string[]; rows: Record<string, string>[]; error?: string } | null>(null)
  const [previewLoading, setPreviewLoading] = useState(false)
  const [scanLoading, setScanLoading] = useState(false)
  const [scanInfo, setScanInfo] = useState<{ delimiter: string; encoding: string; hasHeader: boolean; detectedColumns: number; warnings?: string[] } | null>(null)
  if (!node) return null

  const meta = catalogue.find((b: any) => b.type === node.data.blockType)
  const params = (node.data.params ?? {}) as Record<string, string>

  const validationMap = useNodeValidation(nodes)
  const validation = validationMap.get(nodeId)

  const projectId = useMemo(() => {
    const path = window.location.pathname
    const match = path.match(/projects\/([^/]+)/)
    return match?.[1] ?? 'local'
  }, [])

  const handleDelete = () => {
    setNodes(nds => nds.filter(n => n.id !== nodeId))
    selectNode(null)
  }

  // Blocs avec un paramSchema fourni par le catalogue (ex: transform.dedup)
  const paramSchema: any[] = (meta as any)?.paramSchema ?? []
  const hasParamSchema = paramSchema.length > 0

  const paramFields = getParamFields(node.data.blockType as string)
  const isFilePathField = (key: string) => key === 'path'
  const isCSVBlock = ['source.csv', 'target.csv'].includes(node.data.blockType as string)
  const isCSVSource = (node.data.blockType as string) === 'source.csv'
  const isDataGrid = (node.data.blockType as string) === 'source.data_grid'

  const scanCSV = async () => {
    if (!params.path) return
    setScanLoading(true)
    try {
      const query = new URLSearchParams({ path: params.path })
      const res = await fetch(`/api/v1/projects/${projectId}/csv-scan?${query.toString()}`)
      const text = await res.text()
      let data: any = null
      try { data = JSON.parse(text) } catch { throw new Error(text || 'Réponse invalide du serveur') }
      if (!res.ok || !data.success) throw new Error(data?.error ?? 'Erreur de scan CSV')
      const suggested = data.scan?.suggestedParams ?? {}
      mergeParams(nodeId, suggested, setNodes)
      setScanInfo({
        delimiter: data.scan?.delimiter ?? ',',
        encoding: data.scan?.encoding ?? 'utf-8',
        hasHeader: !!data.scan?.hasHeader,
        detectedColumns: data.scan?.detectedColumns ?? 0,
        warnings: data.scan?.warnings ?? [],
      })
    } catch (e: any) {
      setScanInfo({
        delimiter: params.delimiter ?? ',',
        encoding: params.encoding ?? 'utf-8',
        hasHeader: (params.has_header ?? 'true') !== 'false',
        detectedColumns: 0,
        warnings: [e?.message ?? 'Erreur de scan automatique'],
      })
    } finally {
      setScanLoading(false)
    }
  }

  const runCSVPreview = async () => {
    if (!params.path) {
      setPreview({ columns: [], rows: [], error: 'Le chemin du fichier est requis pour la prévisualisation.' })
      return
    }
    setPreviewLoading(true)
    try {
      const query = new URLSearchParams({
        path: params.path ?? '',
        delimiter: params.delimiter ?? ',',
        encoding: params.encoding ?? 'utf-8',
        newline: params.newline ?? 'auto',
        has_header: params.has_header ?? 'true',
        headers: params.headers ?? '',
        lazy_quotes: params.lazy_quotes ?? 'true',
        trim_leading_space: params.trim_leading_space ?? 'true',
        skip_empty_lines: params.skip_empty_lines ?? 'true',
        fields_per_record: params.fields_per_record ?? '-1',
        limit: '20',
      })
      const res = await fetch(`/api/v1/projects/${projectId}/csv-preview?${query.toString()}`)
      const text = await res.text()
      let data: any = null
      try { data = JSON.parse(text) } catch { throw new Error(text || 'Réponse invalide du serveur') }
      if (!res.ok || !data.success) {
        setPreview({ columns: [], rows: [], error: data.error ?? 'Erreur de prévisualisation' })
      } else {
        setPreview({ columns: data.columns ?? [], rows: data.rows ?? [] })
      }
    } catch (e: any) {
      setPreview({ columns: [], rows: [], error: e?.message ?? 'Erreur réseau' })
    } finally {
      setPreviewLoading(false)
    }
  }

  useEffect(() => {
    if (isCSVSource && params.path) void scanCSV()
    else setScanInfo(null)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [nodeId, isCSVSource, params.path])

  useEffect(() => {
    if (isCSVSource && params.path) void runCSVPreview()
    else setPreview(null)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [nodeId, isCSVSource, params.path, params.delimiter, params.encoding, params.newline,
      params.has_header, params.headers, params.lazy_quotes, params.trim_leading_space, params.fields_per_record])

  return (
    <aside className="w-[26rem] flex-shrink-0 bg-gray-900 border-l border-gray-800 flex flex-col overflow-hidden">
      {/* Header */}
      <div className="px-4 py-3 border-b border-gray-800 flex items-center justify-between">
        <div className="flex items-center gap-2 min-w-0">
          <span className="text-sm font-semibold text-gray-100 truncate">{(meta as any)?.label ?? node.data.blockType as string}</span>
          {validation && (
            validation.valid
              ? <CheckCircle2 size={14} className="text-green-500 flex-shrink-0" />
              : <AlertTriangle size={14} className="text-yellow-500 flex-shrink-0" />
          )}
        </div>
        <button
          onClick={() => selectNode(null)}
          className="text-gray-500 hover:text-gray-200 transition-colors flex-shrink-0"
          aria-label="Fermer le panneau"
        >
          <X size={16} />
        </button>
      </div>

      {/* Corps */}
      <div className="flex-1 overflow-y-auto px-4 py-4 space-y-4">
        {validation && !validation.valid && (
          <div className="bg-yellow-900/30 border border-yellow-700/50 rounded-lg px-3 py-2">
            <div className="flex items-center gap-1.5 text-yellow-400 text-xs font-medium mb-1">
              <AlertTriangle size={12} /> Configuration incomplète
            </div>
            <ul className="space-y-0.5">
              {validation.errors.map((err: string, i: number) => (
                <li key={i} className="text-xs text-yellow-300/80">• {err}</li>
              ))}
            </ul>
          </div>
        )}

        {/* Connexion DB pour les blocs SQL — sélecteur ConnectionRefSelect */}
        {needsConnRef(node.data.blockType as string) && (
          <Field label="Connexion" required>
            <ConnectionRefSelect
              blockType={node.data.blockType as string}
              value={params.connRef ?? ''}
              onChange={v => updateParam(nodeId, 'connRef', v, setNodes)}
            />
          </Field>
        )}

        {/* Data Grid editor */}
        {isDataGrid && (
          <DataGridEditor
            params={params}
            onChange={patch => mergeParams(nodeId, patch, setNodes)}
          />
        )}

        {/* Champs via paramSchema (ex: transform.dedup) */}
        {hasParamSchema && paramSchema.map((field: any) => (
          <Field key={field.name} label={field.label} required={field.required}
            missing={field.required && !params[field.name]}
            help={field.description}>
            {field.type === 'column-multiselect' ? (
              <ColumnMultiSelect
                nodeId={nodeId}
                params={params}
                setNodes={setNodes}
                description={field.description}
              />
            ) : field.type === 'column-select' ? (
              <ColumnSingleSelect
                nodeId={nodeId}
                paramKey={field.name}
                params={params}
                setNodes={setNodes}
              />
            ) : field.type === 'select' ? (
              <select
                className={inputCls()}
                value={params[field.name] ?? field.default ?? ''}
                onChange={e => updateParam(nodeId, field.name, e.target.value, setNodes)}
              >
                <option value="">-- sélectionner --</option>
                {(field.options ?? []).map((opt: string) => (
                  <option key={opt} value={opt}>{opt}</option>
                ))}
              </select>
            ) : (
              <input
                className={inputCls(field.required && !params[field.name])}
                value={params[field.name] ?? ''}
                placeholder={field.placeholder ?? ''}
                onChange={e => updateParam(nodeId, field.name, e.target.value, setNodes)}
              />
            )}
          </Field>
        ))}

        {/* Champs statiques via getParamFields */}
        {!hasParamSchema && paramFields.map(field => (
          <Field key={field.key} label={field.label} required={field.required}
            missing={field.required && !params[field.key]}
            help={field.help}>
            {isFilePathField(field.key) && isCSVBlock ? (
              <FilePickerInput
                value={params[field.key] ?? ''}
                invalid={!!field.required && !params[field.key]}
                onChange={v => updateParam(nodeId, field.key, v, setNodes)}
                placeholder={field.placeholder}
              />
            ) : field.type === 'select' ? (
              <select
                className={inputCls()}
                value={params[field.key] ?? ''}
                onChange={e => updateParam(nodeId, field.key, e.target.value, setNodes)}
              >
                {(field.options ?? []).map(opt => (
                  <option key={opt.value} value={opt.value}>{opt.label}</option>
                ))}
              </select>
            ) : field.multiline ? (
              <textarea
                className={`${inputCls(!!field.required && !params[field.key])} min-h-[80px] resize-y font-mono text-xs`}
                value={params[field.key] ?? ''}
                placeholder={field.placeholder}
                onChange={e => updateParam(nodeId, field.key, e.target.value, setNodes)}
              />
            ) : (
              <input
                className={inputCls(!!field.required && !params[field.key])}
                value={params[field.key] ?? ''}
                placeholder={field.placeholder}
                onChange={e => updateParam(nodeId, field.key, e.target.value, setNodes)}
              />
            )}
          </Field>
        ))}

        {/* Scan CSV automatique pour source.csv */}
        {isCSVSource && (
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={() => void scanCSV()}
                disabled={!params.path || scanLoading}
                className="flex items-center gap-1.5 text-xs px-3 py-1.5 bg-gray-700 hover:bg-gray-600 disabled:opacity-40 disabled:cursor-not-allowed border border-gray-600 rounded-lg text-gray-200 transition-colors"
              >
                {scanLoading
                  ? <RefreshCcw size={13} className="animate-spin" />
                  : <Sparkles size={13} />}
                Analyser automatiquement
              </button>
              <button
                type="button"
                onClick={() => void runCSVPreview()}
                disabled={!params.path || previewLoading}
                className="flex items-center gap-1.5 text-xs px-3 py-1.5 bg-gray-700 hover:bg-gray-600 disabled:opacity-40 disabled:cursor-not-allowed border border-gray-600 rounded-lg text-gray-200 transition-colors"
              >
                {previewLoading
                  ? <RefreshCcw size={13} className="animate-spin" />
                  : <Eye size={13} />}
                Prévisualiser
              </button>
            </div>

            {scanInfo && (
              <div className="bg-gray-800/60 border border-gray-700 rounded-lg px-3 py-2 text-xs space-y-1">
                <div className="flex flex-wrap gap-x-4 gap-y-1 text-gray-300">
                  <span>Délimiteur : <code className="text-brand-400">{scanInfo.delimiter === '\t' ? 'TAB' : scanInfo.delimiter}</code></span>
                  <span>Encodage : <code className="text-brand-400">{scanInfo.encoding}</code></span>
                  <span>En-tête : <code className="text-brand-400">{scanInfo.hasHeader ? 'oui' : 'non'}</code></span>
                  <span>Colonnes : <code className="text-brand-400">{scanInfo.detectedColumns}</code></span>
                </div>
                {scanInfo.warnings && scanInfo.warnings.length > 0 && (
                  <div className="mt-1 space-y-0.5">
                    {scanInfo.warnings.map((w, i) => (
                      <div key={i} className="flex items-center gap-1 text-yellow-400">
                        <AlertTriangle size={10} /> {w}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}

            {preview ? (
              preview.error ? (
                <div className="bg-red-900/20 border border-red-700/40 rounded-lg px-3 py-2 text-xs text-red-300">
                  <AlertTriangle size={12} className="inline mr-1" />{preview.error}
                </div>
              ) : preview.columns.length > 0 ? (
                <div className="rounded-lg border border-gray-700 overflow-auto max-h-56">
                  <table className="text-xs w-full">
                    <thead className="sticky top-0 bg-gray-800">
                      <tr>
                        {preview.columns.map(col => (
                          <th key={col} className="px-3 py-2 text-left font-medium text-gray-300 whitespace-nowrap border-b border-gray-700">{col}</th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {preview.rows.map((row, idx) => (
                        <tr key={idx} className="odd:bg-gray-950 even:bg-gray-900/50">
                          {preview.columns.map(col => (
                            <td key={col} className="px-3 py-2 border-b border-gray-800 text-gray-200 whitespace-nowrap">{String(row[col] ?? '')}</td>
                          ))}
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="p-3 text-xs text-gray-500">Renseigne le chemin du fichier pour lancer l'analyse automatique.</div>
              )
            ) : (
              <div className="p-3 text-xs text-gray-500">Renseigne le chemin du fichier pour lancer l'analyse automatique.</div>
            )}
          </div>
        )}

        <div className="pt-2 border-t border-gray-800">
          <p className="text-xs text-gray-600 font-mono">{node.data.blockType as string}</p>
          {meta && <p className="text-xs text-gray-600 mt-1">{(meta as any).description}</p>}
        </div>
      </div>

      <div className="px-4 py-3 border-t border-gray-800">
        <Button variant="danger" size="sm" className="w-full justify-center" onClick={handleDelete}>
          <Trash2 size={13} /> Supprimer le bloc
        </Button>
      </div>
    </aside>
  )
}

// ─── ColumnSingleSelect (pour type column-select) ─────────────────────────────

function ColumnSingleSelect({
  nodeId, paramKey, params, setNodes,
}: {
  nodeId: string
  paramKey: string
  params: Record<string, string>
  setNodes: Dispatch<SetStateAction<Node[]>>
}) {
  const columns = useUpstreamColumns(nodeId)
  return (
    <select
      className={inputCls()}
      value={params[paramKey] ?? ''}
      onChange={e => updateParam(nodeId, paramKey, e.target.value, setNodes)}
    >
      <option value="">-- sélectionner une colonne --</option>
      {columns.map(col => <option key={col} value={col}>{col}</option>)}
    </select>
  )
}

// ─── FilePickerInput ──────────────────────────────────────────────────────────

function FilePickerInput({
  value, invalid, onChange, placeholder
}: {
  value: string
  invalid: boolean
  onChange: (v: string) => void
  placeholder?: string
}) {
  const fileRef = useRef<HTMLInputElement>(null)
  return (
    <div className="flex gap-2">
      <input
        className={`${inputCls(invalid)} flex-1`}
        value={value}
        placeholder={placeholder ?? '/data/input.csv'}
        onChange={e => onChange(e.target.value)}
      />
      <button type="button" title="Parcourir"
        onClick={() => fileRef.current?.click()}
        className="flex-shrink-0 px-2 py-1.5 bg-gray-700 hover:bg-gray-600 border border-gray-600 rounded-lg text-gray-300 transition-colors">
        <FolderOpen size={15} />
      </button>
      <input ref={fileRef} type="file" accept=".csv,.txt" className="hidden"
        onChange={e => {
          const file = e.target.files?.[0]
          if (!file) return
          onChange((file as any).path || file.name)
        }}
      />
    </div>
  )
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

function inputCls(invalid = false) {
  return `w-full bg-gray-800 border rounded-lg px-3 py-2 text-sm text-gray-100 placeholder-gray-600
    focus:outline-none focus:ring-1 transition-colors ${
    invalid
      ? 'border-red-500 focus:ring-red-500'
      : 'border-gray-700 focus:ring-brand-500 focus:border-brand-500'
  }`
}

function Field({
  label, required, missing, help, children,
}: {
  label: string
  required?: boolean
  missing?: boolean
  help?: string
  children: React.ReactNode
}) {
  return (
    <div>
      <label className="flex items-center gap-1 text-xs text-gray-400 mb-1">
        {label}
        {required && <span className="text-red-400">*</span>}
        {missing && <span className="text-red-400 text-xs ml-auto">requis</span>}
      </label>
      {children}
      {help && <p className="mt-1 text-[11px] text-gray-500">{help}</p>}
    </div>
  )
}

function needsConnRef(blockType: string) {
  return ['source.postgres', 'source.mysql', 'source.mssql', 'target.postgres'].includes(blockType)
}

interface SelectOption { value: string; label: string }
interface ParamField {
  key: string
  label: string
  placeholder?: string
  multiline?: boolean
  required?: boolean
  type?: 'text' | 'select'
  options?: SelectOption[]
  help?: string
}

function getParamFields(blockType: string): ParamField[] {
  switch (blockType) {
    case 'source.csv': return [
      { key: 'path', label: 'Chemin fichier', placeholder: '/data/input.csv', required: true, help: 'Chemin accessible par le backend Go sur la machine serveur.' },
      { key: 'encoding', label: 'Encodage', required: true, type: 'select', help: 'Détecté automatiquement, modifiable si nécessaire.',
        options: [
          { value: 'utf-8', label: 'UTF-8' },
          { value: 'windows-1252', label: 'Windows-1252' },
          { value: 'iso-8859-1', label: 'ISO-8859-1 / Latin-1' },
          { value: 'utf-16le', label: 'UTF-16 LE' },
          { value: 'utf-16be', label: 'UTF-16 BE' },
        ] },
      { key: 'delimiter', label: 'Délimiteur', help: 'Détecté automatiquement.', type: 'select',
        options: [
          { value: ',', label: 'Virgule (,)' },
          { value: ';', label: 'Point-virgule (;)' },
          { value: '|', label: 'Pipe (|)' },
          { value: '\t', label: 'Tabulation (TAB)' },
        ] },
      { key: 'newline', label: 'Retour à la ligne', type: 'select', help: 'Détecté automatiquement.',
        options: [
          { value: 'auto', label: 'Auto (LF / CRLF)' },
          { value: 'cr', label: 'CR uniquement (ancien Mac)' },
        ] },
      { key: 'has_header', label: 'Présence d\u2019en-tête', required: true, type: 'select', help: 'Détectée automatiquement, mais corrigeable.',
        options: [
          { value: 'true', label: 'Oui, la première ligne contient les colonnes' },
          { value: 'false', label: 'Non, le fichier commence directement par les données' },
        ] },
      { key: 'headers', label: 'Colonnes détectées / manuelles', placeholder: 'column_1,column_2,column_3', help: 'Si aucune en-tête n\u2019est détectée, des noms column_N sont proposés automatiquement.' },
      { key: 'skip_empty_lines', label: 'Ignorer lignes vides', type: 'select',
        options: [{ value: 'true', label: 'Oui' }, { value: 'false', label: 'Non' }] },
      { key: 'trim_leading_space', label: 'Supprimer espaces de début', type: 'select',
        options: [{ value: 'true', label: 'Oui' }, { value: 'false', label: 'Non' }] },
      { key: 'lazy_quotes', label: 'Tolérance guillemets imparfaits', type: 'select', help: 'Pratique pour des CSV sales issus d\u2019exports métiers.',
        options: [{ value: 'true', label: 'Oui' }, { value: 'false', label: 'Non' }] },
      { key: 'fields_per_record', label: 'Nb champs attendu', placeholder: '-1', help: '-1 = variable ; sinon impose un nombre fixe de colonnes.' },
    ]
    case 'source.postgres':
    case 'source.mysql':
    case 'source.mssql': return [
      { key: 'query', label: 'Requête SQL', placeholder: 'SELECT * FROM table_name', multiline: true, required: true },
    ]
    case 'target.csv': return [
      { key: 'path', label: 'Chemin fichier', placeholder: '/data/output.csv', required: true },
      { key: 'delimiter', label: 'Délimiteur', placeholder: ',' },
      { key: 'append', label: 'Mode append', placeholder: 'false' },
    ]
    case 'target.postgres': return [
      { key: 'table', label: 'Table cible', placeholder: 'schema.table', required: true },
    ]
    case 'transform.filter_advanced': return [
      { key: 'field', label: 'Champ à évaluer', placeholder: 'amount', required: true },
      { key: 'operator', label: 'Opérateur', required: true, type: 'select',
        options: [
          { value: 'eq', label: '= égal' }, { value: 'neq', label: '\u2260 différent' },
          { value: 'gt', label: '> supérieur' }, { value: 'gte', label: '\u2265 sup. ou égal' },
          { value: 'lt', label: '< inférieur' }, { value: 'lte', label: '\u2264 inf. ou égal' },
          { value: 'contains', label: 'contient' }, { value: 'not_contains', label: 'ne contient pas' },
          { value: 'starts_with', label: 'commence par' }, { value: 'ends_with', label: 'termine par' },
          { value: 'is_null', label: 'est null' }, { value: 'is_not_null', label: "n'est pas null" },
          { value: 'is_true', label: 'est vrai (bool)' }, { value: 'is_false', label: 'est faux (bool)' },
        ] },
      { key: 'value', label: 'Valeur de comparaison', placeholder: '100' },
      { key: 'value_type', label: 'Type de valeur', type: 'select',
        options: [
          { value: 'string', label: 'Texte (string)' },
          { value: 'number', label: 'Nombre (number)' },
          { value: 'bool', label: 'Booléen (bool)' },
        ] },
    ]
    case 'transform.filter': return [
      { key: 'condition', label: 'Condition', placeholder: 'amount > 100', required: true },
    ]
    case 'transform.select': return [
      { key: 'columns', label: 'Colonnes (virgule)', placeholder: 'id, name, amount', required: true },
    ]
    case 'transform.cast': return [
      { key: 'column', label: 'Colonne', placeholder: 'price', required: true },
      { key: 'targetType', label: 'Type cible', placeholder: 'float | int | string | bool', required: true },
    ]
    case 'transform.add_column': return [
      { key: 'name', label: 'Nom colonne', placeholder: 'tax', required: true },
      { key: 'expression', label: 'Expression', placeholder: 'amount * 0.20', required: true },
    ]
    case 'transform.join': return [
      { key: 'leftKey', label: 'Clé gauche', placeholder: 'user_id', required: true },
      { key: 'rightKey', label: 'Clé droite', placeholder: 'id', required: true },
      { key: 'type', label: 'Type de join', placeholder: 'inner | left | right | full' },
    ]
    case 'transform.split': return [
      { key: 'conditions', label: 'Conditions (virgule)', placeholder: 'amount > 1000, amount > 500', required: true },
    ]
    case 'transform.aggregate': return [
      { key: 'groupBy', label: 'Group By (virgule)', placeholder: 'region, category', required: true },
      { key: 'aggregations', label: 'Agrégations', placeholder: 'SUM(amount), COUNT(id)', required: true },
    ]
    case 'transform.sort': return [
      { key: 'columns', label: 'Colonnes (virgule)', placeholder: 'date, amount', required: true },
      { key: 'order', label: 'Ordre', placeholder: 'asc | desc' },
    ]
    case 'transform.pivot': return [
      { key: 'groupBy', label: 'Group By', placeholder: 'region', required: true },
      { key: 'pivotColumn', label: 'Colonne pivot', placeholder: 'product', required: true },
      { key: 'valueColumn', label: 'Colonne valeur', placeholder: 'amount', required: true },
      { key: 'aggregation', label: 'Agrégation', placeholder: 'SUM | COUNT | AVG | MIN | MAX' },
    ]
    case 'transform.unpivot': return [
      { key: 'columns', label: 'Colonnes à dépivoter', placeholder: 'jan, fev, mar', required: true },
      { key: 'keyName', label: 'Nom clé', placeholder: 'mois', required: true },
      { key: 'valueName', label: 'Nom valeur', placeholder: 'montant', required: true },
    ]
    // ─── Blocs bonus Sprint E ────────────────────────────────────────────────
    case 'transform.regex': return [
      { key: 'column', label: 'Colonne source', placeholder: 'email', required: true },
      { key: 'pattern', label: 'Expression régulière', placeholder: '([\\w.]+)@', required: true },
      { key: 'mode', label: 'Mode', type: 'select', help: 'extract: capture le 1er groupe | replace: remplace | match: filtre les lignes',
        options: [
          { value: 'extract', label: 'extract — capture le 1er groupe capturant' },
          { value: 'replace', label: 'replace — remplace les occurrences' },
          { value: 'match', label: 'match — ne garde que les lignes qui matchent' },
        ] },
      { key: 'replace', label: 'Valeur de remplacement', placeholder: 'REDACTED', help: 'Uniquement pour le mode replace.' },
      { key: 'output', label: 'Colonne de sortie', placeholder: 'email_extracted', help: 'Mode extract uniquement. Défaut : colonne + "_extracted".' },
    ]
    case 'transform.find_replace': return [
      { key: 'column', label: 'Colonne', placeholder: 'status', required: true },
      { key: 'find', label: 'Valeur à chercher', placeholder: 'N/A', required: true },
      { key: 'replace', label: 'Valeur de remplacement', placeholder: '' },
      { key: 'mode', label: 'Mode de correspondance', type: 'select',
        options: [
          { value: 'exact', label: 'exact — égalité stricte' },
          { value: 'contains', label: 'contains — sous-chaîne' },
          { value: 'regex', label: 'regex — expression régulière' },
        ] },
    ]
    case 'transform.sampling': return [
      { key: 'mode', label: 'Mode', type: 'select', required: true,
        options: [
          { value: 'first', label: 'first — N premières lignes' },
          { value: 'percent', label: 'percent — % aléatoire' },
          { value: 'every', label: 'every — 1 ligne sur N' },
        ] },
      { key: 'value', label: 'Valeur (N ou %)', placeholder: '100', required: true, help: 'Ex: 100 pour les 100 premières, 10.5 pour 10.5%, 5 pour 1 ligne sur 5.' },
    ]
    case 'transform.text_to_columns': return [
      { key: 'column', label: 'Colonne à découper', placeholder: 'full_name', required: true },
      { key: 'delimiter', label: 'Délimiteur', placeholder: ',', help: 'Défaut : virgule.' },
      { key: 'prefix', label: 'Préfixe colonnes générées', placeholder: 'part_', help: 'Défaut : nom_colonne + "_".' },
      { key: 'maxSplit', label: 'Nb max de colonnes', placeholder: '0', help: '0 = illimité.' },
    ]
    case 'transform.auto_field': return [
      // Aucun paramètre — détection automatique sur toutes les colonnes
    ]
    case 'transform.append_fields': return [
      // Pas de paramètre utilisateur — les colonnes en conflit sont préfixées 'right_' automatiquement
    ]
    case 'transform.data_cleansing': return [
      { key: 'columns', label: 'Colonnes ciblées', placeholder: 'name, email', help: 'Laisser vide pour appliquer à toutes les colonnes.' },
      { key: 'trim', label: 'Supprimer espaces début/fin', type: 'select',
        options: [{ value: 'true', label: 'Oui' }, { value: 'false', label: 'Non' }] },
      { key: 'toLower', label: 'Mettre en minuscules', type: 'select',
        options: [{ value: 'true', label: 'Oui' }, { value: 'false', label: 'Non' }] },
      { key: 'toUpper', label: 'Mettre en majuscules', type: 'select',
        options: [{ value: 'true', label: 'Oui' }, { value: 'false', label: 'Non' }] },
      { key: 'removeSpecial', label: 'Supprimer caractères spéciaux', type: 'select',
        options: [{ value: 'true', label: 'Oui' }, { value: 'false', label: 'Non' }] },
      { key: 'nullifyEmpty', label: 'Nullifier les chaînes vides', type: 'select',
        options: [{ value: 'true', label: 'Oui' }, { value: 'false', label: 'Non' }] },
    ]
    case 'transform.datetime': return [
      { key: 'column', label: 'Colonne date/heure', placeholder: 'created_at', required: true },
      { key: 'mode', label: 'Mode', type: 'select', required: true,
        options: [
          { value: 'parse', label: 'parse / format — reformater la date' },
          { value: 'add', label: 'add — ajouter une durée' },
          { value: 'extract', label: 'extract — extraire une composante' },
        ] },
      { key: 'inputFormat', label: 'Format d\u2019entrée Go', placeholder: '2006-01-02', help: 'Format Go (référence : 2006-01-02 15:04:05). Défaut : 2006-01-02.' },
      { key: 'outputFormat', label: 'Format de sortie Go', placeholder: '2006-01-02T15:04:05', help: 'Pour les modes parse et add.' },
      { key: 'addUnit', label: 'Unité à ajouter', type: 'select', help: 'Mode add uniquement.',
        options: [
          { value: 'days', label: 'Jours' },
          { value: 'hours', label: 'Heures' },
          { value: 'minutes', label: 'Minutes' },
        ] },
      { key: 'addValue', label: 'Valeur à ajouter', placeholder: '7', help: 'Nombre entier ou décimal. Mode add uniquement.' },
      { key: 'extract', label: 'Composante à extraire', type: 'select', help: 'Mode extract uniquement.',
        options: [
          { value: 'year', label: 'Année' },
          { value: 'month', label: 'Mois (numéro)' },
          { value: 'day', label: 'Jour du mois' },
          { value: 'weekday', label: 'Jour de la semaine (texte)' },
          { value: 'hour', label: 'Heure' },
          { value: 'minute', label: 'Minute' },
          { value: 'second', label: 'Seconde' },
        ] },
      { key: 'output', label: 'Colonne de sortie', placeholder: 'created_year', help: 'Défaut : remplace la colonne source.' },
    ]
    // transform.dedup n'est plus ici — géré via paramSchema du catalogue
    case 'transform.dummy':
    case 'source.data_grid': return []
    default: return []
  }
}
