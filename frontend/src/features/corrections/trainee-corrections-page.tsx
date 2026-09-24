/**
 * Trainee corrections — request a fix on an attendance day and track
 * decisions. Proposed fields required depend on the correction type.
 */
import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { correctionsApi } from './api'
import { api } from '@/api/client'
import { qk } from '@/api/queryKeys'
import { ApiError } from '@/api/client'
import type { AttendanceSummary, CorrectionType } from '@/api/types'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input, Label, Textarea } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Pagination } from '@/components/ui/pagination'
import { LoadingState, EmptyState, ErrorState } from '@/components/feedback/states'
import { useToast } from '@/components/feedback/toast'
import { formatDate, formatDateTime } from '@/lib/utils'
import { Link } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'

const TYPES: { value: CorrectionType; label: string; needs: 'in' | 'out' | 'any' | 'none' }[] = [
  { value: 'missed_time_out', label: 'Forgot to time out', needs: 'out' },
  { value: 'incorrect_time', label: 'Wrong recorded time', needs: 'any' },
  { value: 'field_assignment', label: 'Field assignment', needs: 'any' },
  { value: 'gps_issue', label: 'GPS/location problem', needs: 'none' },
  { value: 'other', label: 'Other', needs: 'none' },
]

const statusVariant = { pending: 'warning', approved: 'success', rejected: 'danger', cancelled: 'default' } as const

function RequestForm({ onDone }: { onDone: () => void }) {
  const toast = useToast()
  const qc = useQueryClient()
  const [attendanceId, setAttendanceId] = useState('')
  const [type, setType] = useState<CorrectionType>('missed_time_out')
  const [reason, setReason] = useState('')
  const [propIn, setPropIn] = useState('')
  const [propOut, setPropOut] = useState('')

  // Attendance choices come from the trainee's own history.
  const { data: history } = useQuery({
    queryKey: qk.attendanceHistory({ page_size: 50 }),
    queryFn: () => api.getPaged<AttendanceSummary>('/attendance/history?page_size=50'),
  })

  const needs = TYPES.find((t) => t.value === type)!.needs

  const create = useMutation({
    mutationFn: () =>
      correctionsApi.create(attendanceId, {
        type,
        reason,
        proposed_time_in_at: propIn ? new Date(propIn).toISOString() : undefined,
        proposed_time_out_at: propOut ? new Date(propOut).toISOString() : undefined,
      }),
    onSuccess: () => {
      toast.success('Correction request submitted')
      qc.invalidateQueries({ queryKey: qk.traineeCorrections() })
      onDone()
    },
    onError: (e) => toast.error(e instanceof ApiError ? e.message : 'Request failed'),
  })

  const valid =
    attendanceId && reason.trim() &&
    (needs === 'none' ||
      (needs === 'out' && propOut) ||
      (needs === 'in' && propIn) ||
      (needs === 'any' && (propIn || propOut)))

  return (
    <Card>
      <CardHeader><CardTitle className="text-base">Request a correction</CardTitle></CardHeader>
      <CardContent className="space-y-3">
        <div>
          <Label htmlFor="corr-att">Attendance day</Label>
          <Select id="corr-att" value={attendanceId} onChange={(e) => setAttendanceId(e.target.value)}>
            <option value="">Select a day…</option>
            {(history?.items ?? []).map((h) => (
              <option key={h.id} value={h.id}>
                {formatDate(h.date)} — {h.site_name} ({h.status})
              </option>
            ))}
          </Select>
        </div>
        <div>
          <Label htmlFor="corr-type">Type</Label>
          <Select id="corr-type" value={type} onChange={(e) => setType(e.target.value as CorrectionType)}>
            {TYPES.map((t) => <option key={t.value} value={t.value}>{t.label}</option>)}
          </Select>
        </div>
        <div>
          <Label htmlFor="corr-reason">Reason</Label>
          <Textarea id="corr-reason" rows={2} value={reason} onChange={(e) => setReason(e.target.value)}
            placeholder="Explain what happened" />
        </div>
        {(needs === 'in' || needs === 'any') && (
          <div>
            <Label htmlFor="corr-in">Proposed time in</Label>
            <Input id="corr-in" type="datetime-local" value={propIn} onChange={(e) => setPropIn(e.target.value)} />
          </div>
        )}
        {(needs === 'out' || needs === 'any') && (
          <div>
            <Label htmlFor="corr-out">Proposed time out</Label>
            <Input id="corr-out" type="datetime-local" value={propOut} onChange={(e) => setPropOut(e.target.value)} />
          </div>
        )}
        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={onDone}>Cancel</Button>
          <Button disabled={!valid} loading={create.isPending} onClick={() => create.mutate()}>
            Submit request
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

export function TraineeCorrectionsPage() {
  const [page, setPage] = useState(1)
  const [showForm, setShowForm] = useState(false)

  const { data, isLoading, error } = useQuery({
    queryKey: qk.traineeCorrections({ page }),
    queryFn: () => correctionsApi.mine(page),
  })

  return (
    <div className="space-y-4 p-4">
      <Link
        to="/history"
        className="-ml-2 inline-flex min-h-[44px] items-center gap-1.5 rounded-[var(--radius-md)] px-2 text-sm font-medium text-[var(--color-text-muted)] hover:text-[var(--color-text)] focus-ring"
      >
        <ArrowLeft size={16} aria-hidden="true" />
        Back
      </Link>
      <div className="flex items-center justify-between gap-2">
        <h1 className="text-xl font-semibold">Corrections</h1>
        {!showForm && (
          <Button onClick={() => setShowForm(true)}>New request</Button>
        )}
      </div>

      {showForm && <RequestForm onDone={() => setShowForm(false)} />}

      {isLoading && <LoadingState label="Loading requests…" />}
      {error && <ErrorState description={error instanceof ApiError ? error.message : 'Failed to load'} />}
      {data && data.items.length === 0 && !showForm && (
        <EmptyState title="No requests" description="If a recorded time is wrong, request a correction here." />
      )}

      {data && data.items.length > 0 && (
        <ul className="space-y-3">
          {data.items.map((c) => (
            <Card key={c.id}>
              <CardContent className="pt-4">
                <div className="flex items-start justify-between gap-2">
                  <div className="min-w-0">
                    <p className="font-medium">
                      {formatDate(c.attendance_date ?? '')} · {TYPES.find((t) => t.value === c.type)?.label ?? c.type}
                    </p>
                    <p className="break-words text-sm text-[var(--color-text-muted)]">{c.reason}</p>
                  </div>
                  <Badge variant={statusVariant[c.status]} className="shrink-0 capitalize">
                    {c.status.replace('_', ' ')}
                  </Badge>
                </div>
                {c.decision_comment && (
                  <p className="mt-2 border-l-2 border-[var(--color-border)] pl-3 text-sm">
                    {c.decision_comment}
                  </p>
                )}
                <p className="mt-1 text-xs text-[var(--color-text-muted)]">
                  Requested {formatDateTime(c.requested_at)}
                  {c.decided_at && ` · Decided ${formatDateTime(c.decided_at)}`}
                </p>
              </CardContent>
            </Card>
          ))}
        </ul>
      )}

      {data && <Pagination meta={data.meta} onPage={setPage} />}
    </div>
  )
}
