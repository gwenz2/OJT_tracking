/**
 * Staff attendance monitoring — filterable, paginated table over scoped
 * sessions with flags and journal status.
 */
import { useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { monitoringApi } from '@/features/corrections/api'
import { qk } from '@/api/queryKeys'
import { ApiError } from '@/api/client'
import { Badge } from '@/components/ui/badge'
import { Input, Label } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Pagination } from '@/components/ui/pagination'
import { LoadingState, EmptyState, ErrorState, AccessDenied } from '@/components/feedback/states'
import { formatDate, formatDateTime, formatMinutes } from '@/lib/utils'
import { Camera } from 'lucide-react'

const statusVariant: Record<string, 'default' | 'success' | 'warning' | 'danger'> = {
  open: 'default', valid: 'success', flagged: 'warning', corrected: 'default',
}

export function StaffAttendancePage() {
  const [params, setParams] = useSearchParams()
  const [page, setPage] = useState(1)
  const [from, setFrom] = useState(params.get('from') ?? '')
  const [to, setTo] = useState(params.get('to') ?? '')
  const [status, setStatus] = useState(params.get('status') ?? '')

  const query = { from, to, status, page }
  const { data, isLoading, error } = useQuery({
    queryKey: qk.staffAttendance(query),
    queryFn: () => monitoringApi.attendance(query),
  })

  if (error instanceof ApiError && error.status === 403) return <AccessDenied />

  const applyFilters = () => {
    setPage(1)
    setParams(Object.fromEntries(Object.entries({ from, to, status }).filter(([, v]) => v)))
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-end gap-3">
        <div>
          <Label htmlFor="f-from">From</Label>
          <Input id="f-from" type="date" value={from} onChange={(e) => setFrom(e.target.value)} />
        </div>
        <div>
          <Label htmlFor="f-to">To</Label>
          <Input id="f-to" type="date" value={to} onChange={(e) => setTo(e.target.value)} />
        </div>
        <div>
          <Label htmlFor="f-status">Status</Label>
          <Select id="f-status" value={status} onChange={(e) => setStatus(e.target.value)} className="w-36">
            <option value="">All</option>
            <option value="open">Open</option>
            <option value="valid">Valid</option>
            <option value="flagged">Flagged</option>
            <option value="corrected">Corrected</option>
          </Select>
        </div>
        <button
          type="button"
          onClick={applyFilters}
          className="h-10 rounded-[var(--radius-md)] border border-[var(--color-border-strong)] px-4 text-sm hover:bg-[var(--color-surface-hover)]"
        >
          Apply
        </button>
      </div>

      {isLoading && <LoadingState label="Loading attendance…" />}
      {error && !(error instanceof ApiError && error.status === 403) && (
        <ErrorState description={error instanceof ApiError ? error.message : 'Failed to load'} />
      )}
      {data && data.items.length === 0 && (
        <EmptyState title="No records" description="No attendance sessions match these filters." />
      )}

      {data && data.items.length > 0 && (
        <div className="overflow-x-auto rounded-lg border border-[var(--color-border)]">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-[var(--color-border)] bg-[var(--color-surface-hover)] text-left">
                <th className="px-3 py-2 font-medium">Date</th>
                <th className="px-3 py-2 font-medium">Trainee</th>
                <th className="px-3 py-2 font-medium">Site</th>
                <th className="px-3 py-2 font-medium">Time in</th>
                <th className="px-3 py-2 font-medium">Time out</th>
                <th className="px-3 py-2 font-medium">Credited</th>
                <th className="px-3 py-2 font-medium">Status</th>
                <th className="px-3 py-2 font-medium">Journal</th>
                <th className="px-3 py-2 font-medium">Evidence</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[var(--color-border)]">
              {data.items.map((r) => (
                <tr key={r.id} className="hover:bg-[var(--color-surface-hover)]">
                  <td className="px-3 py-2 whitespace-nowrap">{formatDate(r.date)}</td>
                  <td className="px-3 py-2">
                    <Link to={`/staff/attendance/${r.id}`} className="hover:underline">
                      {r.trainee_name}
                    </Link>
                    <span className="block text-xs text-[var(--color-text-muted)]">{r.student_number}</span>
                  </td>
                  <td className="px-3 py-2">{r.site_name}</td>
                  <td className="px-3 py-2 whitespace-nowrap">{formatDateTime(r.time_in_at)}</td>
                  <td className="px-3 py-2 whitespace-nowrap">
                    {r.time_out_at ? formatDateTime(r.time_out_at) : '—'}
                  </td>
                  <td className="px-3 py-2">
                    {r.credited_minutes != null ? formatMinutes(r.credited_minutes) : '—'}
                  </td>
                  <td className="px-3 py-2">
                    <Badge variant={statusVariant[r.status] ?? 'default'}>{r.status}</Badge>
                    {r.flags.map((f) => (
                      <Badge key={f} variant="warning" className="ml-1">{f.replaceAll('_', ' ')}</Badge>
                    ))}
                  </td>
                  <td className="px-3 py-2">{r.journal_status ?? '—'}</td>
                  <td className="px-3 py-2">
                    <Link
                      to={`/staff/attendance/${r.id}`}
                      className="inline-flex min-h-[36px] items-center gap-1.5 rounded-[var(--radius-md)] px-2 text-sm font-medium text-[var(--color-primary)] hover:bg-[var(--color-primary)]/10 focus-ring"
                    >
                      <Camera size={16} aria-hidden="true" />
                      View evidence
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {data && <Pagination meta={data.meta} onPage={setPage} />}
    </div>
  )
}
