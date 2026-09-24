import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'
import { Inbox, AlertTriangle, Loader2, ShieldAlert } from 'lucide-react'

export function EmptyState({
  title = 'Nothing here yet',
  description,
  action,
  className,
}: {
  title?: string
  description?: string
  action?: ReactNode
  className?: string
}) {
  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center rounded-[var(--radius-lg)] border border-dashed border-[var(--color-border)] p-10 text-center',
        className,
      )}
    >
      <Inbox size={32} className="text-[var(--color-text-subtle)]" />
      <p className="mt-3 text-sm font-medium text-[var(--color-text)]">{title}</p>
      {description && (
        <p className="mt-1 text-sm text-[var(--color-text-muted)]">{description}</p>
      )}
      {action && <div className="mt-4">{action}</div>}
    </div>
  )
}

export function ErrorState({
  title = 'Something went wrong',
  description,
  onRetry,
  className,
}: {
  title?: string
  description?: string
  onRetry?: () => void
  className?: string
}) {
  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center rounded-[var(--radius-lg)] border border-[var(--color-danger)]/30 p-10 text-center',
        className,
      )}
    >
      <AlertTriangle size={32} className="text-[var(--color-danger)]" />
      <p className="mt-3 text-sm font-medium text-[var(--color-text)]">{title}</p>
      {description && (
        <p className="mt-1 text-sm text-[var(--color-text-muted)]">{description}</p>
      )}
      {onRetry && (
        <button
          onClick={onRetry}
          className="mt-4 rounded-[var(--radius-md)] border border-[var(--color-border-strong)] px-4 py-2 text-sm font-medium text-[var(--color-text)] hover:bg-[var(--color-surface-hover)] focus-ring"
        >
          Try again
        </button>
      )}
    </div>
  )
}

export function AccessDenied({
  title = 'You don’t have access to this',
  description = 'Ask your coordinator or administrator if you think this is a mistake.',
  className,
}: {
  title?: string
  description?: string
  className?: string
}) {
  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center rounded-[var(--radius-lg)] border border-dashed border-[var(--color-border)] p-10 text-center',
        className,
      )}
    >
      <ShieldAlert size={32} className="text-[var(--color-warning)]" />
      <p className="mt-3 text-sm font-medium text-[var(--color-text)]">{title}</p>
      <p className="mt-1 text-sm text-[var(--color-text-muted)]">{description}</p>
    </div>
  )
}

export function LoadingState({ label = 'Loading…', className }: { label?: string; className?: string }) {
  return (
    <div
      className={cn(
        'flex items-center justify-center gap-2 rounded-[var(--radius-lg)] p-10 text-sm text-[var(--color-text-muted)]',
        className,
      )}
    >
      <Loader2 size={18} className="animate-spin" />
      {label}
    </div>
  )
}
