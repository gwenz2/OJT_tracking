import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { api, ApiError } from '@/api/client'
import { qk } from '@/api/queryKeys'
import type { TraineeDashboard } from '@/api/types'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Modal } from '@/components/ui/modal'
import { LoadingState, ErrorState, AccessDenied } from '@/components/feedback/states'
import { AttendanceCaptureFlow } from '@/features/attendance/AttendanceCaptureFlow'
import { formatMinutes, formatDate, formatTime } from '@/lib/utils'
import { ChevronRight, FileEdit, LogIn, LogOut, BookOpen } from 'lucide-react'

/**
 * Trainee home: assignment progress + the single next action the attendance
 * workflow wants (time in, time out, complete journal, …).
 */
export function TodayPage() {
  const location = useLocation()
  const navigate = useNavigate()
  const initialCapture = (location.state as { capture?: 'time_in' | 'time_out' } | null)?.capture ?? null
  const [capture, setCapture] = useState<'time_in' | 'time_out' | null>(initialCapture)

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: qk.traineeDashboard,
    queryFn: () => api.get<TraineeDashboard>('/trainee/dashboard'),
    retry: false,
  })

  if (isLoading) return <LoadingState />
  if (error) {
    if (error instanceof ApiError && error.status === 403) return <AccessDenied />
    return <ErrorState description={error.message} onRetry={() => refetch()} />
  }
  if (!data) return null

  const { assignment, today } = data

  return (
    <div className="space-y-4 p-4">
      <div>
        <h1 className="text-lg font-semibold">Today</h1>
        <p className="text-sm text-[var(--color-text-muted)]">{formatDate(new Date())}</p>
      </div>

      {assignment ? (
        <Card>
          <CardContent className="space-y-2">
            <div className="flex items-center justify-between gap-2">
              <span className="min-w-0 break-words text-sm font-medium">{assignment.site.name}</span>
              <Badge className="shrink-0">{assignment.progress_percent}%</Badge>
            </div>
            <div
              className="h-2 overflow-hidden rounded-full bg-[var(--color-surface-hover)]"
              role="progressbar"
              aria-valuenow={assignment.progress_percent}
              aria-valuemin={0}
              aria-valuemax={100}
            >
              <div
                className="h-full bg-[var(--color-primary)] transition-all"
                style={{ width: `${assignment.progress_percent}%` }}
              />
            </div>
            <p className="text-xs text-[var(--color-text-muted)]">
              {formatMinutes(assignment.completed_minutes)} of {formatMinutes(assignment.required_minutes)} ·{' '}
              {formatMinutes(assignment.remaining_minutes)} remaining
            </p>
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardContent>
            <p className="text-sm text-[var(--color-text-muted)]">
              No active OJT assignment yet. Contact your coordinator to get placed at a site.
            </p>
          </CardContent>
        </Card>
      )}

      <AttendanceCard today={today} onCapture={setCapture} />

      <Link
        to="/corrections"
        className="flex min-h-[44px] items-center justify-between gap-2 rounded-[var(--radius-md)] px-2 text-sm font-medium text-[var(--color-text-muted)] hover:text-[var(--color-text)] focus-ring"
      >
        <span className="flex items-center gap-2">
          <FileEdit size={16} />
          Wrong recorded time? Request a correction
        </span>
        <ChevronRight size={16} aria-hidden="true" />
      </Link>

      <Modal
        open={capture != null}
        onClose={() => setCapture(null)}
        title={capture === 'time_in' ? 'Time In' : 'Time Out'}
      >
        {capture && (
          <AttendanceCaptureFlow
            key={capture}
            action={capture}
            attendanceId={today.attendance_id ?? undefined}
            onDone={() => {
              setCapture(null)
              void refetch()
              if (capture === 'time_out' && today.journal_id) {
                navigate(`/journals/${today.journal_id}`)
              }
            }}
            onCancel={() => setCapture(null)}
          />
        )}
      </Modal>
    </div>
  )
}

function AttendanceCard({
  today,
  onCapture,
}: {
  today: TraineeDashboard['today']
  onCapture: (a: 'time_in' | 'time_out') => void
}) {
  const navigate = useNavigate()
  return (
    <Card>
      <CardContent className="space-y-3">
        {today.attendance_status && (
          <div className="space-y-2">
            <div className="flex items-center justify-between gap-2">
              <span className="text-sm text-[var(--color-text-muted)]">Attendance</span>
              <Badge
                variant={today.attendance_status === 'valid' ? 'success' : 'default'}
                className="capitalize"
              >
                {today.attendance_status}
              </Badge>
            </div>
            {(today.time_in_at || today.time_out_at) && (
              <div className="flex gap-6 text-sm">
                <span>
                  <span className="text-[var(--color-text-muted)]">In </span>
                  <span className="font-medium tabular-nums">{formatTime(today.time_in_at)}</span>
                </span>
                <span>
                  <span className="text-[var(--color-text-muted)]">Out </span>
                  <span className="font-medium tabular-nums">{formatTime(today.time_out_at)}</span>
                </span>
              </div>
            )}
            {today.journal_status && (
              <p className="text-xs capitalize text-[var(--color-text-muted)]">
                Journal: {today.journal_status.replace('_', ' ')}
              </p>
            )}
          </div>
        )}
        <NextAction today={today} onCapture={onCapture} onJournal={(id) => navigate(`/journals/${id}`)} />
      </CardContent>
    </Card>
  )
}

function NextAction({
  today,
  onCapture,
  onJournal,
}: {
  today: TraineeDashboard['today']
  onCapture: (a: 'time_in' | 'time_out') => void
  onJournal: (id: string) => void
}) {
  const navigate = useNavigate()
  switch (today.next_action) {
    case 'time_in':
      return (
        <Button fullWidth size="lg" onClick={() => onCapture('time_in')}>
          <LogIn size={18} /> Time In
        </Button>
      )
    case 'time_out':
      return (
        <Button fullWidth size="lg" onClick={() => onCapture('time_out')}>
          <LogOut size={18} /> Time Out
        </Button>
      )
    case 'complete_journal':
      return (
        <Button fullWidth size="lg" onClick={() => today.journal_id && onJournal(today.journal_id)}>
          <BookOpen size={18} /> Complete daily journal
        </Button>
      )
    case 'revise_journal':
      return (
        <Button
          fullWidth
          size="lg"
          variant="secondary"
          onClick={() => today.journal_id && onJournal(today.journal_id)}
        >
          <BookOpen size={18} /> Revise journal
        </Button>
      )
    case 'view_summary':
      return (
        <div className="space-y-2">
          <p className="text-center text-sm text-[var(--color-text-muted)]">All done for today.</p>
          {today.attendance_id && (
            <Button
              variant="outline"
              fullWidth
              onClick={() => navigate(`/attendance/${today.attendance_id}`)}
            >
              View today's record
            </Button>
          )}
        </div>
      )
    case 'contact_coordinator':
      return null
  }
}
