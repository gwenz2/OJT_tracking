import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { AlertTriangle, ArrowRight, BookOpenCheck, CalendarX2, CheckCircle2, FileDown, MapPin, UserPlus } from 'lucide-react'
import { monitoringApi } from '@/features/corrections/api'
import { qk } from '@/api/queryKeys'
import { ApiError } from '@/api/client'
import { Card, CardContent } from '@/components/ui/card'
import { LoadingState, ErrorState, AccessDenied } from '@/components/feedback/states'
import { formatDate } from '@/lib/utils'

const metrics = [
  { key: 'present_today', label: 'Present today', note: 'Recorded attendance', to: '/staff/attendance', icon: CheckCircle2, accent: 'bg-[var(--color-success)]' },
  { key: 'no_attendance', label: 'No record', note: 'Requires follow-up', to: '/staff/attendance', icon: CalendarX2, accent: 'bg-[#f4c430]' },
  { key: 'missing_journals', label: 'Pending journals', note: 'Awaiting submission', to: '/staff/journals', icon: BookOpenCheck, accent: 'bg-[var(--color-warning)]' },
  { key: 'flagged_attendance', label: 'Attendance flags', note: 'Location or evidence', to: '/staff/attendance?status=flagged', icon: AlertTriangle, accent: 'bg-[var(--color-danger)]' },
] as const

export function StaffDashboardPage() {
  const { data, isLoading, error } = useQuery({ queryKey: qk.staffDashboard(), queryFn: monitoringApi.dashboard, refetchInterval: 60_000 })
  if (error instanceof ApiError && error.status === 403) return <AccessDenied />
  if (isLoading) return <LoadingState label="Loading dashboard…" />
  if (error || !data) return <ErrorState description={error instanceof ApiError ? error.message : 'Failed to load'} />
  return (
    <div className="mx-auto max-w-[1200px] space-y-7">
      <div className="flex items-end justify-between gap-4"><div><h1 className="text-3xl font-extrabold">Dashboard</h1><p className="mt-1 text-sm text-[var(--color-text-muted)]">Attendance and journal activity for {formatDate(new Date())}</p></div><span className="hidden text-xs font-medium text-[var(--color-success)] sm:inline">Updated automatically</span></div>
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {metrics.map(({ key, label, note, to, icon: Icon, accent }) => { const unavailable = key === 'no_attendance' && !data.metrics.no_attendance_configured; return <Link key={key} to={to}><Card className="relative overflow-hidden shadow-none transition-colors hover:border-[var(--color-border-strong)]"><span className={`absolute inset-y-5 left-0 w-1 ${accent}`} /><CardContent className="p-5 pl-6"><div className="flex items-center justify-between"><span className="text-sm font-medium text-[var(--color-text-muted)]">{label}</span><Icon size={18} className="text-[var(--color-text-subtle)]" /></div><div className="mt-2 text-3xl font-extrabold">{unavailable ? '—' : data.metrics[key]}</div><p className="mt-2 text-xs text-[var(--color-text-muted)]">{unavailable ? 'Schedule not configured' : note}</p></CardContent></Card></Link> })}
      </div>
      <div className="grid gap-5 lg:grid-cols-[minmax(0,1.7fr)_minmax(280px,.8fr)]">
        <Card className="shadow-none"><CardContent className="p-0"><div className="flex items-center justify-between border-b border-[var(--color-border)] px-5 py-4"><div><h2 className="font-bold">Needs attention</h2><p className="mt-1 text-xs text-[var(--color-text-muted)]">Items that may require coordinator action</p></div><Link to="/staff/attendance" className="flex items-center gap-1 text-sm font-semibold text-[var(--color-primary)]">View all <ArrowRight size={15} /></Link></div><div className="divide-y divide-[var(--color-border)] px-5">
          {data.queues.pending_corrections.length === 0 && data.queues.recent_flags.length === 0 && data.queues.journals_needing_attention.length === 0 && <p className="py-8 text-sm text-[var(--color-text-muted)]">Nothing needs attention right now.</p>}
          {data.queues.pending_corrections.map((x) => <AttentionRow key={x.id} icon={AlertTriangle} tone="warning" name={x.trainee_name} detail={`${x.type.replaceAll('_', ' ')} · ${formatDate(x.date)}`} label="Correction" to="/staff/corrections" />)}
          {data.queues.recent_flags.map((x) => <AttentionRow key={x.id} icon={AlertTriangle} tone="danger" name={x.trainee_name} detail={`${x.flags.map((flag) => flag.replaceAll('_', ' ')).join(', ')} · ${formatDate(x.date)}`} label="Flagged" to="/staff/attendance" />)}
          {data.queues.journals_needing_attention.map((x) => <AttentionRow key={x.id} icon={BookOpenCheck} tone="warning" name={x.trainee_name} detail={`Journal for ${formatDate(x.date)}`} label="Review" to={`/staff/journals/${x.id}`} />)}
        </div></CardContent></Card>
        <div className="space-y-5"><Card className="shadow-none"><CardContent className="p-5"><h2 className="font-bold">Quick actions</h2><div className="mt-3 divide-y divide-[var(--color-border)]"><QuickLink to="/staff/trainees" icon={UserPlus}>Add trainee</QuickLink><QuickLink to="/staff/assignments" icon={MapPin}>Assign OJT site</QuickLink><QuickLink to="/staff/reports" icon={FileDown}>Export report</QuickLink></div></CardContent></Card><Card className="shadow-none"><CardContent className="p-5"><h2 className="font-bold">Program overview</h2><dl className="mt-3 divide-y divide-[var(--color-border)] text-sm"><Overview label="Active trainees" value={data.metrics.active_trainees} /><Overview label="Near completion" value={data.metrics.near_completion} /><Overview label="Completed hours" value={data.metrics.completed_hours} /></dl></CardContent></Card></div>
      </div>
    </div>
  )
}

function AttentionRow({ icon: Icon, tone, name, detail, label, to }: { icon: typeof AlertTriangle; tone: 'warning' | 'danger'; name: string; detail: string; label: string; to: string }) {
  return <div className="flex items-center gap-3 py-4 text-sm"><Icon size={18} className={tone === 'danger' ? 'shrink-0 text-[var(--color-danger)]' : 'shrink-0 text-[var(--color-warning)]'} /><div className="min-w-0 flex-1"><Link to={to} className="font-semibold hover:underline">{name}</Link><p className="mt-1 truncate text-xs capitalize text-[var(--color-text-muted)]">{detail}</p></div><span className="text-xs text-[var(--color-text-subtle)]">{label}</span></div>
}
function QuickLink({ to, icon: Icon, children }: { to: string; icon: typeof UserPlus; children: ReactNode }) { return <Link to={to} className="flex items-center gap-3 py-3 text-sm font-semibold hover:text-[var(--color-primary)]"><Icon size={18} />{children}<ArrowRight size={15} className="ml-auto" /></Link> }
function Overview({ label, value }: { label: string; value: number }) { return <div className="flex justify-between py-3"><dt className="text-[var(--color-text-muted)]">{label}</dt><dd className="font-bold">{value}</dd></div> }
