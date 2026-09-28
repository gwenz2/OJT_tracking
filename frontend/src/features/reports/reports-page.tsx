/**
 * Staff reports — report selector, date/site filters, results table, CSV
 * export. Scope is server-derived from the signed-in staff user.
 */
import { useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { reportsApi } from '@/features/corrections/api'
import { qk } from '@/api/queryKeys'
import { ApiError } from '@/api/client'
import { Button } from '@/components/ui/button'
import { Input, Label } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { LoadingState, EmptyState, ErrorState, AccessDenied } from '@/components/feedback/states'
import { useToast } from '@/components/feedback/toast'
import { FilterDisclosure } from '@/components/ui/filter-disclosure'

const REPORTS = [
  { value: 'daily-attendance', label: 'Daily attendance' },
  { value: 'weekly-summary', label: 'Weekly summary' },
  { value: 'student-progress', label: 'Student progress' },
  { value: 'journal-completion', label: 'Journal completion' },
  { value: 'exceptions', label: 'Exceptions' },
  { value: 'completion', label: 'Completion' },
] as const

function cellText(v: unknown): string {
  if (v == null) return '—'
  if (typeof v === 'string') {
    // ISO timestamps → readable.
    if (/^\d{4}-\d{2}-\d{2}T/.test(v)) return new Date(v).toLocaleString()
    return v
  }
  return String(v)
}

export function ReportsPage() {
  const toast = useToast()
  const [params] = useSearchParams()
  const [report, setReport] = useState(params.get('report') ?? 'daily-attendance')
  const [search, setSearch] = useState('')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [applied, setApplied] = useState({ q: '', from: '', to: '' })
  const [exporting, setExporting] = useState(false)
  const [filtersOpen, setFiltersOpen] = useState(false)

  const filters = { from: applied.from, to: applied.to }
  const { data, isLoading, error } = useQuery({
    queryKey: qk.staffReport(report, filters),
    queryFn: () => reportsApi.run(report, filters),
  })

  const exportCsv = async () => {
    setExporting(true)
    try {
      const { blob, filename } = await reportsApi.exportCsv(report, filters)
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = filename
      a.click()
      URL.revokeObjectURL(url)
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : 'Export failed')
    } finally {
      setExporting(false)
    }
  }

  if (error instanceof ApiError && error.status === 403) return <AccessDenied />

  const selectedReport = REPORTS.find((item) => item.value === report)?.label ?? 'Report'
  const visibleRows = data?.rows.filter((row) => {
    const haystack = Object.values(row).map(cellText).join(' ').toLowerCase()
    return !applied.q.trim() || haystack.includes(applied.q.trim().toLowerCase())
  }) ?? []
  const activeDateCount = [applied.q, applied.from, applied.to].filter(Boolean).length

  return (
    <div className="space-y-4">
      <div className="flex items-start gap-2 md:items-end">
        <FilterDisclosure
          open={filtersOpen}
          onToggle={() => setFiltersOpen((open) => !open)}
          summary={activeDateCount ? `${selectedReport} · Filtered` : selectedReport}
        >
        <div className="col-span-2 min-w-0">
          <Label htmlFor="r-name">Report</Label>
          <Select id="r-name" value={report} onChange={(e) => setReport(e.target.value)} className="w-full md:w-48">
            {REPORTS.map((r) => <option key={r.value} value={r.value}>{r.label}</option>)}
          </Select>
        </div>
        <div className="col-span-2 min-w-0 md:min-w-60">
          <Label htmlFor="r-search">Search</Label>
          <Input
            id="r-search"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search report rows"
          />
        </div>
        <div className="min-w-0">
          <Label htmlFor="r-from">From</Label>
          <Input id="r-from" type="date" value={from} onChange={(e) => setFrom(e.target.value)} />
        </div>
        <div className="min-w-0">
          <Label htmlFor="r-to">To</Label>
          <Input id="r-to" type="date" value={to} onChange={(e) => setTo(e.target.value)} />
        </div>
        <Button
          variant="outline"
          onClick={() => {
            setApplied({ q: search, from, to })
            setFiltersOpen(false)
          }}
        >
          Apply
        </Button>
        </FilterDisclosure>
        <Button variant="secondary" loading={exporting} onClick={exportCsv} disabled={!data || data.rows.length === 0}>
          Export CSV
        </Button>
      </div>

      {isLoading && <LoadingState label="Running report…" />}
      {error && !(error instanceof ApiError && error.status === 403) && (
        <ErrorState description={error instanceof ApiError ? error.message : 'Report failed'} />
      )}
      {data && data.rows.length === 0 && (
        <EmptyState title="No data" description="No rows match the selected filters." />
      )}
      {data && data.rows.length > 0 && visibleRows.length === 0 && (
        <EmptyState title="No matches" description="Try a different search term or date range." />
      )}

      {data && visibleRows.length > 0 && (
        <div className="responsive-table-frame overflow-x-auto rounded-lg border border-[var(--color-border)]">
          <table className="responsive-table w-full text-sm">
            <thead>
              <tr className="border-b border-[var(--color-border)] bg-[var(--color-surface-hover)] text-left">
                {data.columns.map((c) => (
                  <th key={c.key} className="px-3 py-2 font-medium whitespace-nowrap">{c.label}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-[var(--color-border)]">
              {visibleRows.map((row, i) => (
                <tr key={i} className="hover:bg-[var(--color-surface-hover)]">
                  {data.columns.map((c) => (
                    <td key={c.key} data-label={c.label} className="px-3 py-2 whitespace-nowrap">{cellText(row[c.key])}</td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
