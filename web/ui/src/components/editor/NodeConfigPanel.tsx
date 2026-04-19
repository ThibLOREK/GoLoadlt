import { useEffect, useMemo, useRef, useState } from 'react'
import type { Node, Dispatch, SetStateAction } from 'react'
import { useEditorStore } from '@/store/editorStore'
import { useNodeValidation } from '@/hooks/useNodeValidation'
import { X, Trash2, AlertTriangle, CheckCircle2, FolderOpen, RefreshCcw, Eye } from 'lucide-react'
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

export default function NodeConfigPanel({ nodeId, nodes, setNodes }: Props) {
  const { catalogue, selectNode } = useEditorStore()
  const node = nodes.find(n => n.id === nodeId)
  const [preview, setPreview] = useState<{ columns: string[]; rows: Record<string, string>[]; error?: string } | null>(null)
  const [previewLoading, setPreviewLoading] = useState(false)
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

  const paramFields = getParamFields(node.data.blockType as string)
  const isFilePathField = (key: string) => key === 'path'
  const isCSVBlock = ['source.csv', 'target.csv'].includes(node.data.blockType as string)
  const isCSVSource = (node.data.blockType as string) === 'source.csv'

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
      const data = await res.json()
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
    if (isCSVSource && params.path) {
      void runCSVPreview()
    } else {
      setPreview(null)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    nodeId,
    isCSVSource,
    params.path,
    params.delimiter,
    params.encoding,
    params.newline,
    params.has_header,
    params.headers,
    params.lazy_quotes,
    params.trim_leading_space,
    params.fields_per_record,
  ])

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

        {paramFields.map(f => (
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

        {isCSVSource && (
          <div className="rounded-xl border border-gray-800 bg-gray-950 overflow-hidden">
            <div className="px-3 py-2 border-b border-gray-800 flex items-center justify-between">
              <div className="flex items-center gap-2 text-sm text-gray-200">
                <Eye size={14} /> Prévisualisation CSV
              </div>
              <button
                type="button"
                onClick={() => void runCSVPreview()}
                className="text-xs px-2 py-1 rounded-md bg-gray-800 hover:bg-gray-700 text-gray-300 flex items-center gap-1"
              >
                <RefreshCcw size={12} className={previewLoading ? 'animate-spin' : ''} /> Actualiser
              </button>
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
              <div className="p-3 text-xs text-gray-500">Renseigne le chemin puis ajuste l'encodage, le délimiteur et les en-têtes pour voir un aperçu des 20 premières lignes.</div>
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

function FilePickerInput({
  value, invalid, onChange, placeholder
}: {
  value: string
  invalid: boolean
  onChange: (v: string) => void
  placeholder?: string
}) {
  const fileRef = useRef<HTMLInputElement>(null)

  const handleFilePick = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const path = (file as any).path || file.name
    onChange(path)
  }

  return (
    <div className="flex gap-2">
      <input
        className={`${inputCls(invalid)} flex-1`}
        value={value}
        placeholder={placeholder ?? '/data/input.csv'}
        onChange={e => onChange(e.target.value)}
      />
      <button
        type="button"
        title="Parcourir (copier le nom)"
        onClick={() => fileRef.current?.click()}
        className="flex-shrink-0 px-2 py-1.5 bg-gray-700 hover:bg-gray-600 border border-gray-600 rounded-lg text-gray-300 transition-colors"
      >
        <FolderOpen size={15} />
      </button>
      <input
        ref={fileRef}
        type="file"
        accept=".csv,.txt"
        className="hidden"
        onChange={handleFilePick}
      />
    </div>
  )
}

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
      {
        key: 'encoding', label: 'Encodage', required: true, type: 'select', help: 'Choisir l’encodage réel du fichier pour éviter les caractères illisibles.',
        options: [
          { value: 'utf-8', label: 'UTF-8' },
          { value: 'windows-1252', label: 'Windows-1252' },
          { value: 'iso-8859-1', label: 'ISO-8859-1 / Latin-1' },
          { value: 'utf-16le', label: 'UTF-16 LE' },
          { value: 'utf-16be', label: 'UTF-16 BE' },
        ],
      },
      {
        key: 'delimiter', label: 'Délimiteur', help: 'Exemples : , ; | ou tab', type: 'select',
        options: [
          { value: ',', label: 'Virgule (,)' },
          { value: ';', label: 'Point-virgule (;)' },
          { value: '|', label: 'Pipe (|)' },
          { value: '\t', label: 'Tabulation (TAB)' },
        ],
      },
      {
        key: 'newline', label: 'Retour à la ligne', type: 'select', help: 'Auto convient dans la majorité des cas. Utiliser CR pour anciens exports Mac.',
        options: [
          { value: 'auto', label: 'Auto (LF / CRLF)' },
          { value: 'cr', label: 'CR uniquement (ancien Mac)' },
        ],
      },
      {
        key: 'has_header', label: 'Présence d’en-tête', required: true, type: 'select', help: 'Si Non, il faut renseigner les noms de colonnes manuellement.',
        options: [
          { value: 'true', label: 'Oui, la première ligne contient les colonnes' },
          { value: 'false', label: 'Non, le fichier commence directement par les données' },
        ],
      },
      { key: 'headers', label: 'Colonnes manuelles', placeholder: 'id,name,amount', help: 'Obligatoire si le fichier ne contient pas d’en-tête.' },
      {
        key: 'skip_empty_lines', label: 'Ignorer lignes vides', type: 'select',
        options: [
          { value: 'true', label: 'Oui' },
          { value: 'false', label: 'Non' },
        ],
      },
      {
        key: 'trim_leading_space', label: 'Supprimer espaces de début', type: 'select',
        options: [
          { value: 'true', label: 'Oui' },
          { value: 'false', label: 'Non' },
        ],
      },
      {
        key: 'lazy_quotes', label: 'Tolérance guillemets imparfaits', type: 'select', help: 'Pratique pour des CSV sales issus d’exports métiers.',
        options: [
          { value: 'true', label: 'Oui' },
          { value: 'false', label: 'Non' },
        ],
      },
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
    case 'transform.dummy': return []
    case 'transform.filter_advanced': return [
      { key: 'field', label: 'Champ à évaluer', placeholder: 'amount', required: true },
      {
        key: 'operator', label: 'Opérateur', required: true, type: 'select',
        options: [
          { value: 'eq',           label: '= égal' },
          { value: 'neq',          label: '≠ différent' },
          { value: 'gt',           label: '> supérieur' },
          { value: 'gte',          label: '≥ sup. ou égal' },
          { value: 'lt',           label: '< inférieur' },
          { value: 'lte',          label: '≤ inf. ou égal' },
          { value: 'contains',     label: 'contient' },
          { value: 'not_contains', label: 'ne contient pas' },
          { value: 'starts_with',  label: 'commence par' },
          { value: 'ends_with',    label: 'termine par' },
          { value: 'is_null',      label: 'est null' },
          { value: 'is_not_null',  label: 'n\'est pas null' },
          { value: 'is_true',      label: 'est vrai (bool)' },
          { value: 'is_false',     label: 'est faux (bool)' },
        ],
      },
      { key: 'value', label: 'Valeur de comparaison', placeholder: '100' },
      {
        key: 'value_type', label: 'Type de valeur', type: 'select',
        options: [
          { value: 'string', label: 'Texte (string)' },
          { value: 'number', label: 'Nombre (number)' },
          { value: 'bool',   label: 'Booléen (bool)' },
        ],
      },
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
    case 'transform.dedup': return [
      { key: 'keys', label: 'Clés (virgule)', placeholder: 'id, email', required: true },
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
    default: return []
  }
}
