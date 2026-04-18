import { Handle, Position, type NodeProps } from '@xyflow/react'
import Badge from '@/components/ui/Badge'
import { useEditorStore } from '@/store/editorStore'
import { useNodeValidation } from '@/hooks/useNodeValidation'
import { AlertTriangle, CheckCircle2 } from 'lucide-react'

const categoryColors: Record<string, string> = {
  input:     'border-blue-600 bg-blue-950',
  output:    'border-green-600 bg-green-950',
  transform: 'border-purple-600 bg-purple-950',
  analytics: 'border-orange-600 bg-orange-950',
  ml:        'border-pink-600 bg-pink-950',
}

export default function ETLBlockNode({ id, data, selected }: NodeProps) {
  const { catalogue, nodes } = useEditorStore()
  const meta = catalogue.find(b => b.type === data.blockType)
  const category = (meta as any)?.category ?? 'transform'
  const colorCls = categoryColors[category] ?? 'border-gray-600 bg-gray-900'

  const validationMap = useNodeValidation(nodes)
  const validation = validationMap.get(id)
  const isValid = validation?.valid ?? true

  return (
    <div
      className={`min-w-[150px] border-2 rounded-xl px-3 py-2.5 shadow-lg transition-all ${
        selected ? 'ring-2 ring-brand-500 ring-offset-1 ring-offset-transparent' : ''
      } ${
        !isValid ? 'border-red-500 bg-red-950/40' : colorCls
      }`}
    >
      {/* Port d'entrée */}
      {((meta as any)?.minInputs ?? 1) > 0 && (
        <Handle type="target" position={Position.Left} style={{ background: '#60a5fa', width: 10, height: 10, left: -6 }} />
      )}

      <div className="flex flex-col gap-1">
        <div className="flex items-center justify-between gap-1">
          <span className="text-xs font-bold text-gray-100 truncate max-w-[110px]">{data.label as string}</span>
          {isValid
            ? <CheckCircle2 size={12} className="text-green-400 flex-shrink-0" />
            : <AlertTriangle size={12} className="text-red-400 flex-shrink-0 animate-pulse" />
          }
        </div>
        <Badge category={category} />
        {data.connRef && (
          <div className="text-xs text-gray-500 truncate">🔗 {data.connRef as string}</div>
        )}
        {!isValid && (
          <div className="text-xs text-red-400 mt-0.5">
            Manquant : {validation?.missing.join(', ')}
          </div>
        )}
      </div>

      {/* Port de sortie */}
      {((meta as any)?.minOutputs ?? 1) > 0 && (
        <Handle type="source" position={Position.Right} style={{ background: '#a78bfa', width: 10, height: 10, right: -6 }} />
      )}
    </div>
  )
}
