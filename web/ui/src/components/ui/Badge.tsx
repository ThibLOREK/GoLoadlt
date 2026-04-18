import clsx from 'clsx'

const colors: Record<string, string> = {
  input:     'bg-blue-900 text-blue-300',
  output:    'bg-green-900 text-green-300',
  transform: 'bg-purple-900 text-purple-300',
  analytics: 'bg-orange-900 text-orange-300',
  ml:        'bg-pink-900 text-pink-300',
}

export default function Badge({ category }: { category: string }) {
  return (
    <span className={clsx('px-1.5 py-0.5 rounded text-xs font-semibold uppercase', colors[category] ?? 'bg-gray-700 text-gray-300')}>
      {category}
    </span>
  )
}
