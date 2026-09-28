/**
 * Staff correction queue — pending requests with original vs proposed times
 * and an approve/reject decision panel.
 */
import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { correctionsApi } from './api'
import { qk } from '@/api/queryKeys'
import { ApiError } from '@/api/client'
import type { Correction } from '@/api/types'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input, Textarea, Label } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Pagination } from '@/components/ui/pagination'
import { FilterDisclosure } from '@/components/ui/filter-disclosure'
import { LoadingState, EmptyState, ErrorState, AccessDenied } from '@/components/feedback/states'
import { useToast } from '@/components/feedback/toast'
import { formatDate, formatDateTime } from '@/lib/utils'
import { useDebounce } from '@/hooks/use-debounce'

const typeLabel: Record<string, string> = {
  missed_time_out: 'Forgot to time out',
  incorrect_time: 'Wrong recorded time',
  field_assignment: 'Field assignment',
  gps_issue: 'GPS/location problem',
  other: 'Other',
}

function DecisionPanel({ c, onDone }: { c: Correction; onDone: () => void }) {
  const toast = useToast()
  const qc = useQueryClient()
  const [comment, setComment] = useState('')

  const decide = useMutation({
    mutationFn: (d: 'approved' | 'rejected') => correctionsApi.decide(c.id, d, comment.trim() || undefined),
    onSuccess: () => {
      toast.success('Decision recorded')
      qc.invalidateQueries({ queryKey: ['staff', 'corrections'] })
      qc.invalidateQueries({ queryKey: qk.staffDashboard() })
      onDone()
    },
    onError: (e) => toast.error(e instanceof ApiError ? e.message : 'Decision failed'),
  })

  return (
    <CardContent className="space-y-3 border-t border-[var(--color-border)] pt-4">
      <div className="grid gap-2 text-sm sm:grid-cols-2">
        <div>
          <p className="text-xs text-[var(--color-text-muted)]">Proposed time in</p>
          <p>{c.proposed_time_in_at ? formatDateTime(c.proposed_time_in_at) : '—'}</p>
        </div>
        <div>
          <p className="text-xs text-[var(--color-text-muted)]">Proposed time out</p>
          <p>{c.proposed_time_out_at ? formatDateTime(c.proposed_time_out_at) : '—'}</p>
        </div>
      </div>
      <div>
        <Label htmlFor={`comment-${c.id}`}>Comment (visible to trainee)</Label>
        <Textarea id={`comment-${c.id}`} rows={2} value={comment}
          onChange={(e) => setComment(e.target.value)}
          placeholder="e.g. Verified against supervisor report" />
      </div>
      <div className="flex gap-2">
        <Button variant="danger" size="sm" loading={decide.isPending}
          onClick={() => decide.mutate('rejected')}>
          Reject
        </Button>
        <Button size="sm" loading={decide.isPending}
          onClick={() => decide.mutate('approved')}>
          Approve
        </Button>
      </div>
    </CardContent>
  )
}

export function StaffCorrectionsPage() {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState('pending')
  const [type, setType] = useState('')
  const [page, setPage] = useState(1)
  const pageSize = 10
  const [openId, setOpenId] = useState<string | null>(null)
  const [filtersOpen, setFiltersOpen] = useState(false)
  const debouncedSearch = useDebounce(search.trim(), 300)

  const { data, isLoading, error } = useQuery({
    queryKey: qk.staffCorrections({ q: debouncedSearch, status, type, page, pageSize }),
    queryFn: () => correctionsApi.staffList({ q: debouncedSearch, status, type, page, pageSize }),
  })

  if (error instanceof ApiError && error.status === 403) return <AccessDenied />

  const activeFilterCount = [search, status, type].filter(Boolean).length

  return (
    <div className="space-y-4">
      <FilterDisclosure
        open={filtersOpen}
        onToggle={() => setFiltersOpen((open) => !open)}
        summary={activeFilterCount ? `${activeFilterCount} active filter${activeFilterCount === 1 ? '' : 's'}` : 'All requests'}
      >
        <div className="col-span-2 min-w-0 md:min-w-64">
          <Label htmlFor="c-search">Search</Label>
          <Input
            id="c-search"
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
            placeholder="Trainee, site, reason, or type"
          />
        </div>
        <div className="min-w-0">
          <Label htmlFor="c-status">Status</Label>
          <Select id="c-status" value={status} onChange={(e) => { setStatus(e.target.value); setPage(1) }} className="w-full md:w-40">
            <option value="pending">Pending</option>
            <option value="approved">Approved</option>
            <option value="rejected">Rejected</option>
            <option value="">All</option>
          </Select>
        </div>
        <div className="col-span-2 min-w-0 md:col-span-1">
          <Label htmlFor="c-type">Type</Label>
          <Select id="c-type" value={type} onChange={(e) => { setType(e.target.value); setPage(1) }} className="w-full md:w-52">
            <option value="">All types</option>
            {Object.entries(typeLabel).map(([value, label]) => (
              <option key={value} value={value}>{label}</option>
            ))}
          </Select>
        </div>
      </FilterDisclosure>

      {isLoading && <LoadingState label="Loading requests…" />}
      {error && !(error instanceof ApiError && error.status === 403) && (
        <ErrorState description={error instanceof ApiError ? error.message : 'Failed to load'} />
      )}
      {data && data.items.length === 0 && (
        <EmptyState title="Queue clear" description="No correction requests match this filter." />
      )}

      {data && data.items.length > 0 && (
        <ul className="space-y-3">
          {data.items.map((c) => (
            <Card key={c.id}>
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle className="text-base">{c.trainee_name}</CardTitle>
                    <p className="text-sm text-[var(--color-text-muted)]">
                      {formatDate(c.attendance_date ?? '')} · {c.site_name} · {typeLabel[c.type] ?? c.type}
                    </p>
                  </div>
                  <Badge variant={c.status === 'pending' ? 'warning' : c.status === 'approved' ? 'success' : 'danger'}>
                    {c.status}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent className="pb-3">
                <p className="text-sm">{c.reason}</p>
                <p className="mt-1 text-xs text-[var(--color-text-muted)]">
                  Requested {formatDateTime(c.requested_at)}
                </p>
                {c.status === 'pending' && openId !== c.id && (
                  <Button variant="outline" size="sm" className="mt-2" onClick={() => setOpenId(c.id)}>
                    Review
                  </Button>
                )}
                {c.decision_comment && c.status !== 'pending' && (
                  <p className="mt-2 border-l-2 border-[var(--color-border)] pl-3 text-sm">
                    {c.decision_comment}
                  </p>
                )}
              </CardContent>
              {c.status === 'pending' && openId === c.id && (
                <DecisionPanel c={c} onDone={() => setOpenId(null)} />
              )}
            </Card>
          ))}
        </ul>
      )}

      {data && (
        <Pagination
          meta={data.meta}
          onPage={setPage}
        />
      )}
    </div>
  )
}
