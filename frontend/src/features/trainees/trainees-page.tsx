import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { api, ApiError } from '@/api/client'
import { qk } from '@/api/queryKeys'
import { queryClient } from '@/app/query-client'
import type { Trainee } from '@/api/types'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input, Label, FieldError } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Modal } from '@/components/ui/modal'
import { Pagination } from '@/components/ui/pagination'
import { LoadingState, ErrorState, EmptyState, AccessDenied } from '@/components/feedback/states'
import { useToast } from '@/components/feedback/toast'
import { useDebounce } from '@/hooks/use-debounce'
import { ImportDialog } from './import-dialog'
import { ChevronDown, ChevronUp } from 'lucide-react'

const traineeSchema = z.object({
  email: z.string().email('Valid email required'),
  first_name: z.string().min(1, 'Required'),
  last_name: z.string().min(1, 'Required'),
  student_number: z.string().min(1, 'Required'),
  program: z.enum(['BSIT', 'BSIS', 'BSCS'], { message: 'Select a program' }),
  year_level: z.enum(['1', '2', '3', '4', '5'], { message: 'Select a year level' }),
  contact_number: z.string().optional(),
})
type TraineeForm = z.infer<typeof traineeSchema>

function splitName(name: string) {
  const parts = name.trim().split(/\s+/)
  if (parts.length <= 1) return { first_name: parts[0] ?? '', last_name: '' }
  return { first_name: parts.slice(0, -1).join(' '), last_name: parts.at(-1) ?? '' }
}

export function TraineesPage() {
  const [page, setPage] = useState(1)
  const pageSize = 10
  const [q, setQ] = useState('')
  const [status, setStatus] = useState('')
  const [editing, setEditing] = useState<Trainee | 'new' | null>(null)
  const [passwordFor, setPasswordFor] = useState<Trainee | null>(null)
  const [importing, setImporting] = useState(false)
  const [tempPassword, setTempPassword] = useState<{ name: string; pw: string } | null>(null)
  const [expandedIds, setExpandedIds] = useState<Set<string>>(() => new Set())
  const toast = useToast()

  const toggleExpanded = (id: string) => {
    setExpandedIds((current) => {
      const next = new Set(current)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const debouncedQ = useDebounce(q, 300)
  const params = new URLSearchParams({ page: String(page), page_size: String(pageSize) })
  if (debouncedQ) params.set('q', debouncedQ)
  if (status) params.set('status', status)

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: qk.staffTrainees({ page, pageSize, q: debouncedQ, status }),
    queryFn: () => api.getPaged<Trainee>(`/staff/trainees?${params}`),
  })

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
        <Input
          placeholder="Search name, email, student no."
          value={q}
          onChange={(e) => { setQ(e.target.value); setPage(1) }}
          className="sm:max-w-xs"
          aria-label="Search trainees"
        />
        <Select value={status} onChange={(e) => { setStatus(e.target.value); setPage(1) }} className="w-full sm:w-44" aria-label="Filter by status">
          <option value="">All statuses</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
          <option value="locked">Locked</option>
        </Select>
        <div className="flex gap-2 sm:ml-auto">
          <Button className="flex-1 sm:flex-none" variant="outline" onClick={() => setImporting(true)}>Import CSV</Button>
          <Button className="flex-1 sm:flex-none" onClick={() => setEditing('new')}>Add trainee</Button>
        </div>
      </div>

      {isLoading && <LoadingState />}
      {error instanceof ApiError && error.status === 403 && <AccessDenied />}
      {error && !(error instanceof ApiError && error.status === 403) && (
        <ErrorState description={error.message} onRetry={() => refetch()} />
      )}
      {data && data.items.length === 0 && (
        <EmptyState title="No trainees found" description="Add trainees individually or import a CSV roster." />
      )}
      {data && data.items.length > 0 && (
        <Card className="responsive-table-frame">
          <CardContent className="overflow-x-auto p-0">
            <table className="responsive-table w-full text-sm">
              <thead>
                <tr className="border-b border-[var(--color-border)] text-left text-[var(--color-text-muted)]">
                  <th className="px-4 py-3 font-medium">Name</th>
                  <th className="px-4 py-3 font-medium">Student no.</th>
                  <th className="px-4 py-3 font-medium">Site</th>
                  <th className="px-4 py-3 font-medium">Progress</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody>
                {data.items.map((t) => {
                  const expanded = expandedIds.has(t.id)
                  return (
                  <tr
                    key={t.id}
                    data-expanded={expanded}
                    className="mobile-card-collapsible border-b border-[var(--color-border)] last:border-0"
                  >
                    <td data-label="Name" data-card-primary className="px-4 py-3">
                      <div className="flex items-center justify-between gap-3">
                        <div className="min-w-0">
                          <div className="truncate font-medium">{t.display_name}</div>
                          <div className="mobile-card-detail-block text-xs text-[var(--color-text-muted)] md:block">{t.email}</div>
                        </div>
                        <button
                          type="button"
                          onClick={() => toggleExpanded(t.id)}
                          aria-expanded={expanded}
                          aria-label={`${expanded ? 'Collapse' : 'Expand'} details for ${t.display_name}`}
                          className="focus-ring inline-flex h-9 shrink-0 items-center gap-1 rounded-[var(--radius-md)] px-2 text-sm font-medium text-[var(--color-text-muted)] hover:bg-[var(--color-surface-hover)] hover:text-[var(--color-text)] md:hidden"
                        >
                          {expanded ? <ChevronUp size={16} aria-hidden="true" /> : <ChevronDown size={16} aria-hidden="true" />}
                          {expanded ? 'Less' : 'More'}
                        </button>
                      </div>
                    </td>
                    <td data-label="Student no." className="mobile-card-detail px-4 py-3">{t.student_number}</td>
                    <td data-label="Site" className="mobile-card-detail px-4 py-3 text-[var(--color-text-muted)]">{t.site_name ?? '—'}</td>
                    <td data-label="Progress" className="mobile-card-detail px-4 py-3">
                      {t.progress_percent != null ? `${Math.round(t.progress_percent)}%` : '—'}
                    </td>
                    <td data-label="Status" className="mobile-card-detail px-4 py-3">
                      <Badge variant={t.account_status === 'active' ? 'success' : 'default'}>
                        {t.account_status}
                      </Badge>
                    </td>
                    <td data-label="Actions" data-card-actions className="mobile-card-detail px-4 py-3 text-right">
                      <div className="flex justify-end gap-2">
                        <Button variant="ghost" size="sm" onClick={() => setPasswordFor(t)}>
                          Set password
                        </Button>
                        <Button variant="ghost" size="sm" onClick={() => setEditing(t)}>Edit</Button>
                      </div>
                    </td>
                  </tr>
                  )
                })}
              </tbody>
            </table>
          </CardContent>
        </Card>
      )}
      {data && (
        <Pagination
          meta={data.meta}
          onPage={setPage}
        />
      )}

      {editing && (
        <TraineeModal
          trainee={editing === 'new' ? null : editing}
          onClose={() => setEditing(null)}
          onCreated={(name, pw) => {
            setEditing(null)
            setTempPassword({ name, pw })
            void queryClient.invalidateQueries({ queryKey: ['staff', 'trainees'] })
          }}
          onSaved={() => {
            setEditing(null)
            toast.success('Trainee updated')
            void queryClient.invalidateQueries({ queryKey: ['staff', 'trainees'] })
          }}
        />
      )}

      {tempPassword && (
        <Modal open onClose={() => setTempPassword(null)} title="Account created">
          <p className="text-sm text-[var(--color-text-muted)]">
            Share this one-time password with <strong>{tempPassword.name}</strong>. It is shown only
            once — the trainee should change it after signing in.
          </p>
          <code className="mt-3 block rounded-[var(--radius-md)] bg-[var(--color-surface-hover)] p-3 text-center font-mono text-lg">
            {tempPassword.pw}
          </code>
          <div className="mt-4 flex justify-end">
            <Button onClick={() => setTempPassword(null)}>Done</Button>
          </div>
        </Modal>
      )}

      {passwordFor && (
        <TraineePasswordModal
          trainee={passwordFor}
          onClose={() => setPasswordFor(null)}
          onSaved={() => {
            setPasswordFor(null)
            toast.success('Trainee password updated')
          }}
        />
      )}

      {importing && (
        <ImportDialog
          onClose={() => setImporting(false)}
          onDone={(created) => {
            setImporting(false)
            toast.success(`Imported ${created} trainee${created === 1 ? '' : 's'}`)
            void queryClient.invalidateQueries({ queryKey: ['staff', 'trainees'] })
          }}
        />
      )}
    </div>
  )
}

function TraineeModal({
  trainee,
  onClose,
  onCreated,
  onSaved,
}: {
  trainee: Trainee | null
  onClose: () => void
  onCreated: (name: string, tempPassword: string) => void
  onSaved: () => void
}) {
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<TraineeForm>({
    resolver: zodResolver(traineeSchema),
    defaultValues: trainee
      ? {
          email: trainee.email,
          ...splitName(trainee.display_name),
          student_number: trainee.student_number,
          program: ['BSIT', 'BSIS', 'BSCS'].includes(trainee.program) ? trainee.program as TraineeForm['program'] : 'BSIT',
          year_level: ['1', '2', '3', '4', '5'].includes(trainee.year_level) ? trainee.year_level as TraineeForm['year_level'] : '1',
          contact_number: trainee.contact_number,
        }
      : undefined,
  })
  const [accountStatus, setAccountStatus] = useState(trainee?.account_status ?? 'active')

  const save = useMutation({
    mutationFn: async (f: TraineeForm) => {
      const body = {
        email: f.email,
        display_name: `${f.first_name.trim()} ${f.last_name.trim()}`.trim(),
        student_number: f.student_number,
        program: f.program,
        year_level: f.year_level,
        contact_number: f.contact_number,
      }
      if (trainee) {
        return api.patch<Trainee>(`/staff/trainees/${trainee.id}`, { ...body, account_status: accountStatus })
      }
      return api.post<{ trainee: Trainee; temporary_password: string }>('/staff/trainees', body)
    },
    onSuccess: (res) => {
      if (trainee) {
        onSaved()
      } else {
        const r = res as { trainee: Trainee; temporary_password: string }
        onCreated(r.trainee.display_name, r.temporary_password)
      }
    },
    onError: (e) => {
      if (e instanceof ApiError) {
        for (const [k, v] of Object.entries(e.fields)) setError(k as keyof TraineeForm, { message: v })
      }
    },
  })

  return (
    <Modal open onClose={onClose} title={trainee ? 'Edit trainee' : 'Add trainee'}>
      <form onSubmit={handleSubmit((f) => save.mutate(f))} className="space-y-3">
        <div>
          <Label htmlFor="t-email">Email</Label>
          <Input id="t-email" type="email" invalid={!!errors.email} {...register('email')} />
          <FieldError>{errors.email?.message}</FieldError>
        </div>
        <div>
          <Label htmlFor="t-first">First name</Label>
          <Input id="t-first" invalid={!!errors.first_name} {...register('first_name')} />
          <FieldError>{errors.first_name?.message}</FieldError>
        </div>
        <div>
          <Label htmlFor="t-last">Last name</Label>
          <Input id="t-last" invalid={!!errors.last_name} {...register('last_name')} />
          <FieldError>{errors.last_name?.message}</FieldError>
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div>
            <Label htmlFor="t-student">Student number</Label>
            <Input id="t-student" invalid={!!errors.student_number} {...register('student_number')} />
            <FieldError>{errors.student_number?.message}</FieldError>
          </div>
          <div>
            <Label htmlFor="t-year">Year level</Label>
            <Select id="t-year" invalid={!!errors.year_level} {...register('year_level')}>
              <option value="1">1</option>
              <option value="2">2</option>
              <option value="3">3</option>
              <option value="4">4</option>
              <option value="5">5</option>
            </Select>
            <FieldError>{errors.year_level?.message}</FieldError>
          </div>
        </div>
        <div>
          <Label htmlFor="t-program">Program</Label>
          <Select id="t-program" invalid={!!errors.program} {...register('program')}>
            <option value="BSIT">BSIT</option>
            <option value="BSIS">BSIS</option>
            <option value="BSCS">BSCS</option>
          </Select>
          <FieldError>{errors.program?.message}</FieldError>
        </div>
        <div>
          <Label htmlFor="t-contact">Contact number (optional)</Label>
          <Input id="t-contact" invalid={!!errors.contact_number} {...register('contact_number')} />
          <FieldError>{errors.contact_number?.message}</FieldError>
        </div>
        {trainee && (
          <div>
            <Label htmlFor="t-status">Account status</Label>
            <Select id="t-status" value={accountStatus} onChange={(e) => setAccountStatus(e.target.value as Trainee['account_status'])}>
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
            </Select>
          </div>
        )}
        {save.isError && save.error instanceof ApiError && Object.keys(save.error.fields).length === 0 && (
          <p role="alert" className="text-sm text-[var(--color-danger)]">{save.error.message}</p>
        )}
        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
          <Button type="submit" loading={save.isPending}>{trainee ? 'Save' : 'Create account'}</Button>
        </div>
      </form>
    </Modal>
  )
}

function TraineePasswordModal({
  trainee,
  onClose,
  onSaved,
}: {
  trainee: Trainee
  onClose: () => void
  onSaved: () => void
}) {
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')

  const save = useMutation({
    mutationFn: () => api.post(`/staff/trainees/${trainee.id}/password`, { new_password: password }),
    onSuccess: onSaved,
    onError: (e) => {
      if (e instanceof ApiError) {
        setError(e.fields.new_password ?? e.message)
      } else {
        setError('Password update failed')
      }
    },
  })

  return (
    <Modal open onClose={onClose} title="Set trainee password">
      <form
        onSubmit={(e) => {
          e.preventDefault()
          setError('')
          save.mutate()
        }}
        className="space-y-3"
      >
        <p className="text-sm text-[var(--color-text-muted)]">
          Assign a new password for <strong>{trainee.display_name}</strong>. Active trainee sessions will be signed out.
        </p>
        <div>
          <Label htmlFor="trainee-new-password">New password</Label>
          <Input
            id="trainee-new-password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            invalid={!!error}
            autoComplete="new-password"
          />
          <FieldError>{error}</FieldError>
        </div>
        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
          <Button type="submit" loading={save.isPending} disabled={password.length < 8}>Save password</Button>
        </div>
      </form>
    </Modal>
  )
}
