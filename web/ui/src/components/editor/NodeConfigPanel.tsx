import type { Node, Dispatch, SetStateAction } from 'react'
import { useEditorStore } from '@/store/editorStore'
import { X, Trash2 } from 'lucide-react'
import Button from '@/components/ui/Button'
import Badge from '@/components/ui/Badge'

interface Props {
  nodeId: string
  nodes: Node[]
  setNodes: Dispatch<SetStateAction<Node[]>>
}

export default function NodeConfigPanel({ nodeId, nodes, setNodes }: Props) {
  const { catalogue, selectNode, updateNodeParam } = useEditorStore()
  const node = nodes.find(n => n.id === nodeId)
  if (!node) return null

  const meta = catalogue.find(b => b.type === node.data.blockType)
  const params = node.data.params as Record<string, string>

  const handleDelete = () => {
    setNodes(nds => nds.filter(n => n.id !== nodeId))
    selectNode(null)
  }

  // Définit les champs à afficher selon le type de bloc
  const paramFields = getParamFields(node.data.blockType as string)

  return (
    <aside className="w-72 flex-shrink-0 bg-gray-900 border-l border-gray-800 flex flex-col overflow-y-auto">
      <div className="flex items-center justify-between px-4 py-3 border-b border-gray-800">
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold text-gray-100">{node.data.label as string}</span>
          {meta && <Badge category={meta.category} />}
        </div>
        <button onClick={() => selectNode(null)} className="text-gray-500 hover:text-gray-200">
          <X size={16} />
        </button>
      </div>

      <div className="px-4 py-4 space-y-4 flex-1">
        {/* Label */}
        <Field label="Label">
          <input
            className="input-base"
            value={node.data.label as string}
            onChange={e => {
              setNodes(nds => nds.map(n => n.id === nodeId ? { ...n, data: { ...n.data, label: e.target.value } } : n))
            }}
          />
        </Field>

        {/* Connexion */}
        {(meta?.minInputs ?? 0) > 0 || (node.data.blockType as string).startsWith('source.') ? (
          <Field label="Réf. connexion (connRef)">
            <input
              className="input-base"
              value={node.data.connRef as string}
              placeholder="ex: conn-crm"
              onChange={e => setNodes(nds => nds.map(n => n.id === nodeId ? { ...n, data: { ...n.data, connRef: e.target.value } } : n))}
            />
          </Field>
        ) : null}

        {/* Paramètres spécifiques au type de bloc */}
        {paramFields.map(f => (
          <Field key={f.key} label={f.label}>
            {f.multiline ? (
              <textarea
                className="input-base resize-none h-20"
                value={params[f.key] ?? ''}
                placeholder={f.placeholder}
                onChange={e => updateNodeParam(nodeId, f.key, e.target.value)}
              />
            ) : (
              <input
                className="input-base"
                value={params[f.key] ?? ''}
                placeholder={f.placeholder}
                onChange={e => updateNodeParam(nodeId, f.key, e.target.value)}
              />
            )}
          </Field>
        ))}

        {/* Type du bloc (lecture seule) */}
        <div className="pt-2 border-t border-gray-800">
          <p className="text-xs text-gray-600 font-mono">{node.data.blockType as string}</p>
          {meta && <p className="text-xs text-gray-600 mt-1">{meta.description}</p>}
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

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <label className="block text-xs text-gray-400 mb-1">{label}</label>
      {children}
    </div>
  )
}

interface ParamField { key: string; label: string; placeholder?: string; multiline?: boolean }

function getParamFields(blockType: string): ParamField[] {
  switch (blockType) {
    case 'source.csv':      return [{ key: 'path', label: 'Chemin fichier', placeholder: '/data/input.csv' }, { key: 'delimiter', label: 'Délimiteur', placeholder: ',' }]
    case 'source.postgres':
    case 'source.mysql':
    case 'source.mssql':   return [{ key: 'query', label: 'Requête SQL', placeholder: 'SELECT * FROM ...', multiline: true }]
    case 'target.csv':     return [{ key: 'path', label: 'Chemin fichier', placeholder: '/data/output.csv' }, { key: 'delimiter', label: 'Délimiteur', placeholder: ',' }, { key: 'append', label: 'Mode append (true/false)', placeholder: 'false' }]
    case 'target.postgres': return [{ key: 'table', label: 'Table cible', placeholder: 'schema.table' }]
    case 'transform.filter': return [{ key: 'condition', label: 'Condition', placeholder: 'amount > 100' }]
    case 'transform.select': return [{ key: 'columns', label: 'Colonnes (virgule)', placeholder: 'id, name, amount' }]
    case 'transform.cast':  return [{ key: 'column', label: 'Colonne', placeholder: 'price' }, { key: 'targetType', label: 'Type cible', placeholder: 'float' }]
    case 'transform.add_column': return [{ key: 'name', label: 'Nom colonne', placeholder: 'tax' }, { key: 'expression', label: 'Expression', placeholder: 'amount * 0.20' }]
    case 'transform.split': return [{ key: 'conditions', label: 'Conditions (virgule)', placeholder: 'amount > 1000, amount > 500' }]
    case 'transform.pivot': return [{ key: 'groupBy', label: 'Group By', placeholder: 'region' }, { key: 'pivotColumn', label: 'Colonne pivot', placeholder: 'product' }, { key: 'valueColumn', label: 'Colonne valeur', placeholder: 'amount' }, { key: 'aggregation', label: 'Agrégation', placeholder: 'SUM' }]
    case 'transform.aggregate': return [{ key: 'groupBy', label: 'Group By (virgule)', placeholder: 'region,category' }, { key: 'aggregations', label: 'Agrégations', placeholder: 'SUM(amount),COUNT(id)' }]
    case 'transform.sort':  return [{ key: 'columns', label: 'Colonnes (virgule)', placeholder: 'date,amount' }, { key: 'order', label: 'Ordre', placeholder: 'asc' }]
    case 'transform.dedup': return [{ key: 'keys', label: 'Clés (virgule)', placeholder: 'id,email' }]
    case 'transform.join':  return [{ key: 'leftKey', label: 'Clé gauche', placeholder: 'id' }, { key: 'rightKey', label: 'Clé droite', placeholder: 'user_id' }, { key: 'type', label: 'Type', placeholder: 'inner' }]
    case 'transform.unpivot': return [{ key: 'columns', label: 'Colonnes à dépivoter', placeholder: 'jan,fev,mar' }, { key: 'keyName', label: 'Nom clé', placeholder: 'mois' }, { key: 'valueName', label: 'Nom valeur', placeholder: 'montant' }]
    default: return []
  }
}
