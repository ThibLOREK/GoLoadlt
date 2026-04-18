import { useEditorStore } from '@/store/editorStore'
import Badge from '@/components/ui/Badge'
import { useState } from 'react'
import { Search } from 'lucide-react'

const CATEGORIES = ['input', 'output', 'transform', 'analytics', 'ml']

export default function BlockPalette() {
  const { catalogue } = useEditorStore()
  const [search, setSearch] = useState('')
  const [activeCategory, setActiveCategory] = useState<string | null>(null)

  const filtered = catalogue.filter(b => {
    const matchCat = !activeCategory || b.category === activeCategory
    const matchSearch = !search || b.label.toLowerCase().includes(search.toLowerCase()) || b.type.includes(search.toLowerCase())
    return matchCat && matchSearch
  })

  const onDragStart = (e: React.DragEvent, blockType: string) => {
    e.dataTransfer.setData('application/goloadit-block', blockType)
    e.dataTransfer.effectAllowed = 'move'
  }

  return (
    <aside className="w-56 flex-shrink-0 bg-gray-900 border-r border-gray-800 flex flex-col overflow-hidden">
      <div className="px-3 py-3 border-b border-gray-800">
        <p className="text-xs font-bold text-gray-400 uppercase tracking-wider mb-2">Blocs</p>
        <div className="relative">
          <Search size={12} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-500" />
          <input
            className="w-full bg-gray-800 rounded-lg pl-7 pr-2 py-1.5 text-xs text-gray-200 placeholder-gray-600 focus:outline-none focus:ring-1 focus:ring-brand-500"
            placeholder="Rechercher..."
            value={search}
            onChange={e => setSearch(e.target.value)}
          />
        </div>
        <div className="flex flex-wrap gap-1 mt-2">
          {CATEGORIES.map(cat => (
            <button
              key={cat}
              onClick={() => setActiveCategory(activeCategory === cat ? null : cat)}
              className={`text-xs px-2 py-0.5 rounded transition-colors ${
                activeCategory === cat ? 'bg-brand-600 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
              }`}
            >
              {cat}
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto py-2 px-2 space-y-1">
        {filtered.map(b => (
          <div
            key={b.type}
            draggable
            onDragStart={e => onDragStart(e, b.type)}
            title={b.description}
            className="flex items-center gap-2 px-2 py-2 rounded-lg cursor-grab active:cursor-grabbing hover:bg-gray-800 transition-colors group"
          >
            <div className="flex-1 min-w-0">
              <div className="text-xs font-medium text-gray-200 truncate">{b.label}</div>
              <div className="text-xs text-gray-600 truncate">{b.type}</div>
            </div>
            <Badge category={b.category} />
          </div>
        ))}
        {filtered.length === 0 && (
          <p className="text-xs text-gray-600 text-center py-4">Aucun bloc trouvé</p>
        )}
      </div>
    </aside>
  )
}
