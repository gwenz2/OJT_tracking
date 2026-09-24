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

const traineeSchema = z.object({
  email: z.string().email('Valid email required'),
  display_name: z.string().min(1, 'Required'),
  student_number: z.string().min(1, 'Required'),
  program: z.string().min(1, 'Required'),
  year_level: z.string().min(1, 'Required'),
  contact_number: z.string().optional(),
})
type TraineeForm = z.infer<typeof traineeSchema>

export function TraineesPage() {
  const [page, setPage] = useState(1)
  const [q, setQ] = useState('')
  const [status, setStatus] = useState('')
  const [editing, setEditing] = useState<Trainee | 'new' | null>(null)
  const [importing, setImporting] = useState(false)
  const [tempPassword, setTempPassword] = useState<{ name: string; pw: string } | null>(null)
  const toast = useToast()

  const debouncedQ = useDebounce(q, 300)
  const params = new URLSearchParams({ page: String(page), page_size: '20' })
  if (debouncedQ) params.set('q', debouncedQ)
  if (status) params.set('status', status)

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: qk.staffTrainees({ page, q: debouncedQ, status }),
    queryFn: () => api.getPaged<Trainee>(`/staff/trainees?${params}`),
  })

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h1 className="text-lg font-semibold">Trainees</h1>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => setImporting(true)}>Import CSV</Button>
          <Button onClick={() => setEditing('new')}>Add trainee</Button>
        </div>
      </div>

      <div className="flex flex-wrap gap-2">
        <Input
          placeholder="Search name, email, student no."
          value={q}
          onChange={(e) => { setQ(e.target.value); setPage(1) }}
          className="max-w-xs"
          aria-label="Search trainees"
        />
        <Select value={status} onChange={(e) => { setStatus(e.target.value); setPage(1) }} className="w-36" aria-label="Filter by status">
          <option value="">All statuses</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
          <option value="locked">Locked</option>
        </Select>
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
        <Card>
          <CardContent className="overflow-x-auto p-0">
            <table className="w-full text-sm">
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
                {data.items.map((t) => (
                  <tr key={t.id} className="border-b border-[var(--color-border)] last:border-0">
                    <td className="px-4 py-3">
                      <div className="font-medium">{t.display_name}</div>
                      <div className="text-xs text-[var(--color-text-muted)]">{t.email}</div>
                    </td>
                    <td className="px-4 py-3">{t.student_number}</td>
                    <td className="px-4 py-3 text-[var(--color-text-muted)]">{t.site_name ?? '—'}</td>
                    <td className="px-4 py-3">
                      {t.progress_percent != null ? `${Math.round(t.progress_percent)}%` : '—'}
                    </td>
                    <td className="px-4 py-3">
                      <Badge variant={t.account_status === 'active' ? 'success' : 'default'}>
                        {t.account_status}
                      </Badge>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <Button variant="ghost" size="sm" onClick={() => setEditing(t)}>Edit</Button>
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
          display_name: trainee.display_name,
          student_number: trainee.student_number,
          program: trainee.program,
          year_level: trainee.year_level,
          contact_number: trainee.contact_number,
        }
      : undefined,
  })
  const [accountStatus, setAccountStatus] = useState(trainee?.account_status ?? 'active')

  const save = useMutation({
    mutationFn: async (f: TraineeForm) => {
      if (trainee) {
        const { email: _omit, ...rest } = f // email is immutable post-create
        return api.patch<Trainee>(`/staff/trainees/${trainee.id}`, { ...rest, account_status: accountStatus })
      }
      return api.post<{ trainee: Trainee; temporary_password: string }>('/staff/trainees', f)
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
        {!trainee && (
          <div>
            <Label htmlFor="t-email">Email</Label>
            <Input id="t-email" type="email" invalid={!!errors.email} {...register('email')} />
            <FieldError>{errors.email?.message}</FieldError>
          </div>
        )}
        <div>
          <Label htmlFor="t-name">Full name</Label>
          <Input id="t-name" invalid={!!errors.display_name} {...register('display_name')} />
          <FieldError>{errors.display_name?.message}</FieldError>
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div>
            <Label htmlFor="t-student">Student number</Label>
            <Input id="t-student" invalid={!!errors.student_number} {...register('student_number')} />
            <FieldError>{errors.student_number?.message}</FieldError>
          </div>
          <div>
            <Label htmlFor="t-year">Year level</Label>
            <Input id="t-year" invalid={!!errors.year_level} {...register('year_level')} placeholder="e.g. 4" />
            <FieldError>{errors.year_level?.message}</FieldError>
          </div>
        </div>
        <div>
          <Label htmlFor="t-program">Program</Label>
          <Input id="t-program" invalid={!!errors.program} {...register('program')} placeholder="e.g. BSIT" />
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
