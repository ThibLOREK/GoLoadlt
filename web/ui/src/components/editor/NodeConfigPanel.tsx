import { useRef } from 'react'
import type { Node, Dispatch, SetStateAction } from 'react'
import { useEditorStore } from '@/store/editorStore'
import { useNodeValidation } from '@/hooks/useNodeValidation'
import { X, Trash2, AlertTriangle, CheckCircle2, FolderOpen } from 'lucide-react'
import Button from '@/components/ui/Button'
import Badge from '@/components/ui/Badge'

interface Props {
  nodeId: string
  nodes: Node[]
  setNodes: Dispatch<SetStateAction<Node[]>>
}

// Met à jour un param dans rfNodes (source de vérité ReactFlow)
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
  if (!node) return null

  const meta = catalogue.find((b: any) => b.type === node.data.blockType)
  const params = (node.data.params ?? {}) as Record<string, string>

  const validationMap = useNodeValidation(nodes)
  const validation = validationMap.get(nodeId)

  const handleDelete = () => {
    setNodes(nds => nds.filter(n => n.id !== nodeId))
    selectNode(null)
  }

  const paramFields = getParamFields(node.data.blockType as string)
  const isFilePathField = (key: string) => key === 'path'
  const isCSVBlock = ['source.csv', 'target.csv'].includes(node.data.blockType as string)

  return (
    <aside className="w-80 flex-shrink-0 bg-gray-900 border-l border-gray-800 flex flex-col overflow-y-auto">
      {/* En-tête */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-gray-800">
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold text-gray-100">{node.data.label as string}</span>
          {meta && <Badge category={(meta as any).category} />}
        </div>
        <button onClick={() => selectNode(null)} className="text-gray-500 hover:text-gray-200">
          <X size={16} />
        </button>
      </div>

      {/* Bannière de validation */}
      {validation && (
        <div className={`mx-4 mt-3 px-3 py-2 rounded-lg flex items-start gap-2 text-xs ${
          validation.valid
            ? 'bg-green-900/30 border border-green-700 text-green-300'
            : 'bg-red-900/30 border border-red-700 text-red-300'
        }`}>
          {validation.valid
            ? <><CheckCircle2 size={13} className="mt-0.5 flex-shrink-0" /> Bloc correctement configuré</>
            : <><AlertTriangle size={13} className="mt-0.5 flex-shrink-0" /> Champs manquants : <strong className="ml-1">{validation.missing.join(', ')}</strong></>
          }
        </div>
      )}

      <div className="px-4 py-4 space-y-4 flex-1">
        {/* Label */}
        <Field label="Label">
          <input
            className={inputCls()}
            value={node.data.label as string}
            onChange={e => setNodes(nds => nds.map(n =>
              n.id === nodeId ? { ...n, data: { ...n.data, label: e.target.value } } : n
            ))}
          />
        </Field>

        {/* Connexion */}
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

        {/* Params spécifiques */}
        {paramFields.map(f => (
          <Field key={f.key} label={f.label} required={f.required} missing={validation?.missing.includes(f.key)}>
            {isCSVBlock && isFilePathField(f.key) ? (
              <FilePickerInput
                value={params[f.key] ?? ''}
                invalid={!!validation?.missing.includes(f.key)}
                onChange={v => updateParam(nodeId, f.key, v, setNodes)}
                placeholder={f.placeholder}
              />
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

// Champ avec bouton Parcourir
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
    if (file) onChange(file.name)
  }

  return (
    <div className="flex gap-2">
      <input
        className={`${inputCls(invalid)} flex-1`}
        value={value}
        placeholder={placeholder}
        onChange={e => onChange(e.target.value)}
      />
      <button
        type="button"
        title="Parcourir"
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
  label, required, missing, children,
}: {
  label: string
  required?: boolean
  missing?: boolean
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
    </div>
  )
}

function needsConnRef(blockType: string) {
  return ['source.postgres', 'source.mysql', 'source.mssql', 'target.postgres'].includes(blockType)
}

interface ParamField { key: string; label: string; placeholder?: string; multiline?: boolean; required?: boolean }

function getParamFields(blockType: string): ParamField[] {
  switch (blockType) {
    case 'source.csv': return [
      { key: 'path', label: 'Chemin fichier', placeholder: '/data/input.csv', required: true },
      { key: 'delimiter', label: 'Délimiteur', placeholder: ',' },
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
