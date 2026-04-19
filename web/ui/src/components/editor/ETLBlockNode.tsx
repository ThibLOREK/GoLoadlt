import { Handle, Position } from '@xyflow/react'
import { useEditorStore } from '@/store/editorStore'
import { useNodeValidation } from '@/hooks/useNodeValidation'

interface ETLBlockNodeData {
  label: string
  blockType: string
  params: Record<string, string>
  connRef?: string
  disabled?: boolean
}

interface Props {
  id: string
  data: ETLBlockNodeData
  selected?: boolean
}

// Couleur d'accentuation par catégorie de bloc
const CATEGORY_COLORS: Record<string, string> = {
  source:    '#4ade80', // vert
  transform: '#60a5fa', // bleu
  target:    '#f97316', // orange
  lookup:    '#c084fc', // violet
}

function categoryOf(blockType: string): string {
  return blockType.split('.')[0] ?? 'transform'
}

export default function ETLBlockNode({ id, data, selected }: Props) {
  const { nodes } = useEditorStore()
  const validationMap = useNodeValidation(nodes)
  const validation = validationMap.get(id)
  const isValid = validation?.valid ?? true
  const missingParams = validation?.missing ?? []

  const category = categoryOf(data.blockType)
  const accentColor = CATEGORY_COLORS[category] ?? '#94a3b8'
  const isDisabled = data.disabled ?? false

  return (
    <div
      style={{
        background: isDisabled ? '#1a1d26' : '#1e2235',
        border: `1.5px solid ${
          isDisabled ? '#374151'
          : selected ? '#818cf8'
          : !isValid ? '#ef4444'
          : '#2e3554'
        }`,
        borderRadius: 10,
        minWidth: 160,
        maxWidth: 220,
        fontFamily: 'inherit',
        boxShadow: isDisabled
          ? 'none'
          : selected
            ? '0 0 0 2px #818cf866'
            : '0 2px 8px #0007',
        opacity: isDisabled ? 0.45 : 1,
        transition: 'opacity 0.2s, border-color 0.2s, box-shadow 0.2s',
        position: 'relative',
        overflow: 'hidden',
      }}
    >
      {/* Bande colorée en haut selon la catégorie */}
      <div style={{
        height: 3,
        background: isDisabled ? '#374151' : accentColor,
        borderRadius: '10px 10px 0 0',
        transition: 'background 0.2s',
      }} />

      <div style={{ padding: '8px 12px 10px' }}>
        {/* Nom du type (petit + muted) */}
        <div style={{
          fontSize: 9,
          color: isDisabled ? '#4b5563' : '#6b7280',
          textTransform: 'uppercase',
          letterSpacing: '0.08em',
          marginBottom: 3,
        }}>
          {data.blockType}
        </div>

        {/* Label principal */}
        <div style={{
          fontWeight: 600,
          fontSize: 13,
          color: isDisabled ? '#4b5563' : '#e2e8f0',
          whiteSpace: 'nowrap',
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          display: 'flex',
          alignItems: 'center',
          gap: 5,
        }}>
          {isDisabled && <span style={{ fontSize: 11 }}>⏸</span>}
          {!isDisabled && !isValid && <span style={{ fontSize: 11 }}>⚠️</span>}
          {data.label}
        </div>

        {/* Paramètres manquants (uniquement si actif) */}
        {!isDisabled && missingParams.length > 0 && (
          <div style={{ marginTop: 5 }}>
            {missingParams.map(p => (
              <div key={p} style={{
                fontSize: 9,
                color: '#f87171',
                background: '#450a0a55',
                borderRadius: 3,
                padding: '1px 5px',
                marginBottom: 2,
                display: 'inline-block',
                marginRight: 3,
              }}>
                {p} manquant
              </div>
            ))}
          </div>
        )}

        {/* Badge désactivé */}
        {isDisabled && (
          <div style={{
            marginTop: 5,
            fontSize: 9,
            color: '#4b5563',
            fontStyle: 'italic',
          }}>
            désactivé
          </div>
        )}
      </div>

      <Handle
        type="target"
        position={Position.Left}
        style={{ background: isDisabled ? '#374151' : accentColor, width: 10, height: 10, border: '2px solid #1e2235' }}
      />
      <Handle
        type="source"
        position={Position.Right}
        style={{ background: isDisabled ? '#374151' : accentColor, width: 10, height: 10, border: '2px solid #1e2235' }}
      />
    </div>
  )
}
