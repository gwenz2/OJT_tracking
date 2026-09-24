import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import { qk } from '@/api/queryKeys'
import type { AttendanceSummary } from '@/api/types'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Select } from '@/components/ui/select'
import { Pagination } from '@/components/ui/pagination'
import { LoadingState, ErrorState, EmptyState } from '@/components/feedback/states'
import { formatMinutes, formatDate, formatTime } from '@/lib/utils'
import { ChevronRight, FileEdit } from 'lucide-react'

const statusVariant: Record<string, 'default' | 'success' | 'warning' | 'danger'> = {
  open: 'warning',
  valid: 'success',
  flagged: 'warning',
  corrected: 'default',
}

export function HistoryPage() {
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState('')

  const params = new URLSearchParams({ page: String(page), page_size: '25' })
  if (status) params.set('status', status)

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: qk.attendanceHistory({ page, status }),
    queryFn: () => api.getPaged<AttendanceSummary>(`/attendance/history?${params}`),
  })

  return (
    <div className="space-y-4 p-4">
      <div className="flex items-center justify-between gap-2">
        <h1 className="text-lg font-semibold">History</h1>
        <Link
          to="/corrections"
          className="flex min-h-[44px] items-center gap-1.5 rounded-[var(--radius-md)] px-2 text-sm font-medium text-[var(--color-primary)] hover:bg-[var(--color-surface-hover)] focus-ring"
        >
          <FileEdit size={16} />
          Corrections
        </Link>
      </div>

      <Select
        value={status}
        onChange={(e) => { setStatus(e.target.value); setPage(1) }}
        aria-label="Filter by status"
      >
        <option value="">All statuses</option>
        <option value="open">Open</option>
        <option value="valid">Valid</option>
        <option value="flagged">Flagged</option>
        <option value="corrected">Corrected</option>
      </Select>

      {isLoading && <LoadingState />}
      {error && <ErrorState description={error.message} onRetry={() => refetch()} />}
      {data && data.items.length === 0 && (
        <EmptyState title="No attendance yet" description="Your Time In/Out records will appear here." />
      )}
      {data?.items.map((a) => (
        <Link key={a.id} to={`/attendance/${a.id}`} className="block focus-ring rounded-[var(--radius-lg)]">
          <Card className="transition-colors hover:bg-[var(--color-surface-hover)]">
            <CardContent className="flex items-center justify-between gap-3 py-4">
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="font-medium">{formatDate(a.date)}</span>
                  <Badge variant={statusVariant[a.status] ?? 'default'} className="capitalize">
                    {a.status}
                  </Badge>
                  {a.journal_status && (
                    <Badge variant={a.journal_status === 'reviewed' ? 'success' : 'default'}>
                      journal {a.journal_status.replace('_', ' ')}
                    </Badge>
                  )}
                </div>
                <p className="mt-1 break-words text-xs text-[var(--color-text-muted)]">
                  {a.site_name} · In {formatTime(a.time_in_at)}
                  {a.time_out_at && ` · Out ${formatTime(a.time_out_at)}`}
                </p>
                {a.flags.length > 0 && (
                  <p className="mt-0.5 break-words text-xs text-[var(--color-warning)]">{a.flags.join(', ')}</p>
                )}
              </div>
              <div className="flex shrink-0 items-center gap-1 text-sm text-[var(--color-text-muted)]">
                {a.credited_minutes != null && <span className="tabular-nums">{formatMinutes(a.credited_minutes)}</span>}
                <ChevronRight size={16} aria-hidden="true" />
              </div>
            </CardContent>
          </Card>
        </Link>
      ))}
      {data && <Pagination meta={data.meta} onPage={setPage} />}
    </div>
  )
}
