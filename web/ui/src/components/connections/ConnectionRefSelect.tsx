import { useConnections } from '@/hooks/useConnections'

interface Props {
  blockType: string
  value: string
  onChange: (v: string) => void
}

/**
 * Déduit le filterType depuis le blockType.
 * Ex: "source.postgres" → "postgres", "target.mysql" → "mysql"
 */
function filterTypeFromBlock(blockType: string): string | undefined {
  const parts = blockType.split('.')
  return parts.length >= 2 ? parts[1] : undefined
}

export default function ConnectionRefSelect({ blockType, value, onChange }: Props) {
  const filterType = filterTypeFromBlock(blockType)
  const connections = useConnections(filterType)

  return (
    <div>
      <select
        className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-gray-100
          focus:outline-none focus:ring-1 focus:ring-brand-500 focus:border-brand-500 transition-colors"
        value={value}
        onChange={e => onChange(e.target.value)}
      >
        <option value="">— aucune —</option>
        {connections.map(c => (
          <option key={c.id} value={c.id}>
            {c.name} ({c.type})
          </option>
        ))}
      </select>
      {connections.length === 0 && (
        <p className="mt-1 text-[11px] text-gray-500">
          Aucune connexion{filterType ? ` de type « ${filterType} »` : ''} disponible.
          Créez-en une dans l&rsquo;onglet Connexions.
        </p>
      )}
    </div>
  )
}
