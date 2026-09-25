/**
 * Admin institution settings — timezone, thresholds, retention.
 * Warns while retention_days is unset (institution must approve a policy).
 */
import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { settingsApi } from '@/features/corrections/api'
import { qk } from '@/api/queryKeys'
import { ApiError } from '@/api/client'
import type { InstitutionSettings } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input, Label, FieldError } from '@/components/ui/input'
import { LoadingState, ErrorState, AccessDenied } from '@/components/feedback/states'
import { useToast } from '@/components/feedback/toast'

export function SettingsPage() {
  const qc = useQueryClient()
  const toast = useToast()
  const [form, setForm] = useState<Partial<InstitutionSettings> | null>(null)

  const { data, isLoading, error } = useQuery({
    queryKey: qk.adminSettings,
    queryFn: settingsApi.get,
  })

  const save = useMutation({
    mutationFn: (body: Partial<InstitutionSettings>) => settingsApi.patch(body),
    onSuccess: () => {
      toast.success('Settings saved')
      qc.invalidateQueries({ queryKey: qk.adminSettings })
    },
    onError: (e) => toast.error(e instanceof ApiError ? e.message : 'Save failed'),
  })

  if (error instanceof ApiError && error.status === 403) return <AccessDenied />
  if (isLoading) return <LoadingState label="Loading settings…" />
  if (error || !data) {
    return <ErrorState description={error instanceof ApiError ? error.message : 'Failed to load'} />
  }

  const f = form ?? data
  const set = <K extends keyof InstitutionSettings>(k: K, v: InstitutionSettings[K]) =>
    setForm({ ...f, [k]: v })

  const num = (v: string) => (v === '' ? undefined : Number(v))

  return (
    <div className="mx-auto max-w-xl space-y-4">
      {f.retention_days == null && (
        <Card className="border-[var(--color-warning)]/50">
          <CardContent className="pt-4 text-sm text-[var(--color-warning)]">
            No evidence retention policy is set. The institution should approve a
            retention duration before go-live.
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Attendance policy</CardTitle>
          <CardDescription>Applied server-side to all time calculations.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <div>
            <Label htmlFor="s-tz">Timezone (IANA)</Label>
            <Input id="s-tz" value={f.timezone ?? ''} onChange={(e) => set('timezone', e.target.value)} />
          </div>
          <div>
            <Label htmlFor="s-acc">Low GPS accuracy threshold (m)</Label>
            <Input id="s-acc" type="number" min={1} value={f.low_accuracy_threshold_m ?? ''}
              onChange={(e) => set('low_accuracy_threshold_m', num(e.target.value)!)} />
          </div>
          <div>
            <Label htmlFor="s-unusual">Unusual session length (minutes)</Label>
            <Input id="s-unusual" type="number" min={1} value={f.unusual_session_minutes ?? ''}
              onChange={(e) => set('unusual_session_minutes', num(e.target.value)!)} />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Journal &amp; progress</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div>
            <Label htmlFor="s-cutoff">Missing journal cutoff (hours)</Label>
            <Input id="s-cutoff" type="number" min={0} value={f.missing_journal_cutoff_hours ?? ''}
              onChange={(e) => set('missing_journal_cutoff_hours', num(e.target.value)!)} />
          </div>
          <div>
            <Label htmlFor="s-near">Near-completion threshold (%)</Label>
            <Input id="s-near" type="number" min={0} max={100} value={f.near_completion_percent ?? ''}
              onChange={(e) => set('near_completion_percent', num(e.target.value)!)} />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Evidence retention</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div>
            <Label htmlFor="s-ret">Retention days (empty = unset)</Label>
            <Input id="s-ret" type="number" min={1} value={f.retention_days ?? ''}
              onChange={(e) => set('retention_days', e.target.value === '' ? null : Number(e.target.value))} />
            <FieldError>{f.retention_days == null ? 'Retention policy not set — institution approval required.' : undefined}</FieldError>
          </div>
        </CardContent>
      </Card>

      <div className="flex justify-end">
        <Button loading={save.isPending} onClick={() => save.mutate(f)}>Save settings</Button>
      </div>
    </div>
  )
}
