/**
 * Staff journal queue — submitted/needs_revision journals for in-scope
 * trainees, ordered so submitted items surface first.
 */
import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { journalApi } from './api'
import { qk as queryKeys } from '@/api/queryKeys'
import { ApiError } from '@/api/client'
import type { JournalStatus } from '@/api/types'
import { Badge } from '@/components/ui/badge'
import { Pagination } from '@/components/ui/pagination'
import { Input, Label } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { FilterDisclosure } from '@/components/ui/filter-disclosure'
import { LoadingState, EmptyState, ErrorState, AccessDenied } from '@/components/feedback/states'
import { formatDate, formatDateTime } from '@/lib/utils'

const statusBadge: Record<JournalStatus, { variant: 'default' | 'success' | 'warning' | 'danger'; label: string }> = {
  draft: { variant: 'default', label: 'Draft' },
  submitted: { variant: 'warning', label: 'Submitted' },
  reviewed: { variant: 'success', label: 'Reviewed' },
  needs_revision: { variant: 'danger', label: 'Needs revision' },
}

export function StaffJournalQueuePage() {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState('')
  const [page, setPage] = useState(1)
  const [filtersOpen, setFiltersOpen] = useState(false)

  const { data, isLoading, error } = useQuery({
    queryKey: queryKeys.staffJournalQueue({ status, page }),
    queryFn: () => journalApi.queue(status, page),
  })

  if (error instanceof ApiError && error.status === 403) return <AccessDenied />

  const filteredItems = data?.items.filter((j) => {
    const haystack = `${j.trainee_name} ${j.student_number} ${j.site_name} ${j.status}`.toLowerCase()
    return !search.trim() || haystack.includes(search.trim().toLowerCase())
  }) ?? []
  const activeFilterCount = [search, status].filter(Boolean).length

  return (
    <div className="space-y-4">
      <FilterDisclosure
        open={filtersOpen}
        onToggle={() => setFiltersOpen((open) => !open)}
        summary={activeFilterCount ? `${activeFilterCount} active filter${activeFilterCount === 1 ? '' : 's'}` : 'All journals'}
      >
        <div className="col-span-2 min-w-0 md:min-w-64">
          <Label htmlFor="j-search">Search</Label>
          <Input
            id="j-search"
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
            placeholder="Trainee, student no., or site"
          />
        </div>
        <div className="min-w-0">
          <Label htmlFor="j-status">Status</Label>
          <Select
            id="j-status"
            value={status}
            onChange={(e) => { setStatus(e.target.value); setPage(1) }}
            className="w-full md:w-44"
        >
            <option value="">All statuses</option>
            <option value="submitted">Submitted</option>
            <option value="needs_revision">Needs revision</option>
            <option value="reviewed">Reviewed</option>
            <option value="draft">Draft</option>
          </Select>
        </div>
      </FilterDisclosure>

      {isLoading && <LoadingState label="Loading journals…" />}
      {error && !(error instanceof ApiError && error.status === 403) && (
        <ErrorState description={error instanceof ApiError ? error.message : 'Failed to load'} />
      )}

      {data && data.items.length === 0 && (
        <EmptyState title="No journals" description="Nothing needs your attention right now." />
      )}
      {data && data.items.length > 0 && filteredItems.length === 0 && (
        <EmptyState title="No matches" description="Try a different search term or status." />
      )}

      {data && filteredItems.length > 0 && (
        <ul className="divide-y divide-[var(--color-border)] rounded-lg border border-[var(--color-border)]">
          {filteredItems.map((j) => {
            const b = statusBadge[j.status]
            return (
              <li key={j.id}>
                <Link
                  to={`/staff/journals/${j.id}`}
                  className="flex items-center justify-between gap-3 px-4 py-3 hover:bg-[var(--color-surface-hover)]"
                >
                  <div className="min-w-0">
                    <p className="truncate font-medium">{j.trainee_name}</p>
                    <p className="truncate text-sm text-[var(--color-text-muted)]">
                      {j.student_number} · {j.site_name} · {formatDate(j.date)}
                    </p>
                  </div>
                  <div className="flex shrink-0 items-center gap-3">
                    {j.submitted_at && (
                      <span className="hidden text-xs text-[var(--color-text-muted)] sm:block">
                        {formatDateTime(j.submitted_at)}
                      </span>
                    )}
                    <Badge variant={b.variant}>{b.label}</Badge>
                  </div>
                </Link>
              </li>
            )
          })}
        </ul>
      )}

      {data && <Pagination meta={data.meta} onPage={setPage} />}
    </div>
  )
}
