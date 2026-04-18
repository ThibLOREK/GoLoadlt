import { Handle, Position, type NodeProps } from '@xyflow/react'
import Badge from '@/components/ui/Badge'
import { useEditorStore } from '@/store/editorStore'

const categoryColors: Record<string, string> = {
  input:     'border-blue-600 bg-blue-950',
  output:    'border-green-600 bg-green-950',
  transform: 'border-purple-600 bg-purple-950',
  analytics: 'border-orange-600 bg-orange-950',
  ml:        'border-pink-600 bg-pink-950',
}

export default function ETLBlockNode({ id, data, selected }: NodeProps) {
  const { catalogue } = useEditorStore()
  const meta = catalogue.find(b => b.type === data.blockType)
  const category = meta?.category ?? 'transform'
  const colorCls = categoryColors[category] ?? 'border-gray-600 bg-gray-900'

  return (
    <div className={`min-w-[140px] border-2 rounded-xl px-3 py-2 shadow-lg transition-all ${
      selected ? 'ring-2 ring-brand-500 ring-offset-1 ring-offset-transparent' : ''
    } ${colorCls}`}>
      {/* Port d'entrée */}
      {(meta?.minInputs ?? 1) > 0 && (
        <Handle type="target" position={Position.Left} style={{ background: '#60a5fa', width: 10, height: 10, left: -6 }} />
      )}

      <div className="flex flex-col gap-1">
        <div className="text-xs font-bold text-gray-100 truncate max-w-[120px]">{data.label as string}</div>
        <Badge category={category} />
        {data.connRef && (
          <div className="text-xs text-gray-500 truncate">&#128279; {data.connRef as string}</div>
        )}
      </div>

      {/* Port de sortie */}
      {(meta?.minOutputs ?? 1) > 0 && (
        <Handle type="source" position={Position.Right} style={{ background: '#a78bfa', width: 10, height: 10, right: -6 }} />
      )}
    </div>
  )
}
