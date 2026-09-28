import type { PageMeta } from '@/api/types'
import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight } from 'lucide-react'

function visiblePages(current: number, total: number) {
  const start = Math.max(1, Math.min(current - 2, total - 4))
  const end = Math.min(total, start + 4)
  return Array.from({ length: end - start + 1 }, (_, index) => start + index)
}

/** Shared, keyboard-accessible pagination driven by the contract PageMeta. */
export function Pagination({ meta, onPage }: { meta: PageMeta | undefined; onPage: (page: number) => void }) {
  if (!meta || meta.total_pages <= 1) return null

  const pages = visiblePages(meta.page, meta.total_pages)

  return (
    <nav aria-label="Pagination" className="flex flex-col gap-3 pt-2 text-sm text-[var(--color-text-muted)] sm:flex-row sm:items-center sm:justify-between">
      <p aria-live="polite">
        Page {meta.page} of {meta.total_pages} · {meta.total} total
      </p>
      <div className="flex items-center gap-1">
        <Button
          variant="outline"
          size="sm"
          aria-label="Go to previous page"
          disabled={meta.page <= 1}
          onClick={() => onPage(meta.page - 1)}
          className="px-2 sm:px-3"
        >
          <ChevronLeft size={16} aria-hidden="true" />
          <span className="sr-only sm:not-sr-only">Previous</span>
        </Button>
        <div className="flex items-center gap-1" aria-label="Page selection">
          {pages.map((page) => (
            <Button
              key={page}
              variant={page === meta.page ? 'secondary' : 'ghost'}
              size="sm"
              aria-label={`Go to page ${page}`}
              aria-current={page === meta.page ? 'page' : undefined}
              onClick={() => onPage(page)}
              className="w-9 px-0"
            >
              {page}
            </Button>
          ))}
        </div>
        <Button
          variant="outline"
          size="sm"
          aria-label="Go to next page"
          disabled={meta.page >= meta.total_pages}
          onClick={() => onPage(meta.page + 1)}
          className="px-2 sm:px-3"
        >
          <span className="sr-only sm:not-sr-only">Next</span>
          <ChevronRight size={16} aria-hidden="true" />
        </Button>
      </div>
    </nav>
  )
}
