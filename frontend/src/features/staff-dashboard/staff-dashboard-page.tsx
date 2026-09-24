/**
 * Staff dashboard — scope-aware metric cards + attention queues.
 * Each card links to the matching pre-filtered list (T112).
 */
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { monitoringApi } from '@/features/corrections/api'
import { qk } from '@/api/queryKeys'
import { ApiError } from '@/api/client'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { LoadingState, ErrorState, AccessDenied } from '@/components/feedback/states'
import { formatDate, formatDateTime } from '@/lib/utils'

const cards = [
  { key: 'active_trainees', label: 'Active trainees', to: '/staff/trainees' },
  { key: 'present_today', label: 'Present today', to: '/staff/attendance' },
  { key: 'no_attendance', label: 'No attendance today', to: '/staff/attendance' },
  { key: 'flagged_attendance', label: 'Flagged today', to: '/staff/attendance?status=flagged' },
  { key: 'missing_journals', label: 'Missing journals', to: '/staff/journals' },
  { key: 'near_completion', label: 'Near completion', to: '/staff/reports?report=student-progress' },
  { key: 'completed_hours', label: 'Completed hours', to: '/staff/reports?report=completion' },
] as const

export function StaffDashboardPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: qk.staffDashboard(),
    queryFn: monitoringApi.dashboard,
    refetchInterval: 60_000,
  })

  if (error instanceof ApiError && error.status === 403) return <AccessDenied />
  if (isLoading) return <LoadingState label="Loading dashboard…" />
  if (error || !data) {
    return <ErrorState description={error instanceof ApiError ? error.message : 'Failed to load'} />
  }

  return (
    <div className="space-y-6">
      <h1 className="text-lg font-semibold">Dashboard</h1>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {cards.map(({ key, label, to }) => {
          const value = data.metrics[key]
          if (key === 'no_attendance' && !data.metrics.no_attendance_configured) {
            return (
              <Card key={key}>
                <CardHeader>
                  <CardDescription>{label}</CardDescription>
                  <CardTitle className="text-2xl">—</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-xs text-[var(--color-text-subtle)]">Schedule not configured</p>
                </CardContent>
              </Card>
            )
          }
          return (
            <Link key={key} to={to}>
              <Card className="transition-colors hover:bg-[var(--color-surface-hover)]">
                <CardHeader>
                  <CardDescription>{label}</CardDescription>
                  <CardTitle className="text-2xl">{value}</CardTitle>
                </CardHeader>
              </Card>
            </Link>
          )
        })}
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        <Card>
          <CardHeader><CardTitle className="text-base">Pending corrections</CardTitle></CardHeader>
          <CardContent>
            {data.queues.pending_corrections.length === 0 && (
              <p className="text-sm text-[var(--color-text-muted)]">None pending.</p>
            )}
            <ul className="space-y-2">
              {data.queues.pending_corrections.map((c) => (
                <li key={c.id} className="text-sm">
                  <Link to="/staff/corrections" className="hover:underline">
                    {c.trainee_name}
                  </Link>
                  <span className="text-[var(--color-text-muted)]">
                    {' '}· {c.type.replaceAll('_', ' ')} · {formatDate(c.date)}
                  </span>
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle className="text-base">Recent flags</CardTitle></CardHeader>
          <CardContent>
            {data.queues.recent_flags.length === 0 && (
              <p className="text-sm text-[var(--color-text-muted)]">No flagged records.</p>
            )}
            <ul className="space-y-2">
              {data.queues.recent_flags.map((f) => (
                <li key={f.id} className="text-sm">
                  <Link to={`/staff/attendance`} className="hover:underline">{f.trainee_name}</Link>
                  <span className="text-[var(--color-text-muted)]"> · {formatDate(f.date)} </span>
                  {f.flags.map((fl) => (
                    <Badge key={fl} variant="warning" className="ml-1">{fl.replaceAll('_', ' ')}</Badge>
                  ))}
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle className="text-base">Journals to review</CardTitle></CardHeader>
          <CardContent>
            {data.queues.journals_needing_attention.length === 0 && (
              <p className="text-sm text-[var(--color-text-muted)]">Queue clear.</p>
            )}
            <ul className="space-y-2">
              {data.queues.journals_needing_attention.map((j) => (
                <li key={j.id} className="text-sm">
                  <Link to={`/staff/journals/${j.id}`} className="hover:underline">
                    {j.trainee_name}
                  </Link>
                  <span className="text-[var(--color-text-muted)]">
                    {' '}· {formatDate(j.date)}
                    {j.submitted_at && ` · ${formatDateTime(j.submitted_at)}`}
                  </span>
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
