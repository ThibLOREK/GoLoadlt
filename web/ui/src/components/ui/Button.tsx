import clsx from 'clsx'
import type { ButtonHTMLAttributes } from 'react'

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'ghost' | 'danger'
  size?: 'sm' | 'md'
}

export default function Button({ variant = 'primary', size = 'md', className, children, ...props }: Props) {
  return (
    <button
      className={clsx(
        'inline-flex items-center gap-2 rounded-lg font-medium transition-colors disabled:opacity-50',
        size === 'sm' ? 'px-3 py-1.5 text-xs' : 'px-4 py-2 text-sm',
        variant === 'primary' && 'bg-brand-600 hover:bg-brand-700 text-white',
        variant === 'ghost'   && 'bg-transparent hover:bg-gray-800 text-gray-300',
        variant === 'danger'  && 'bg-red-700 hover:bg-red-800 text-white',
        className
      )}
      {...props}
    >
      {children}
    </button>
  )
}
