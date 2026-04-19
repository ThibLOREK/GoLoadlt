import { useState } from 'react'
import { X, ChevronDown, Eye } from 'lucide-react'

interface PreviewNode {
  id: string
  data: { label: string; blockType: string }
}

interface Props {
  preview: Record<string, Record<string, any>[]>
  nodes: PreviewNode[]
  onClose: () => void
}

export default function DataPreviewPanel({ preview, nodes, onClose }: Props) {
  const blockIds = Object.keys(preview).filter(id => (preview[id] ?? []).length > 0)
  const [selectedId, setSelectedId] = useState(blockIds[0] ?? '')

  const rows = preview[selectedId] ?? []
  const headers = rows.length > 0 ? Object.keys(rows[0]) : []

  const getLabel = (id: string) =>
    nodes.find(n => n.id === id)?.data.label ?? id

  if (blockIds.length === 0) {
    return (
      <div className="border-t border-gray-800 bg-gray-950 flex items-center justify-between px-4 py-3" style={{ height: 52 }}>
        <div className="flex items-center gap-2 text-gray-500">
          <Eye size={14} />
          <span className="text-xs">Aucune donnée capturée — aucun flux de sortie trouvé</span>
        </div>
        <button onClick={onClose} className="text-gray-600 hover:text-gray-300 transition-colors">
          <X size={14} />
        </button>
      </div>
    )
  }

  return (
    <div className="border-t border-gray-800 bg-gray-950 flex flex-col" style={{ height: 300 }}>
      {/* Header style PDI */}
      <div className="flex items-center gap-3 px-4 py-2 border-b border-gray-800 flex-shrink-0 bg-gray-900">
        <Eye size={13} className="text-brand-400 flex-shrink-0" />
        <span className="text-xs font-semibold text-gray-400 uppercase tracking-wider">
          Aperçu données
        </span>

        {/* Sélecteur de bloc */}
        <div className="relative ml-2">
          <select
            value={selectedId}
            onChange={e => setSelectedId(e.target.value)}
            className="bg-gray-800 border border-gray-700 text-gray-200 text-xs rounded-lg pl-3 pr-7 py-1.5 appearance-none focus:outline-none focus:ring-1 focus:ring-brand-500 cursor-pointer"
          >
            {blockIds.map(id => (
              <option key={id} value={id}>
                {getLabel(id)}
              </option>
            ))}
          </select>
          <ChevronDown size={11} className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-500 pointer-events-none" />
        </div>

        <span className="text-xs text-gray-600">
          {rows.length >= 1000
            ? <span className="text-yellow-500">≥1000 lignes (troncature)</span>
            : `${rows.length} ligne${rows.length > 1 ? 's' : ''}`
          }
          {' — '}{headers.length} colonne{headers.length > 1 ? 's' : ''}
        </span>

        <button
          onClick={onClose}
          className="ml-auto text-gray-600 hover:text-gray-300 transition-colors"
          title="Fermer l'aperçu"
        >
          <X size={14} />
        </button>
      </div>

      {/* Tableau */}
      <div className="overflow-auto flex-1 text-xs">
        {rows.length === 0 ? (
          <div className="flex items-center justify-center h-full text-gray-600">
            Aucune donnée pour ce bloc.
          </div>
        ) : (
          <table className="w-full border-separate border-spacing-0">
            <thead className="sticky top-0 z-10">
              <tr>
                <th className="bg-gray-800/90 backdrop-blur text-gray-500 font-medium px-3 py-1.5 text-right border-b border-r border-gray-700/50 w-10 select-none">#</th>
                {headers.map(h => (
                  <th
                    key={h}
                    className="bg-gray-800/90 backdrop-blur text-gray-300 font-medium px-3 py-1.5 text-left border-b border-r border-gray-700/50 whitespace-nowrap"
                  >
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {rows.map((row, i) => (
                <tr
                  key={i}
                  className={`hover:bg-gray-800/40 transition-colors ${
                    i % 2 === 0 ? 'bg-gray-950' : 'bg-gray-900/30'
                  }`}
                >
                  <td className="px-3 py-1 text-gray-700 text-right border-r border-gray-800/50 select-none font-mono">
                    {i + 1}
                  </td>
                  {headers.map(h => (
                    <td
                      key={h}
                      className="px-3 py-1 text-gray-300 border-r border-gray-800/50 max-w-xs"
                      title={row[h] != null ? String(row[h]) : ''}
                    >
                      {row[h] == null
                        ? <span className="text-gray-700 italic">null</span>
                        : <span className="truncate block max-w-[200px]">{String(row[h])}</span>
                      }
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
