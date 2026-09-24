import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { api, ApiError } from '@/api/client'
import { qk } from '@/api/queryKeys'
import { queryClient } from '@/app/query-client'
import type { Assignment, Site, Trainee } from '@/api/types'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input, Label, FieldError } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Modal } from '@/components/ui/modal'
import { Pagination } from '@/components/ui/pagination'
import { LoadingState, ErrorState, EmptyState, AccessDenied } from '@/components/feedback/states'
import { useToast } from '@/components/feedback/toast'
import { formatMinutes, formatDate } from '@/lib/utils'

const WEEKDAYS = [
  { v: 1, label: 'Mon' },
  { v: 2, label: 'Tue' },
  { v: 3, label: 'Wed' },
  { v: 4, label: 'Thu' },
  { v: 5, label: 'Fri' },
  { v: 6, label: 'Sat' },
  { v: 7, label: 'Sun' },
]

const statusVariant: Record<string, 'default' | 'success' | 'warning' | 'danger'> = {
  active: 'success',
  planned: 'default',
  completed: 'default',
  suspended: 'warning',
  cancelled: 'danger',
}

export function AssignmentsPage() {
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState('')
  const [editing, setEditing] = useState<Assignment | 'new' | null>(null)
  const toast = useToast()

  const params = new URLSearchParams({ page: String(page), page_size: '20' })
  if (status) params.set('status', status)

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: qk.staffAssignments({ page, status }),
    queryFn: () => api.getPaged<Assignment>(`/staff/assignments?${params}`),
  })

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h1 className="text-lg font-semibold">Assignments</h1>
        <Button onClick={() => setEditing('new')}>New assignment</Button>
      </div>

      <Select value={status} onChange={(e) => { setStatus(e.target.value); setPage(1) }} className="w-40" aria-label="Filter by status">
        <option value="">All statuses</option>
        {['planned', 'active', 'completed', 'suspended', 'cancelled'].map((s) => (
          <option key={s} value={s}>{s}</option>
        ))}
      </Select>

      {isLoading && <LoadingState />}
      {error instanceof ApiError && error.status === 403 && <AccessDenied />}
      {error && !(error instanceof ApiError && error.status === 403) && (
        <ErrorState description={error.message} onRetry={() => refetch()} />
      )}
      {data && data.items.length === 0 && (
        <EmptyState title="No assignments" description="Assign a trainee to a site with required hours." />
      )}
      {data && data.items.length > 0 && (
        <Card>
          <CardContent className="overflow-x-auto p-0">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-[var(--color-border)] text-left text-[var(--color-text-muted)]">
                  <th className="px-4 py-3 font-medium">Trainee</th>
                  <th className="px-4 py-3 font-medium">Site</th>
                  <th className="px-4 py-3 font-medium">Dates</th>
                  <th className="px-4 py-3 font-medium">Hours</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody>
                {data.items.map((a) => (
                  <tr key={a.id} className="border-b border-[var(--color-border)] last:border-0">
                    <td className="px-4 py-3 font-medium">{a.trainee_name}</td>
                    <td className="px-4 py-3">{a.site_name}</td>
                    <td className="px-4 py-3 text-[var(--color-text-muted)]">
                      {formatDate(a.start_date)} – {a.end_date ? formatDate(a.end_date) : 'ongoing'}
                    </td>
                    <td className="px-4 py-3">
                      {formatMinutes(a.completed_minutes)} / {formatMinutes(a.required_minutes)}
                    </td>
                    <td className="px-4 py-3">
                      <Badge variant={statusVariant[a.status] ?? 'default'}>{a.status}</Badge>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <Button variant="ghost" size="sm" onClick={() => setEditing(a)}>Edit</Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </CardContent>
        </Card>
      )}
      {data && <Pagination meta={data.meta} onPage={setPage} />}

      {editing && (
        <AssignmentModal
          assignment={editing === 'new' ? null : editing}
          onClose={() => setEditing(null)}
          onSaved={(msg) => {
            setEditing(null)
            toast.success(msg)
            void queryClient.invalidateQueries({ queryKey: ['staff', 'assignments'] })
          }}
        />
      )}
    </div>
  )
}

function AssignmentModal({
  assignment,
  onClose,
  onSaved,
}: {
  assignment: Assignment | null
  onClose: () => void
  onSaved: (msg: string) => void
}) {
  const [formError, setFormError] = useState('')
  const [fieldErrs, setFieldErrs] = useState<Record<string, string>>({})
  const [weekdays, setWeekdays] = useState<number[]>(assignment?.expected_weekdays ?? [1, 2, 3, 4, 5])
  const [breakRule, setBreakRule] = useState<'none' | 'fixed_after_threshold'>(
    assignment?.break_rule_type ?? 'none',
  )
  const [status, setStatus] = useState<Assignment['status']>(assignment?.status ?? 'active')

  const trainees = useQuery({
    queryKey: qk.staffTrainees({ all: true }),
    queryFn: () => api.getPaged<Trainee>('/staff/trainees?page=1&page_size=100&status=active'),
    enabled: !assignment,
  })
  const sites = useQuery({
    queryKey: qk.staffSites({ all: true }),
    queryFn: () => api.getPaged<Site>('/staff/sites?page=1&page_size=100&is_active=true'),
  })

  const save = useMutation({
    mutationFn: async (body: Record<string, unknown>) =>
      assignment
        ? api.patch<Assignment>(`/staff/assignments/${assignment.id}`, body)
        : api.post<Assignment>('/staff/assignments', body),
    onSuccess: () => onSaved(assignment ? 'Assignment updated' : 'Assignment created'),
    onError: (e) => {
      if (e instanceof ApiError) {
        setFieldErrs(e.fields)
        if (Object.keys(e.fields).length === 0) setFormError(e.message)
      } else {
        setFormError('Request failed')
      }
    },
  })

  function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setFormError('')
    setFieldErrs({})
    const fd = new FormData(e.currentTarget)
    const body: Record<string, unknown> = {
      start_date: fd.get('start_date'),
      required_minutes: Math.round(Number(fd.get('required_hours')) * 60),
      expected_weekdays: weekdays,
      break_rule:
        breakRule === 'none'
          ? { type: 'none' }
          : {
              type: 'fixed_after_threshold',
              threshold_minutes: Math.round(Number(fd.get('break_threshold_hours')) * 60),
              deduction_minutes: Math.round(Number(fd.get('break_deduction_minutes'))),
            },
      status,
    }
    const end = String(fd.get('end_date') ?? '')
    if (end) body.end_date = end
    if (!assignment) {
      body.trainee_id = fd.get('trainee_id')
      body.site_id = fd.get('site_id')
    } else {
      body.site_id = fd.get('site_id')
    }
    save.mutate(body)
  }

  const fe = (k: string) => fieldErrs[k]

  return (
    <Modal open onClose={onClose} title={assignment ? 'Edit assignment' : 'New assignment'}>
      <form onSubmit={onSubmit} className="space-y-3">
        {!assignment && (
          <div>
            <Label htmlFor="a-trainee">Trainee</Label>
            <Select id="a-trainee" name="trainee_id" required invalid={!!fe('trainee_id')}>
              <option value="">Select trainee…</option>
              {trainees.data?.items.map((t) => (
                <option key={t.id} value={t.id}>{t.display_name} ({t.student_number})</option>
              ))}
            </Select>
            <FieldError>{fe('trainee_id')}</FieldError>
          </div>
        )}
        <div>
          <Label htmlFor="a-site">Site</Label>
          <Select id="a-site" name="site_id" required defaultValue={assignment?.site_id} invalid={!!fe('site_id')}>
            <option value="">Select site…</option>
            {sites.data?.items.map((s) => (
              <option key={s.id} value={s.id}>{s.name}</option>
            ))}
          </Select>
          <FieldError>{fe('site_id')}</FieldError>
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div>
            <Label htmlFor="a-start">Start date</Label>
            <Input id="a-start" name="start_date" type="date" required defaultValue={assignment?.start_date} invalid={!!fe('start_date')} />
            <FieldError>{fe('start_date')}</FieldError>
          </div>
          <div>
            <Label htmlFor="a-end">End date (optional)</Label>
            <Input id="a-end" name="end_date" type="date" defaultValue={assignment?.end_date ?? ''} />
          </div>
        </div>
        <div>
          <Label htmlFor="a-hours">Required hours</Label>
          <Input
            id="a-hours"
            name="required_hours"
            type="number"
            step="0.5"
            min="1"
            required
            defaultValue={assignment ? assignment.required_minutes / 60 : ''}
            invalid={!!fe('required_minutes')}
          />
          <FieldError>{fe('required_minutes')}</FieldError>
        </div>
        <fieldset>
          <legend className="mb-1.5 text-sm font-medium">Expected days</legend>
          <div className="flex flex-wrap gap-2">
            {WEEKDAYS.map((d) => (
              <label
                key={d.v}
                className={`cursor-pointer rounded-full border px-3 py-1 text-xs font-medium ${
                  weekdays.includes(d.v)
                    ? 'border-[var(--color-primary)] bg-[var(--color-primary)]/10 text-[var(--color-primary)]'
                    : 'border-[var(--color-border)] text-[var(--color-text-muted)]'
                }`}
              >
                <input
                  type="checkbox"
                  className="sr-only"
                  checked={weekdays.includes(d.v)}
                  onChange={() =>
                    setWeekdays((w) => (w.includes(d.v) ? w.filter((x) => x !== d.v) : [...w, d.v].sort()))
                  }
                />
                {d.label}
              </label>
            ))}
          </div>
          <FieldError>{fe('expected_weekdays')}</FieldError>
        </fieldset>
        <div className="grid grid-cols-3 items-end gap-3">
          <div>
            <Label htmlFor="a-break">Break rule</Label>
            <Select
              id="a-break"
              value={breakRule}
              onChange={(e) => setBreakRule(e.target.value as 'none' | 'fixed_after_threshold')}
            >
              <option value="none">None</option>
              <option value="fixed_after_threshold">After threshold</option>
            </Select>
          </div>
          {breakRule === 'fixed_after_threshold' && (
            <>
              <div>
                <Label htmlFor="a-bt">Threshold (h)</Label>
                <Input
                  id="a-bt"
                  name="break_threshold_hours"
                  type="number"
                  step="0.5"
                  min="0.5"
                  defaultValue={assignment?.break_threshold_minutes ? assignment.break_threshold_minutes / 60 : 4}
                />
                <FieldError>{fe('break_rule')}</FieldError>
              </div>
              <div>
                <Label htmlFor="a-bd">Deduct (min)</Label>
                <Input
                  id="a-bd"
                  name="break_deduction_minutes"
                  type="number"
                  min="0"
                  defaultValue={assignment?.break_deduction_minutes || 30}
                />
              </div>
            </>
          )}
        </div>
        <div>
          <Label htmlFor="a-status">Status</Label>
          <Select
            id="a-status"
            value={status}
            onChange={(e) => setStatus(e.target.value as Assignment['status'])}
          >
            {['planned', 'active', 'completed', 'suspended', 'cancelled'].map((s) => (
              <option key={s} value={s}>{s}</option>
            ))}
          </Select>
        </div>
        {formError && <p role="alert" className="text-sm text-[var(--color-danger)]">{formError}</p>}
        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
          <Button type="submit" loading={save.isPending}>Save</Button>
        </div>
      </form>
    </Modal>
  )
}
