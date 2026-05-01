import { useState } from 'react'
import type { ConnEnv } from '@/types/api'
import Button from '@/components/ui/Button'

interface Props {
  envName: string
  initial?: ConnEnv
  onSave: (env: ConnEnv) => void
  onCancel: () => void
}

const DEFAULT: ConnEnv = { name: '', host: '', port: 5432, database: '', user: '', secretRef: '' }

const TEXT_FIELDS: { key: keyof ConnEnv; label: string; placeholder?: string }[] = [
  { key: 'host',     label: 'Host',     placeholder: 'db.example.com' },
  { key: 'database', label: 'Database', placeholder: 'mydb' },
  { key: 'user',     label: 'User',     placeholder: 'app_user' },
]

export default function ConnectionEnvForm({ envName, initial, onSave, onCancel }: Props) {
  const [form, setForm] = useState<ConnEnv>({ ...DEFAULT, ...initial, name: envName })

  const set = (k: keyof ConnEnv, v: string | number) =>
    setForm(f => ({ ...f, [k]: v }))

  const canSave = form.host.trim() !== '' && form.database.trim() !== '' && form.user.trim() !== ''

  return (
    <div className="space-y-3">
      <h3 className="text-sm font-semibold text-gray-300 uppercase tracking-wider">{envName}</h3>

      {TEXT_FIELDS.map(({ key, label, placeholder }) => (
        <div key={key}>
          <label className="block text-xs text-gray-400 mb-1">{label}</label>
          <input
            className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-gray-100 focus:outline-none focus:border-brand-500 transition-colors"
            value={form[key] as string}
            onChange={e => set(key, e.target.value)}
            placeholder={placeholder}
          />
        </div>
      ))}

      <div>
        <label className="block text-xs text-gray-400 mb-1">Port</label>
        <input
          type="number"
          className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-gray-100 focus:outline-none focus:border-brand-500 transition-colors"
          value={form.port}
          onChange={e => set('port', parseInt(e.target.value, 10) || 5432)}
        />
      </div>

      <div>
        <label className="block text-xs text-gray-400 mb-1">
          Secret Ref{' '}
          <span className="text-gray-600 font-normal normal-case tracking-normal">
            (ex&nbsp;: {'${DB_PASSWORD}'})
          </span>
        </label>
        {/* Affiche la référence symbolique uniquement — jamais la valeur résolue */}
        <input
          className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm font-mono text-gray-100 focus:outline-none focus:border-brand-500 transition-colors"
          value={form.secretRef}
          onChange={e => set('secretRef', e.target.value)}
          placeholder="${DB_PASSWORD}"
          autoComplete="off"
        />
      </div>

      <div className="flex justify-end gap-2 pt-1">
        <Button variant="ghost" size="sm" onClick={onCancel}>
          Annuler
        </Button>
        <Button
          size="sm"
          onClick={() => onSave(form)}
          disabled={!canSave}
        >
          Enregistrer
        </Button>
      </div>
    </div>
  )
}
