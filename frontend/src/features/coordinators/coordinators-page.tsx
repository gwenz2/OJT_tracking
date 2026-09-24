/**
 * Admin coordinator management — list, create (temp password shown once),
 * activate/deactivate, and trainee scope assignment.
 */
import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api, ApiError } from '@/api/client'
import { qk } from '@/api/queryKeys'
import type { Coordinator, Trainee } from '@/api/types'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input, Label } from '@/components/ui/input'
import { Pagination } from '@/components/ui/pagination'
import { LoadingState, EmptyState, ErrorState, AccessDenied } from '@/components/feedback/states'
import { useToast } from '@/components/feedback/toast'
import { formatDate } from '@/lib/utils'

const coordApi = {
  list: (page: number) => api.getPaged<Coordinator>(`/admin/coordinators?page=${page}`),
  create: (body: { email: string; display_name: string }) =>
    api.post<Coordinator & { temporary_password?: string }>('/admin/coordinators', body),
  patch: (id: string, body: { display_name?: string; account_status?: string }) =>
    api.patch<Coordinator>(`/admin/coordinators/${id}`, body),
  setScope: (id: string, traineeIds: string[]) =>
    api.put(`/admin/coordinators/${id}/trainee-scope`, { trainee_ids: traineeIds }),
}

function ScopePanel({ coord, onDone }: { coord: Coordinator; onDone: () => void }) {
  const toast = useToast()
  const qc = useQueryClient()
  const [selected, setSelected] = useState<Set<string>>(new Set())

  const { data: trainees } = useQuery({
    queryKey: qk.staffTrainees({ page_size: 200 }),
    queryFn: () => api.getPaged<Trainee>('/staff/trainees?page_size=200'),
  })

  const save = useMutation({
    mutationFn: () => coordApi.setScope(coord.id, [...selected]),
    onSuccess: () => {
      toast.success('Scope updated')
      qc.invalidateQueries({ queryKey: qk.adminCoordinators({}) })
      onDone()
    },
    onError: (e) => toast.error(e instanceof ApiError ? e.message : 'Failed'),
  })

  const toggle = (id: string) => {
    const next = new Set(selected)
    if (next.has(id)) next.delete(id)
    else next.add(id)
    setSelected(next)
  }

  return (
    <CardContent className="space-y-2 border-t border-[var(--color-border)] pt-4">
      <p className="text-sm text-[var(--color-text-muted)]">
        Select the trainees {coord.display_name} may monitor (replaces current scope).
      </p>
      <div className="max-h-56 space-y-1 overflow-y-auto">
        {(trainees?.items ?? []).map((t) => (
          <label key={t.id} className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={selected.has(t.id)}
              onChange={() => toggle(t.id)}
              className="accent-[var(--color-primary)]"
            />
            {t.display_name} <span className="text-[var(--color-text-muted)]">({t.student_number})</span>
          </label>
        ))}
      </div>
      <div className="flex gap-2">
        <Button variant="ghost" size="sm" onClick={onDone}>Cancel</Button>
        <Button size="sm" loading={save.isPending} onClick={() => save.mutate()}>Save scope</Button>
      </div>
    </CardContent>
  )
}

export function CoordinatorsPage() {
  const toast = useToast()
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [email, setEmail] = useState('')
  const [name, setName] = useState('')
  const [tempPw, setTempPw] = useState<string | null>(null)
  const [scopeFor, setScopeFor] = useState<string | null>(null)

  const { data, isLoading, error } = useQuery({
    queryKey: qk.adminCoordinators({ page }),
    queryFn: () => coordApi.list(page),
  })

  const create = useMutation({
    mutationFn: () => coordApi.create({ email, display_name: name }),
    onSuccess: (res) => {
      setTempPw(res.temporary_password ?? null)
      setEmail(''); setName('')
      qc.invalidateQueries({ queryKey: qk.adminCoordinators({}) })
      toast.success('Coordinator created')
    },
    onError: (e) => toast.error(e instanceof ApiError ? e.message : 'Create failed'),
  })

  const toggleStatus = useMutation({
    mutationFn: (c: Coordinator) =>
      coordApi.patch(c.id, { account_status: c.account_status === 'active' ? 'inactive' : 'active' }),
    onSuccess: () => qc.invalidateQueries({ queryKey: qk.adminCoordinators({}) }),
    onError: (e) => toast.error(e instanceof ApiError ? e.message : 'Update failed'),
  })

  if (error instanceof ApiError && error.status === 403) return <AccessDenied />

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Coordinators</h1>

      <Card>
        <CardHeader><CardTitle className="text-base">Add coordinator</CardTitle></CardHeader>
        <CardContent className="flex flex-wrap items-end gap-3">
          <div className="flex-1 min-w-40">
            <Label htmlFor="c-name">Name</Label>
            <Input id="c-name" value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <div className="flex-1 min-w-40">
            <Label htmlFor="c-email">Email</Label>
            <Input id="c-email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
          </div>
          <Button disabled={!email || !name} loading={create.isPending} onClick={() => create.mutate()}>
            Create
          </Button>
        </CardContent>
        {tempPw && (
          <CardContent className="pt-0">
            <p className="rounded bg-[var(--color-warning)]/10 p-2 text-sm text-[var(--color-warning)]">
              Temporary password (shown once): <code className="font-mono">{tempPw}</code>
            </p>
          </CardContent>
        )}
      </Card>

      {isLoading && <LoadingState label="Loading coordinators…" />}
      {error && !(error instanceof ApiError && error.status === 403) && (
        <ErrorState description={error instanceof ApiError ? error.message : 'Failed to load'} />
      )}
      {data && data.items.length === 0 && <EmptyState title="No coordinators" />}

      {data && data.items.length > 0 && (
        <ul className="space-y-3">
          {data.items.map((c) => (
            <Card key={c.id}>
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle className="text-base">{c.display_name}</CardTitle>
                    <p className="text-sm text-[var(--color-text-muted)]">
                      {c.email} · {c.trainee_count} trainees · since {formatDate(c.created_at)}
                    </p>
                  </div>
                  <Badge variant={c.account_status === 'active' ? 'success' : 'default'}>
                    {c.account_status}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent className="flex gap-2 pb-3">
                <Button variant="outline" size="sm"
                  onClick={() => setScopeFor(scopeFor === c.id ? null : c.id)}>
                  {scopeFor === c.id ? 'Close scope' : 'Trainee scope'}
                </Button>
                <Button variant="ghost" size="sm" loading={toggleStatus.isPending}
                  onClick={() => toggleStatus.mutate(c)}>
                  {c.account_status === 'active' ? 'Deactivate' : 'Activate'}
                </Button>
              </CardContent>
              {scopeFor === c.id && <ScopePanel coord={c} onDone={() => setScopeFor(null)} />}
            </Card>
          ))}
        </ul>
      )}

      {data && <Pagination meta={data.meta} onPage={setPage} />}
    </div>
  )
}
