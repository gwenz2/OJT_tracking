import type { ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import {
  AlertTriangle,
  ArrowRight,
  BookOpenCheck,
  CalendarX2,
  CheckCircle2,
  FileDown,
  MapPin,
  UserCog,
  UserPlus,
} from 'lucide-react'
import { qk } from '@/api/queryKeys'
import { ApiError } from '@/api/client'
import { Card, CardContent } from '@/components/ui/card'
import { AccessDenied, ErrorState, LoadingState } from '@/components/feedback/states'
import { useSession } from '@/features/auth/session'
import { monitoringApi } from '@/features/corrections/api'
import { formatDate } from '@/lib/utils'

const metrics = [
  {
    key: 'present_today',
    label: 'Present today',
    note: 'Attendance recorded',
    to: '/staff/attendance',
    icon: CheckCircle2,
    tone: 'text-[var(--color-success)]',
  },
  {
    key: 'no_attendance',
    label: 'No record',
    note: 'May need follow-up',
    to: '/staff/attendance',
    icon: CalendarX2,
    tone: 'text-[var(--color-warning)]',
  },
  {
    key: 'missing_journals',
    label: 'Pending journals',
    note: 'Awaiting submission',
    to: '/staff/journals',
    icon: BookOpenCheck,
    tone: 'text-[var(--color-warning)]',
  },
  {
    key: 'flagged_attendance',
    label: 'Attendance flags',
    note: 'Needs verification',
    to: '/staff/attendance?status=flagged',
    icon: AlertTriangle,
    tone: 'text-[var(--color-danger)]',
  },
] as const

export function StaffDashboardPage() {
  const { user } = useSession()
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

  const attentionCount =
    data.queues.pending_corrections.length +
    data.queues.recent_flags.length +
    data.queues.journals_needing_attention.length

  return (
    <div className="mx-auto max-w-[1180px] space-y-8">
      <p className="text-sm text-[var(--color-text-muted)]">{formatDate(new Date())}</p>

      <section aria-labelledby="today-summary">
        <div className="mb-3 flex items-center justify-between">
          <h2 id="today-summary" className="text-sm font-semibold">Today</h2>
          <span className="text-xs text-[var(--color-text-subtle)]">Updates every minute</span>
        </div>
        <Card className="overflow-hidden shadow-none">
          <CardContent className="grid p-0 sm:grid-cols-2 lg:grid-cols-4">
            {metrics.map(({ key, label, note, to, icon: Icon, tone }, index) => {
              const unavailable = key === 'no_attendance' && !data.metrics.no_attendance_configured
              return (
                <Link
                  key={key}
                  to={to}
                  className={`group p-5 text-[var(--color-text)] transition-colors hover:bg-[var(--color-surface-hover)] ${
                    index > 0 ? 'border-t border-[var(--color-border)] sm:border-l lg:border-t-0' : ''
                  } ${index === 2 ? 'sm:border-l-0 lg:border-l' : ''}`}
                >
                  <div className="flex items-center justify-between gap-3">
                    <span className="text-sm font-medium text-[var(--color-text-muted)]">{label}</span>
                    <Icon size={18} className={tone} aria-hidden="true" />
                  </div>
                  <p className="mt-3 text-2xl font-bold tabular-nums text-[var(--color-text)]">
                    {unavailable ? '—' : data.metrics[key]}
                  </p>
                  <p className="mt-1 text-xs text-[var(--color-text-subtle)]">
                    {unavailable ? 'Schedule not configured' : note}
                  </p>
                </Link>
              )
            })}
          </CardContent>
        </Card>
      </section>

      <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1.65fr)_minmax(280px,.8fr)]">
        <section aria-labelledby="attention-heading">
          <div className="mb-3 flex items-end justify-between gap-4">
            <div>
              <h2 id="attention-heading" className="text-sm font-semibold">Needs attention</h2>
              <p className="mt-1 text-xs text-[var(--color-text-muted)]">
                {attentionCount === 0 ? 'Your work queue is clear.' : `${attentionCount} item${attentionCount === 1 ? '' : 's'} to review`}
              </p>
            </div>
            <Link to="/staff/attendance" className="inline-flex items-center gap-1 text-sm font-semibold text-[var(--color-primary)] hover:underline">
              View attendance <ArrowRight size={15} />
            </Link>
          </div>

          <Card className="shadow-none">
            <CardContent className="divide-y divide-[var(--color-border)] p-0 px-5">
              {attentionCount === 0 && (
                <div className="flex items-center gap-3 py-8">
                  <CheckCircle2 size={20} className="text-[var(--color-success)]" />
                  <div>
                    <p className="text-sm font-semibold">Nothing needs review</p>
                    <p className="mt-1 text-xs text-[var(--color-text-muted)]">New issues will appear here.</p>
                  </div>
                </div>
              )}
              {data.queues.pending_corrections.map((item) => (
                <AttentionRow
                  key={`correction-${item.id}`}
                  icon={AlertTriangle}
                  tone="warning"
                  name={item.trainee_name}
                  detail={`${item.type.replaceAll('_', ' ')} · ${formatDate(item.date)}`}
                  label="Correction"
                  to="/staff/corrections"
                />
              ))}
              {data.queues.recent_flags.map((item) => (
                <AttentionRow
                  key={`flag-${item.id}`}
                  icon={AlertTriangle}
                  tone="danger"
                  name={item.trainee_name}
                  detail={`${item.flags.map((flag) => flag.replaceAll('_', ' ')).join(', ')} · ${formatDate(item.date)}`}
                  label="Flagged"
                  to="/staff/attendance"
                />
              ))}
              {data.queues.journals_needing_attention.map((item) => (
                <AttentionRow
                  key={`journal-${item.id}`}
                  icon={BookOpenCheck}
                  tone="warning"
                  name={item.trainee_name}
                  detail={`Journal for ${formatDate(item.date)}`}
                  label="Review"
                  to={`/staff/journals/${item.id}`}
                />
              ))}
            </CardContent>
          </Card>
        </section>

        <aside className="space-y-6">
          <DashboardSection title="Program overview">
            <dl className="divide-y divide-[var(--color-border)] text-sm">
              <Overview label="Active trainees" value={data.metrics.active_trainees} />
              <Overview label="Near completion" value={data.metrics.near_completion} />
              <Overview label="Completed hours" value={data.metrics.completed_hours} />
            </dl>
          </DashboardSection>

          <DashboardSection title="Shortcuts">
            <nav className="divide-y divide-[var(--color-border)]" aria-label="Dashboard shortcuts">
              <QuickLink to="/staff/trainees" icon={UserPlus}>Add trainee</QuickLink>
              <QuickLink to="/staff/assignments" icon={MapPin}>Assign OJT site</QuickLink>
              <QuickLink to="/staff/reports" icon={FileDown}>Export report</QuickLink>
              {user?.role === 'admin' && (
                <QuickLink to="/staff/coordinators" icon={UserCog}>Manage coordinators</QuickLink>
              )}
            </nav>
          </DashboardSection>
        </aside>
      </div>
    </div>
  )
}

function DashboardSection({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section>
      <h2 className="mb-3 text-sm font-semibold">{title}</h2>
      <Card className="shadow-none">
        <CardContent className="px-5 py-2">{children}</CardContent>
      </Card>
    </section>
  )
}

function AttentionRow({
  icon: Icon,
  tone,
  name,
  detail,
  label,
  to,
}: {
  icon: typeof AlertTriangle
  tone: 'warning' | 'danger'
  name: string
  detail: string
  label: string
  to: string
}) {
  return (
    <Link to={to} className="group flex items-center gap-3 py-4 text-sm">
      <Icon
        size={18}
        className={tone === 'danger' ? 'shrink-0 text-[var(--color-danger)]' : 'shrink-0 text-[var(--color-warning)]'}
        aria-hidden="true"
      />
      <div className="min-w-0 flex-1">
        <p className="font-semibold group-hover:underline">{name}</p>
        <p className="mt-1 truncate text-xs capitalize text-[var(--color-text-muted)]">{detail}</p>
      </div>
      <span className="hidden text-xs text-[var(--color-text-subtle)] sm:block">{label}</span>
      <ArrowRight size={15} className="shrink-0 text-[var(--color-text-subtle)]" aria-hidden="true" />
    </Link>
  )
}

function QuickLink({ to, icon: Icon, children }: { to: string; icon: typeof UserPlus; children: ReactNode }) {
  return (
    <Link to={to} className="flex items-center gap-3 py-3 text-sm font-medium hover:text-[var(--color-primary)]">
      <Icon size={17} aria-hidden="true" />
      {children}
      <ArrowRight size={15} className="ml-auto text-[var(--color-text-subtle)]" aria-hidden="true" />
    </Link>
  )
}

function Overview({ label, value }: { label: string; value: number }) {
  return (
    <div className="flex items-center justify-between py-3">
      <dt className="text-[var(--color-text-muted)]">{label}</dt>
      <dd className="font-semibold tabular-nums">{value}</dd>
    </div>
  )
}
