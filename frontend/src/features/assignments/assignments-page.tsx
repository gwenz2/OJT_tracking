import { useMemo, useState } from 'react'
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
import { FilterDisclosure } from '@/components/ui/filter-disclosure'
import { LoadingState, ErrorState, EmptyState, AccessDenied } from '@/components/feedback/states'
import { useToast } from '@/components/feedback/toast'
import { formatMinutes, formatDate } from '@/lib/utils'
import { useDebounce } from '@/hooks/use-debounce'
import { Check, ChevronDown } from 'lucide-react'

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
  const pageSize = 10
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState('')
  const [site, setSite] = useState('')
  const [filtersOpen, setFiltersOpen] = useState(false)
  const [editing, setEditing] = useState<Assignment | 'new' | null>(null)
  const toast = useToast()
  const debouncedSearch = useDebounce(search.trim(), 300)

  const params = new URLSearchParams({ page: String(page), page_size: String(pageSize) })
  if (debouncedSearch) params.set('q', debouncedSearch)
  if (status) params.set('status', status)
  if (site) params.set('site_id', site)

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: qk.staffAssignments({ page, pageSize, q: debouncedSearch, status, site }),
    queryFn: () => api.getPaged<Assignment>(`/staff/assignments?${params}`),
  })

  const sites = useQuery({
    queryKey: qk.staffSites({ all: true, active: true }),
    queryFn: () => api.getPaged<Site>('/staff/sites?page=1&page_size=100&is_active=true'),
  })

  const activeFilterCount = [search, status, site].filter(Boolean).length

  return (
    <div className="space-y-4">
      <div className="flex items-start gap-2 md:items-end">
        <FilterDisclosure
          open={filtersOpen}
          onToggle={() => setFiltersOpen((open) => !open)}
          summary={activeFilterCount ? `${activeFilterCount} active filter${activeFilterCount === 1 ? '' : 's'}` : 'All assignments'}
        >
          <div className="col-span-2 min-w-0 md:min-w-64">
            <Label htmlFor="a-search">Search</Label>
            <Input
              id="a-search"
              value={search}
              onChange={(e) => { setSearch(e.target.value); setPage(1) }}
              placeholder="Trainee or site"
            />
          </div>
          <div className="min-w-0">
            <Label htmlFor="a-status-filter">Status</Label>
            <Select id="a-status-filter" value={status} onChange={(e) => { setStatus(e.target.value); setPage(1) }} className="w-full md:w-44">
              <option value="">All statuses</option>
              {['planned', 'active', 'completed', 'suspended', 'cancelled'].map((s) => (
                <option key={s} value={s}>{s}</option>
              ))}
            </Select>
          </div>
          <div className="min-w-0">
            <Label htmlFor="a-site-filter">Site</Label>
            <Select id="a-site-filter" value={site} onChange={(e) => { setSite(e.target.value); setPage(1) }} className="w-full md:w-52">
              <option value="">All sites</option>
              {sites.data?.items.map((s) => (
                <option key={s.id} value={s.id}>{s.name}</option>
              ))}
            </Select>
          </div>
        </FilterDisclosure>
        <Button className="shrink-0" onClick={() => setEditing('new')}>New assignment</Button>
      </div>

      {isLoading && <LoadingState />}
      {error instanceof ApiError && error.status === 403 && <AccessDenied />}
      {error && !(error instanceof ApiError && error.status === 403) && (
        <ErrorState description={error.message} onRetry={() => refetch()} />
      )}
      {data && data.items.length === 0 && (
        <EmptyState title="No assignments" description="Assign a trainee to a site with required hours." />
      )}
      {data && data.items.length > 0 && (
        <Card className="responsive-table-frame">
          <CardContent className="overflow-x-auto p-0">
            <table className="responsive-table w-full text-sm">
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
                    <td data-label="Trainee" data-card-primary className="px-4 py-3 font-medium">{a.trainee_name}</td>
                    <td data-label="Site" className="px-4 py-3">{a.site_name}</td>
                    <td data-label="Dates" className="px-4 py-3 text-[var(--color-text-muted)]">
                      {formatDate(a.start_date)} – {a.end_date ? formatDate(a.end_date) : 'ongoing'}
                    </td>
                    <td data-label="Hours" className="px-4 py-3">
                      {formatMinutes(a.completed_minutes)} / {formatMinutes(a.required_minutes)}
                    </td>
                    <td data-label="Status" className="px-4 py-3">
                      <Badge variant={statusVariant[a.status] ?? 'default'}>{a.status}</Badge>
                    </td>
                    <td data-label="Actions" data-card-actions className="px-4 py-3 text-right">
                      <Button variant="ghost" size="sm" onClick={() => setEditing(a)}>Edit</Button>
                    </td>
                  </tr>
                ))}
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
  const [traineeIds, setTraineeIds] = useState<string[]>([])
  const [siteId, setSiteId] = useState(assignment?.site_id ?? '')

  const trainees = useQuery({
    queryKey: qk.staffTrainees({ all: true }),
    queryFn: () => api.getPaged<Trainee>('/staff/trainees?page=1&page_size=100&status=active'),
    enabled: !assignment,
  })
  const sites = useQuery({
    queryKey: qk.staffSites({ all: true }),
    queryFn: () => api.getPaged<Site>('/staff/sites?page=1&page_size=100&is_active=true'),
  })
  const availableTrainees = useMemo(
    () => (trainees.data?.items ?? []).filter((trainee) => !trainee.assignment_id),
    [trainees.data?.items],
  )

  const save = useMutation({
    mutationFn: async (body: Record<string, unknown>) => {
      if (assignment) return api.patch<Assignment>(`/staff/assignments/${assignment.id}`, body)
      const ids = body.trainee_ids as string[]
      const { trainee_ids: _omit, ...rest } = body
      return Promise.all(ids.map((id) => api.post<Assignment>('/staff/assignments', { ...rest, trainee_id: id })))
    },
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
      if (traineeIds.length === 0) {
        setFieldErrs({ trainee_id: 'Select at least one trainee.' })
        return
      }
      body.trainee_ids = traineeIds
      body.site_id = siteId
    } else {
      body.site_id = siteId
    }
    if (!siteId) {
      setFieldErrs({ site_id: 'Select a site.' })
      return
    }
    save.mutate(body)
  }

  const fe = (k: string) => fieldErrs[k]

  return (
    <Modal open onClose={onClose} title={assignment ? 'Edit assignment' : 'New assignment'}>
      <form onSubmit={onSubmit} className="space-y-3">
        {!assignment && (
          <div>
            <MultiCombobox
              id="a-trainee-combobox"
              label="Trainee"
              values={traineeIds}
              onChange={setTraineeIds}
              invalid={!!fe('trainee_id')}
              placeholder="Search trainees"
              emptyLabel={trainees.isLoading ? 'Loading trainees...' : 'No available trainee found'}
              options={availableTrainees.map((t) => ({
                value: t.id,
                label: t.display_name,
                description: t.student_number,
              }))}
            />
            <div className="hidden">
            <Label htmlFor="a-trainee">Trainee</Label>
            <Select id="a-trainee" tabIndex={-1}>
              <option value="">Select trainee…</option>
              {availableTrainees.map((t) => (
                <option key={t.id} value={t.id}>{t.display_name} ({t.student_number})</option>
            ))}
            </Select>
            </div>
            <FieldError>{fe('trainee_id')}</FieldError>
          </div>
        )}
        <div>
          <Combobox
            id="a-site-combobox"
            label="Site"
            value={siteId}
            onChange={setSiteId}
            invalid={!!fe('site_id')}
            placeholder="Search site"
            emptyLabel={sites.isLoading ? 'Loading sites...' : 'No site found'}
            options={(sites.data?.items ?? []).map((s) => ({
              value: s.id,
              label: s.name,
              description: s.address,
            }))}
          />
          <input type="hidden" name="site_id" value={siteId} />
          <div className="hidden">
          <Label htmlFor="a-site">Site</Label>
          <Select id="a-site" tabIndex={-1} defaultValue={assignment?.site_id}>
            <option value="">Select site…</option>
            {sites.data?.items.map((s) => (
              <option key={s.id} value={s.id}>{s.name}</option>
            ))}
          </Select>
          </div>
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

function MultiCombobox({
  id,
  label,
  values,
  onChange,
  options,
  placeholder,
  emptyLabel,
  invalid,
}: {
  id: string
  label: string
  values: string[]
  onChange: (values: string[]) => void
  options: { value: string; label: string; description?: string }[]
  placeholder: string
  emptyLabel: string
  invalid?: boolean
}) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const selected = options.filter((option) => values.includes(option.value))
  const filtered = useMemo(() => {
    const needle = query.trim().toLowerCase()
    if (!needle) return options
    return options.filter((option) =>
      `${option.label} ${option.description ?? ''}`.toLowerCase().includes(needle),
    )
  }, [options, query])

  const toggle = (value: string) => {
    onChange(values.includes(value) ? values.filter((item) => item !== value) : [...values, value])
  }

  return (
    <div className="relative">
      <Label htmlFor={id}>{label}</Label>
      {selected.length > 0 && (
        <div className="mb-2 flex flex-wrap gap-1.5">
          {selected.map((option) => (
            <button
              key={option.value}
              type="button"
              onClick={() => toggle(option.value)}
              className="focus-ring rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-muted)] px-2 py-1 text-xs font-medium text-[var(--color-text)] hover:bg-[var(--color-surface-hover)]"
            >
              {option.label} ×
            </button>
          ))}
        </div>
      )}
      <div className="relative">
        <Input
          id={id}
          role="combobox"
          aria-expanded={open}
          aria-controls={`${id}-listbox`}
          aria-autocomplete="list"
          value={query}
          onFocus={() => setOpen(true)}
          onChange={(event) => {
            setQuery(event.target.value)
            setOpen(true)
          }}
          onKeyDown={(event) => {
            if (event.key === 'Escape') setOpen(false)
          }}
          onBlur={() => window.setTimeout(() => setOpen(false), 120)}
          invalid={invalid}
          placeholder={selected.length > 0 ? 'Add another trainee' : placeholder}
          autoComplete="off"
          className="pr-10"
        />
        <button
          type="button"
          onMouseDown={(event) => event.preventDefault()}
          onClick={() => {
            setOpen((next) => !next)
            setQuery('')
          }}
          className="focus-ring absolute right-0 top-0 flex size-11 items-center justify-center text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
          aria-label={`Toggle ${label.toLowerCase()} options`}
        >
          <ChevronDown size={16} aria-hidden="true" />
        </button>
      </div>
      {open && (
        <div
          id={`${id}-listbox`}
          role="listbox"
          className="absolute z-50 mt-1 max-h-56 w-full overflow-y-auto rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-surface)] shadow-[var(--shadow-lg)]"
        >
          {filtered.length === 0 && (
            <div className="px-3 py-2 text-sm text-[var(--color-text-muted)]">{emptyLabel}</div>
          )}
          {filtered.map((option) => {
            const picked = values.includes(option.value)
            return (
              <button
                key={option.value}
                type="button"
                role="option"
                aria-selected={picked}
                onMouseDown={(event) => event.preventDefault()}
                onClick={() => {
                  toggle(option.value)
                  setQuery('')
                  setOpen(true)
                }}
                className="focus-ring flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:bg-[var(--color-surface-hover)]"
              >
                <span className="flex min-w-0 flex-1 flex-col">
                  <span className="truncate font-medium">{option.label}</span>
                  {option.description && (
                    <span className="truncate text-xs text-[var(--color-text-muted)]">{option.description}</span>
                  )}
                </span>
                {picked && <Check size={16} className="shrink-0 text-[var(--color-primary)]" aria-hidden="true" />}
              </button>
            )
          })}
        </div>
      )}
    </div>
  )
}

function Combobox({
  id,
  label,
  value,
  onChange,
  options,
  placeholder,
  emptyLabel,
  invalid,
}: {
  id: string
  label: string
  value: string
  onChange: (value: string) => void
  options: { value: string; label: string; description?: string }[]
  placeholder: string
  emptyLabel: string
  invalid?: boolean
}) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const selected = options.find((option) => option.value === value)
  const filtered = useMemo(() => {
    const needle = query.trim().toLowerCase()
    if (!needle) return options
    return options.filter((option) =>
      `${option.label} ${option.description ?? ''}`.toLowerCase().includes(needle),
    )
  }, [options, query])

  return (
    <div className="relative">
      <Label htmlFor={id}>{label}</Label>
      <div className="relative">
        <Input
          id={id}
          role="combobox"
          aria-expanded={open}
          aria-controls={`${id}-listbox`}
          aria-autocomplete="list"
          value={open ? query : selected?.label ?? ''}
          onFocus={() => {
            setOpen(true)
            setQuery('')
          }}
          onChange={(event) => {
            setQuery(event.target.value)
            setOpen(true)
          }}
          onKeyDown={(event) => {
            if (event.key === 'Escape') setOpen(false)
          }}
          onBlur={() => window.setTimeout(() => setOpen(false), 120)}
          invalid={invalid}
          placeholder={placeholder}
          autoComplete="off"
          className="pr-10"
        />
        <button
          type="button"
          onMouseDown={(event) => event.preventDefault()}
          onClick={() => {
            setOpen((next) => !next)
            setQuery('')
          }}
          className="focus-ring absolute right-0 top-0 flex size-11 items-center justify-center text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
          aria-label={`Toggle ${label.toLowerCase()} options`}
        >
          <ChevronDown size={16} aria-hidden="true" />
        </button>
      </div>
      {open && (
        <div
          id={`${id}-listbox`}
          role="listbox"
          className="absolute z-50 mt-1 max-h-56 w-full overflow-y-auto rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-surface)] shadow-[var(--shadow-lg)]"
        >
          {filtered.length === 0 && (
            <div className="px-3 py-2 text-sm text-[var(--color-text-muted)]">{emptyLabel}</div>
          )}
          {filtered.map((option) => (
            <button
              key={option.value}
              type="button"
              role="option"
              aria-selected={option.value === value}
              onMouseDown={(event) => event.preventDefault()}
              onClick={() => {
                onChange(option.value)
                setQuery('')
                setOpen(false)
              }}
              className="focus-ring flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:bg-[var(--color-surface-hover)]"
            >
              <span className="flex min-w-0 flex-1 flex-col">
                <span className="truncate font-medium">{option.label}</span>
                {option.description && (
                  <span className="truncate text-xs text-[var(--color-text-muted)]">{option.description}</span>
                )}
              </span>
              {option.value === value && <Check size={16} className="shrink-0 text-[var(--color-primary)]" aria-hidden="true" />}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
