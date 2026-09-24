import { useQuery } from '@tanstack/react-query'
import { useParams, Link } from 'react-router-dom'
import { api } from '@/api/client'
import { qk } from '@/api/queryKeys'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { LoadingState, ErrorState, AccessDenied } from '@/components/feedback/states'
import { ApiError } from '@/api/client'
import { useSession } from '@/features/auth/session'
import { formatMinutes, formatDate, formatDateTime } from '@/lib/utils'
import { ArrowLeft } from 'lucide-react'

interface EvidenceRow {
  action: string
  location_status: string
  distance_from_site_m?: number
  gps_accuracy_m?: number
  server_captured_at: string
  device_captured_at?: string
  location_exception_reason?: string
}

interface CorrectionRow {
  id: string
  type: string
  status: string
  reason: string
  proposed_time_in_at?: string
  proposed_time_out_at?: string
  decision_comment?: string
  requested_at: string
  decided_at?: string
}

interface Detail {
  id: string
  date: string
  status: string
  site_name: string
  original_time_in_at: string
  original_time_out_at?: string
  effective_time_in_at: string
  effective_time_out_at?: string
  credited_minutes?: number
  flags: string[]
  evidence: EvidenceRow[]
  correction_history: CorrectionRow[]
  journal_id?: string
}

export function AttendanceDetailPage() {
  const { id } = useParams<{ id: string }>()
  const { user } = useSession()
  const staffView = user?.role !== 'trainee'
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: qk.attendanceDetail(id!),
    queryFn: () => api.get<Detail>(`/attendance/${id}`),
    enabled: !!id,
    retry: false,
  })

  if (isLoading) return <LoadingState />
  if (error) {
    if (error instanceof ApiError && (error.status === 403 || error.status === 404)) {
      return <AccessDenied title="Attendance record not found" description="It may not exist, or it isn't yours." />
    }
    return <ErrorState description={error.message} onRetry={() => refetch()} />
  }
  if (!data) return null

  return (
    <div className="space-y-4 p-4">
      <Link
        to={staffView ? '/staff/attendance' : '/history'}
        className="-ml-2 inline-flex min-h-[44px] items-center gap-1.5 rounded-[var(--radius-md)] px-2 text-sm font-medium text-[var(--color-text-muted)] hover:text-[var(--color-text)] focus-ring"
      >
        <ArrowLeft size={16} aria-hidden="true" />
        Back
      </Link>
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h1 className="text-lg font-semibold">{formatDate(data.date)}</h1>
        <Badge variant={data.status === 'valid' ? 'success' : data.status === 'flagged' ? 'warning' : 'default'} className="capitalize">
          {data.status}
        </Badge>
      </div>

      <Card>
        <CardContent className="space-y-2 text-sm">
          <Row label="Site" value={data.site_name} />
          <Row label="Time in" value={formatDateTime(data.effective_time_in_at)} />
          <Row label="Time out" value={data.effective_time_out_at ? formatDateTime(data.effective_time_out_at) : '—'} />
          <Row label="Credited" value={data.credited_minutes != null ? formatMinutes(data.credited_minutes) : '—'} />
          {data.flags.length > 0 && <Row label="Flags" value={data.flags.join(', ')} />}
          {(data.status === 'corrected' ||
            data.original_time_in_at !== data.effective_time_in_at ||
            data.original_time_out_at !== data.effective_time_out_at) && (
            <>
              <Row label="Original in" value={formatDateTime(data.original_time_in_at)} />
              <Row label="Original out" value={data.original_time_out_at ? formatDateTime(data.original_time_out_at) : '—'} />
            </>
          )}
        </CardContent>
      </Card>

      {data.correction_history && data.correction_history.length > 0 && (
        <>
          <h2 className="text-sm font-semibold">Correction history</h2>
          {data.correction_history.map((cr) => (
            <Card key={cr.id}>
              <CardContent className="space-y-1 text-sm">
                <div className="flex items-center justify-between">
                  <span className="capitalize">{cr.type.replaceAll('_', ' ')}</span>
                  <Badge variant={cr.status === 'approved' ? 'success' : cr.status === 'rejected' ? 'danger' : 'warning'}>
                    {cr.status}
                  </Badge>
                </div>
                <p className="text-[var(--color-text-muted)]">{cr.reason}</p>
                {(cr.proposed_time_in_at || cr.proposed_time_out_at) && (
                  <p className="text-xs">
                    Proposed: {cr.proposed_time_in_at ? formatDateTime(cr.proposed_time_in_at) : '—'}
                    {' → '}
                    {cr.proposed_time_out_at ? formatDateTime(cr.proposed_time_out_at) : '—'}
                  </p>
                )}
                {cr.decision_comment && (
                  <p className="border-l-2 border-[var(--color-border)] pl-2 text-xs">
                    {cr.decision_comment}
                  </p>
                )}
              </CardContent>
            </Card>
          ))}
        </>
      )}

      <h2 className="text-sm font-semibold">Evidence</h2>
      {data.evidence.map((e) => (
        <Card key={e.action}>
          <CardContent className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium capitalize">{e.action.replace('_', ' ')}</span>
              <Badge variant={e.location_status === 'verified' ? 'success' : 'warning'}>
                {e.location_status.replace('_', ' ')}
              </Badge>
            </div>
            <a
              href={`/api/v1/attendance/${data.id}/evidence/${e.action}/image?variant=watermarked`}
              target="_blank"
              rel="noreferrer"
            >
              <img
                src={`/api/v1/attendance/${data.id}/evidence/${e.action}/image?variant=watermarked`}
                alt={`${e.action} evidence photo`}
                className="w-full rounded-[var(--radius-md)]"
                loading="lazy"
              />
            </a>
            <div className="text-xs text-[var(--color-text-muted)]">
              <p>Server time: {formatDateTime(e.server_captured_at)}</p>
              {e.gps_accuracy_m != null && <p>GPS accuracy: ±{Math.round(e.gps_accuracy_m)} m</p>}
              {e.distance_from_site_m != null && <p>Distance from site: {Math.round(e.distance_from_site_m)} m</p>}
              {e.location_exception_reason && <p>Reason: {e.location_exception_reason}</p>}
            </div>
          </CardContent>
        </Card>
      ))}

      {data.journal_id && (
        <Link to={staffView ? `/staff/journals/${data.journal_id}` : `/journals/${data.journal_id}`} className="block">
          <Button variant="secondary" fullWidth>Open daily journal</Button>
        </Link>
      )}
    </div>
  )
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between gap-3">
      <span className="shrink-0 text-[var(--color-text-muted)]">{label}</span>
      <span className="min-w-0 break-words text-right font-medium">{value}</span>
    </div>
  )
}
