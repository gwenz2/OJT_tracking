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
import { Select } from '@/components/ui/select'
import { LoadingState, EmptyState, ErrorState, AccessDenied } from '@/components/feedback/states'
import { formatDate, formatDateTime } from '@/lib/utils'

const statusBadge: Record<JournalStatus, { variant: 'default' | 'success' | 'warning' | 'danger'; label: string }> = {
  draft: { variant: 'default', label: 'Draft' },
  submitted: { variant: 'warning', label: 'Submitted' },
  reviewed: { variant: 'success', label: 'Reviewed' },
  needs_revision: { variant: 'danger', label: 'Needs revision' },
}

export function StaffJournalQueuePage() {
  const [status, setStatus] = useState('')
  const [page, setPage] = useState(1)

  const { data, isLoading, error } = useQuery({
    queryKey: queryKeys.staffJournalQueue({ status, page }),
    queryFn: () => journalApi.queue(status, page),
  })

  if (error instanceof ApiError && error.status === 403) return <AccessDenied />

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p className="text-sm text-[var(--color-text-muted)]">
            Submitted journals from trainees in your scope
          </p>
        </div>
        <Select
          value={status}
          onChange={(e) => { setStatus(e.target.value); setPage(1) }}
          aria-label="Filter by status"
          className="w-44"
        >
          <option value="">All statuses</option>
          <option value="submitted">Submitted</option>
          <option value="needs_revision">Needs revision</option>
          <option value="reviewed">Reviewed</option>
          <option value="draft">Draft</option>
        </Select>
      </div>

      {isLoading && <LoadingState label="Loading journals…" />}
      {error && !(error instanceof ApiError && error.status === 403) && (
        <ErrorState description={error instanceof ApiError ? error.message : 'Failed to load'} />
      )}

      {data && data.items.length === 0 && (
        <EmptyState title="No journals" description="Nothing needs your attention right now." />
      )}

      {data && data.items.length > 0 && (
        <ul className="divide-y divide-[var(--color-border)] rounded-lg border border-[var(--color-border)]">
          {data.items.map((j) => {
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
