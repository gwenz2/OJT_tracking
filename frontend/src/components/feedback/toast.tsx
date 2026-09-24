import {
  createContext,
  useCallback,
  useContext,
  useState,
  type ReactNode,
} from 'react'
import { CheckCircle2, AlertCircle, Info, X } from 'lucide-react'
import { cn } from '@/lib/utils'

type ToastVariant = 'success' | 'error' | 'info'

interface Toast {
  id: string
  title: string
  description?: string
  variant: ToastVariant
}

interface ToastContextValue {
  toast: (t: Omit<Toast, 'id'>) => void
  success: (title: string, description?: string) => void
  error: (title: string, description?: string) => void
  info: (title: string, description?: string) => void
}

const ToastContext = createContext<ToastContextValue | null>(null)

export function useToast(): ToastContextValue {
  const ctx = useContext(ToastContext)
  if (!ctx) throw new Error('useToast must be used within ToastProvider')
  return ctx
}

const variantStyles: Record<ToastVariant, string> = {
  success: 'border-[var(--color-success)]/30 text-[var(--color-success)]',
  error: 'border-[var(--color-danger)]/30 text-[var(--color-danger)]',
  info: 'border-[var(--color-primary)]/30 text-[var(--color-primary)]',
}

const variantIcon: Record<ToastVariant, ReactNode> = {
  success: <CheckCircle2 size={18} />,
  error: <AlertCircle size={18} />,
  info: <Info size={18} />,
}

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([])

  const remove = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }, [])

  const toast = useCallback(
    (t: Omit<Toast, 'id'>) => {
      const id = Math.random().toString(36).slice(2)
      setToasts((prev) => [...prev, { ...t, id }])
      setTimeout(() => remove(id), 5000)
    },
    [remove],
  )

  const value: ToastContextValue = {
    toast,
    success: (title, description) => toast({ title, description, variant: 'success' }),
    error: (title, description) => toast({ title, description, variant: 'error' }),
    info: (title, description) => toast({ title, description, variant: 'info' }),
  }

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div className="pointer-events-none fixed inset-x-4 bottom-[calc(var(--tabbar-h)+var(--safe-bottom)+1rem)] z-50 flex flex-col gap-2 sm:inset-x-auto sm:bottom-4 sm:right-4 sm:w-full sm:max-w-sm">
        {toasts.map((t) => (
          <div
            key={t.id}
            className={cn(
              'pointer-events-auto flex items-start gap-3 rounded-[var(--radius-md)] border bg-[var(--color-surface)] p-4 shadow-[var(--shadow-lg)]',
              variantStyles[t.variant],
            )}
            role="status"
          >
            <span className="mt-0.5 shrink-0">{variantIcon[t.variant]}</span>
            <div className="min-w-0 flex-1">
              <p className="text-sm font-medium text-[var(--color-text)]">{t.title}</p>
              {t.description && (
                <p className="mt-0.5 text-xs text-[var(--color-text-muted)]">{t.description}</p>
              )}
            </div>
            <button
              onClick={() => remove(t.id)}
              className="shrink-0 text-[var(--color-text-subtle)] hover:text-[var(--color-text)]"
              aria-label="Dismiss"
            >
              <X size={16} />
            </button>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  )
}
