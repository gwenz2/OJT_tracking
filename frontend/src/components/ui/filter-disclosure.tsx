import type { ReactNode } from 'react'
import { ChevronDown, ChevronUp, SlidersHorizontal } from 'lucide-react'

export function FilterDisclosure({
  open,
  onToggle,
  summary,
  children,
}: {
  open: boolean
  onToggle: () => void
  summary: string
  children: ReactNode
}) {
  return (
    <div className="min-w-0 flex-1">
      <button
        type="button"
        onClick={onToggle}
        aria-expanded={open}
        className="focus-ring flex h-11 w-full items-center gap-2 rounded-[var(--radius-md)] border border-[var(--color-border-strong)] bg-[var(--color-surface)] px-3 text-left text-sm md:hidden"
      >
        <SlidersHorizontal size={17} aria-hidden="true" />
        <span className="font-medium">Filters</span>
        <span className="min-w-0 flex-1 truncate text-right text-[var(--color-text-muted)]">{summary}</span>
        {open
          ? <ChevronUp size={16} aria-hidden="true" className="shrink-0" />
          : <ChevronDown size={16} aria-hidden="true" className="shrink-0" />}
      </button>
      <div
        data-open={open}
        className="filter-disclosure-panel grid grid-cols-2 rounded-[var(--radius-md)] bg-[var(--color-surface)] md:flex md:flex-wrap md:items-end md:gap-3 md:bg-transparent"
      >
        {children}
      </div>
    </div>
  )
}
