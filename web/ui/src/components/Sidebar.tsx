import { NavLink } from 'react-router-dom'
import { GitBranch, Database, Settings } from 'lucide-react'
import clsx from 'clsx'

const nav = [
  { to: '/projects',    label: 'Projets',    icon: GitBranch },
  { to: '/connections', label: 'Connexions', icon: Database },
]

export default function Sidebar() {
  return (
    <aside className="w-56 flex-shrink-0 bg-gray-900 border-r border-gray-800 flex flex-col">
      {/* Logo */}
      <div className="px-5 py-4 border-b border-gray-800">
        <span className="text-brand-500 font-bold text-xl tracking-tight">GoLoadIt</span>
        <span className="ml-2 text-gray-500 text-xs">ETL Studio</span>
      </div>

      {/* Navigation */}
      <nav className="flex-1 py-4 space-y-1 px-2">
        {nav.map(({ to, label, icon: Icon }) => (
          <NavLink
            key={to}
            to={to}
            className={({ isActive }) =>
              clsx(
                'flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors',
                isActive
                  ? 'bg-brand-600 text-white'
                  : 'text-gray-400 hover:bg-gray-800 hover:text-gray-100'
              )
            }
          >
            <Icon size={16} />
            {label}
          </NavLink>
        ))}
      </nav>

      {/* Env badge */}
      <EnvBadge />
    </aside>
  )
}

function EnvBadge() {
  const envColors: Record<string, string> = {
    dev:     'bg-green-900 text-green-300',
    preprod: 'bg-yellow-900 text-yellow-300',
    prod:    'bg-red-900 text-red-300',
  }
  // Lu depuis l'API au montage — simplifié ici en lecture localStorage
  const env = localStorage.getItem('activeEnv') ?? 'dev'
  return (
    <div className="px-5 py-4 border-t border-gray-800">
      <span className={clsx('px-2 py-1 rounded text-xs font-mono font-bold', envColors[env] ?? envColors.dev)}>
        ENV : {env.toUpperCase()}
      </span>
    </div>
  )
}
