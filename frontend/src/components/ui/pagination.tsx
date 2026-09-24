import type { PageMeta } from '@/api/types'
import { Button } from '@/components/ui/button'

/** Shared prev/next pagination bar driven by the contract PageMeta. */
export function Pagination({ meta, onPage }: { meta: PageMeta | undefined; onPage: (page: number) => void }) {
  if (!meta || meta.total_pages <= 1) return null
  return (
    <div className="flex items-center justify-between pt-4 text-sm text-[var(--color-text-muted)]">
      <span>
        Page {meta.page} of {meta.total_pages} · {meta.total} total
      </span>
      <div className="flex gap-2">
        <Button variant="outline" size="sm" disabled={meta.page <= 1} onClick={() => onPage(meta.page - 1)}>
          Previous
        </Button>
        <Button
          variant="outline"
          size="sm"
          disabled={meta.page >= meta.total_pages}
          onClick={() => onPage(meta.page + 1)}
        >
          Next
        </Button>
      </div>
    </div>
  )
}
