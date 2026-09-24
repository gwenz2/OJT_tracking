import { useEffect, useRef, type ReactNode } from 'react'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'

interface ModalProps {
  open: boolean
  onClose: () => void
  title: string
  children: ReactNode
  footer?: ReactNode
  className?: string
}

/**
 * Accessible dialog that behaves as a bottom sheet on phones and a centered
 * dialog on larger screens: Escape closes, focus moves inside on open and
 * returns on close, backdrop click closes, body scroll locks while open.
 * Not a full focus-trap — the forms inside are short and linear.
 */
export function Modal({ open, onClose, title, children, footer, className }: ModalProps) {
  const ref = useRef<HTMLDivElement>(null)
  const previousFocus = useRef<HTMLElement | null>(null)

  useEffect(() => {
    if (!open) return
    previousFocus.current = document.activeElement as HTMLElement | null
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', onKey)
    document.body.style.overflow = 'hidden'
    ref.current?.querySelector<HTMLElement>('input, select, textarea, button')?.focus()
    return () => {
      document.removeEventListener('keydown', onKey)
      document.body.style.overflow = ''
      previousFocus.current?.focus()
    }
  }, [open, onClose])

  if (!open) return null

  return (
    <div
      className="fixed inset-0 z-50 flex items-end justify-center sm:items-center"
      role="dialog"
      aria-modal="true"
      aria-label={title}
    >
      <div className="absolute inset-0 bg-black/40" onClick={onClose} aria-hidden="true" />
      <div
        ref={ref}
        tabIndex={-1}
        className={cn(
          'sheet-panel relative flex max-h-[88dvh] w-full max-w-lg flex-col overflow-hidden rounded-t-[var(--radius-lg)] bg-[var(--color-surface)] shadow-[var(--shadow-lg)] sm:rounded-[var(--radius-lg)]',
          className,
        )}
      >
        <div
          aria-hidden="true"
          className="mx-auto mt-2 h-1 w-10 shrink-0 rounded-full bg-[var(--color-border-strong)] sm:hidden"
        />
        <div className="flex shrink-0 items-center justify-between gap-3 px-5 pb-2 pt-3 sm:pt-5">
          <h2 className="text-base font-semibold">{title}</h2>
          <button
            onClick={onClose}
            aria-label="Close"
            className="-mr-2 flex h-11 w-11 items-center justify-center rounded-[var(--radius-md)] text-[var(--color-text-muted)] hover:bg-[var(--color-surface-hover)] focus-ring"
          >
            <X size={20} />
          </button>
        </div>
        <div className="overflow-y-auto overscroll-contain px-5 pb-[max(1.25rem,var(--safe-bottom))]">
          {children}
          {footer && <div className="mt-5 flex justify-end gap-2">{footer}</div>}
        </div>
      </div>
    </div>
  )
}
