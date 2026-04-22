import { useEffect, useMemo, useRef, useState } from 'react'
import type { Node, Dispatch, SetStateAction } from 'react'
import { useEditorStore } from '@/store/editorStore'
import { useNodeValidation } from '@/hooks/useNodeValidation'
import { X, Trash2, AlertTriangle, CheckCircle2, FolderOpen, RefreshCcw, Eye, Sparkles, Plus, Minus, GripVertical } from 'lucide-react'
import Button from '@/components/ui/Button'
import Badge from '@/components/ui/Badge'

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
    <aside className="w-[26rem] flex-shrink-0 bg-gray-900 border-l border-gray-800 flex flex-col overflow-y-auto">
      <div className="flex items-center justify-between px-4 py-3 border-b border-gray-800 sticky top-0 bg-gray-900 z-10">
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold text-gray-100">{node.data.label as string}</span>
          {meta && <Badge category={(meta as any).category} />}
        </div>
        <button onClick={() => selectNode(null)} className="text-gray-500 hover:text-gray-200">
          <X size={16} />
        </button>
      </div>

      {validation && (
        <div className={`mx-4 mt-3 px-3 py-2 rounded-lg flex items-start gap-2 text-xs ${
          validation.valid
            ? 'bg-green-900/30 border border-green-700 text-green-300'
            : 'bg-red-900/30 border border-red-700 text-red-300'
        }`}>
          {validation.valid
            ? <><CheckCircle2 size={13} className="mt-0.5 flex-shrink-0" /> Bloc correctement configuré</>
            : <><AlertTriangle size={13} className="mt-0.5 flex-shrink-0" /> Champs manquants : <strong className="ml-1">{validation.missing.join(', ')}</strong></>}
        </div>
      )}

      <div className="px-4 py-4 space-y-4 flex-1">
        <Field label="Label">
          <input
            className={inputCls()}
            value={node.data.label as string}
            onChange={e => setNodes(nds => nds.map(n =>
              n.id === nodeId ? { ...n, data: { ...n.data, label: e.target.value } } : n
            ))}
          />
        </Field>

        {needsConnRef(node.data.blockType as string) && (
          <Field label="Réf. connexion (connRef)" required missing={validation?.missing.includes('connRef')}>
            <input
              className={inputCls(validation?.missing.includes('connRef'))}
              value={node.data.connRef as string}
              placeholder="ex: conn-crm-prod"
              onChange={e => setNodes(nds => nds.map(n =>
                n.id === nodeId ? { ...n, data: { ...n.data, connRef: e.target.value } } : n
              ))}
            />
          </Field>
        )}

        {/* ── Data Grid ── */}
        {isDataGrid && (
          <DataGridEditor
            params={params}
            onChange={patch => mergeParams(nodeId, patch, setNodes)}
          />
        )}

        {/* ── Blocs avec paramSchema du catalogue (ex: dedup) ── */}
        {!isDataGrid && hasParamSchema && paramSchema.map((def: any) => (
          <Field key={def.name} label={def.label} required={def.required} help={def.description}>
            {def.type === 'column-multiselect' ? (
              <ColumnMultiSelect
                nodeId={nodeId}
                params={params}
                setNodes={setNodes}
                description={def.description}
              />
            ) : def.type === 'column-select' ? (
              <ColumnSingleSelect
                nodeId={nodeId}
                paramKey={def.name}
                params={params}
                setNodes={setNodes}
              />
            ) : def.type === 'select' ? (
              <select
                className={inputCls()}
                value={params[def.name] ?? def.default ?? ''}
                onChange={e => updateParam(nodeId, def.name, e.target.value, setNodes)}
              >
                <option value="">-- sélectionner --</option>
                {(def.options ?? []).map((opt: string) => (
                  <option key={opt} value={opt}>{opt}</option>
                ))}
              </select>
            ) : (
              <input
                className={inputCls()}
                value={params[def.name] ?? def.default ?? ''}
                placeholder={def.default ?? ''}
                onChange={e => updateParam(nodeId, def.name, e.target.value, setNodes)}
              />
            )}
          </Field>
        ))}

        {/* ── Champs génériques (blocs sans paramSchema) ── */}
        {!isDataGrid && !hasParamSchema && paramFields.map(f => (
          <Field key={f.key} label={f.label} required={f.required} missing={validation?.missing.includes(f.key)} help={f.help}>
            {isCSVBlock && isFilePathField(f.key) ? (
              <FilePickerInput
                value={params[f.key] ?? ''}
                invalid={!!validation?.missing.includes(f.key)}
                onChange={v => updateParam(nodeId, f.key, v, setNodes)}
                placeholder={f.placeholder}
              />
            ) : f.type === 'select' && f.options ? (
              <select
                className={inputCls(validation?.missing.includes(f.key))}
                value={params[f.key] ?? ''}
                onChange={e => updateParam(nodeId, f.key, e.target.value, setNodes)}
              >
                <option value="">-- sélectionner --</option>
                {f.options.map(opt => (
                  <option key={opt.value} value={opt.value}>{opt.label}</option>
                ))}
              </select>
            ) : f.multiline ? (
              <textarea
                className={`${inputCls(validation?.missing.includes(f.key))} resize-none h-24`}
                value={params[f.key] ?? ''}
                placeholder={f.placeholder}
                onChange={e => updateParam(nodeId, f.key, e.target.value, setNodes)}
              />
            ) : (
              <input
                className={inputCls(validation?.missing.includes(f.key))}
                value={params[f.key] ?? ''}
                placeholder={f.placeholder}
                onChange={e => updateParam(nodeId, f.key, e.target.value, setNodes)}
              />
            )}
          </Field>
        ))}

        {/* ── Scan / Preview CSV ── */}
        {isCSVSource && scanInfo && (
          <div className="rounded-xl border border-emerald-700/40 bg-emerald-950/20 overflow-hidden">
            <div className="px-3 py-2 border-b border-emerald-700/30 flex items-center gap-2 text-sm text-emerald-200">
              <Sparkles size={14} /> Détection automatique appliquée
              {scanLoading && <span className="text-[11px] text-emerald-300/80">analyse…</span>}
            </div>
            <div className="p-3 text-xs text-emerald-100 space-y-1">
              <div>Délimiteur : <strong>{scanInfo.delimiter === '\t' ? 'TAB' : scanInfo.delimiter}</strong></div>
              <div>Encodage : <strong>{scanInfo.encoding}</strong></div>
              <div>Colonnes détectées : <strong>{scanInfo.detectedColumns}</strong></div>
              <div>En-tête détectée : <strong>{scanInfo.hasHeader ? 'oui' : 'non'}</strong></div>
              {scanInfo.warnings && scanInfo.warnings.length > 0 && (
                <div className="pt-1 text-amber-300">
                  {scanInfo.warnings.map((w, i) => <div key={i}>• {w}</div>)}
                </div>
              )}
            </div>
          </div>
        )}

        {isCSVSource && (
          <div className="rounded-xl border border-gray-800 bg-gray-950 overflow-hidden">
            <div className="px-3 py-2 border-b border-gray-800 flex items-center justify-between">
              <div className="flex items-center gap-2 text-sm text-gray-200">
                <Eye size={14} /> Prévisualisation CSV
              </div>
              <div className="flex items-center gap-2">
                <button type="button" onClick={() => void scanCSV()}
                  className="text-xs px-2 py-1 rounded-md bg-emerald-900/40 hover:bg-emerald-800/50 text-emerald-200 flex items-center gap-1">
                  <Sparkles size={12} className={scanLoading ? 'animate-spin' : ''} /> Scanner
                </button>
                <button type="button" onClick={() => void runCSVPreview()}
                  className="text-xs px-2 py-1 rounded-md bg-gray-800 hover:bg-gray-700 text-gray-300 flex items-center gap-1">
                  <RefreshCcw size={12} className={previewLoading ? 'animate-spin' : ''} /> Actualiser
                </button>
              </div>
            </div>
            {preview?.error ? (
              <div className="p-3 text-xs text-red-300 bg-red-950/30">{preview.error}</div>
            ) : previewLoading ? (
              <div className="p-3 text-xs text-gray-400">Chargement de la prévisualisation…</div>
            ) : preview && preview.columns.length > 0 ? (
              <div className="overflow-auto max-h-80">
                <table className="min-w-full text-xs">
                  <thead className="sticky top-0 bg-gray-900 z-10">
                    <tr>
                      {preview.columns.map(col => (
                        <th key={col} className="px-3 py-2 text-left text-gray-300 border-b border-gray-800 whitespace-nowrap">{col}</th>
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
          { value: 'eq', label: '= égal' }, { value: 'neq', label: '≠ différent' },
          { value: 'gt', label: '> supérieur' }, { value: 'gte', label: '≥ sup. ou égal' },
          { value: 'lt', label: '< inférieur' }, { value: 'lte', label: '≤ inf. ou égal' },
          { value: 'contains', label: 'contient' }, { value: 'not_contains', label: 'ne contient pas' },
          { value: 'starts_with', label: 'commence par' }, { value: 'ends_with', label: 'termine par' },
          { value: 'is_null', label: 'est null' }, { value: 'is_not_null', label: 'n\'est pas null' },
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
    // transform.dedup n'est plus ici — géré via paramSchema du catalogue
    case 'transform.dummy':
    case 'source.data_grid': return []
    default: return []
  }
}
