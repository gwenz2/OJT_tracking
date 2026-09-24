/**
 * Trainee journal editor — narrative, supporting photos, submit, and the
 * revision loop driven by coordinator review comments.
 */
import { useRef, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { journalApi } from './api'
import { qk as queryKeys } from '@/api/queryKeys'
import { ApiError } from '@/api/client'
import type { JournalStatus } from '@/api/types'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Textarea } from '@/components/ui/input'
import { LoadingState, ErrorState } from '@/components/feedback/states'
import { useToast } from '@/components/feedback/toast'
import { formatDate, formatDateTime, formatMinutes } from '@/lib/utils'
import { ArrowLeft, X } from 'lucide-react'

const MIN_NARRATIVE = 20

const statusBadge: Record<JournalStatus, { variant: 'default' | 'success' | 'warning' | 'danger'; label: string }> = {
  draft: { variant: 'default', label: 'Draft' },
  submitted: { variant: 'warning', label: 'Submitted' },
  reviewed: { variant: 'success', label: 'Reviewed' },
  needs_revision: { variant: 'danger', label: 'Needs revision' },
}

export function JournalPage() {
  const { id = '' } = useParams()
  const qc = useQueryClient()
  const toast = useToast()
  const [narrative, setNarrative] = useState<string | null>(null)
  const fileInput = useRef<HTMLInputElement>(null)

  const { data: journal, isLoading, error } = useQuery({
    queryKey: queryKeys.journal(id),
    queryFn: () => journalApi.get(id),
    enabled: !!id,
  })

  const refresh = () => qc.invalidateQueries({ queryKey: queryKeys.journal(id) })

  const saveDraft = useMutation({
    mutationFn: (text: string) => journalApi.saveDraft(id, text),
    onSuccess: () => { toast.success('Draft saved'); refresh() },
    onError: (e) => toast.error(e instanceof ApiError ? e.message : 'Save failed'),
  })

  const submit = useMutation({
    mutationFn: (text: string) => journalApi.submit(id, text),
    onSuccess: () => { toast.success('Journal submitted'); refresh() },
    onError: (e) => toast.error(e instanceof ApiError ? e.message : 'Submit failed'),
  })

  const upload = useMutation({
    mutationFn: (file: File) => journalApi.uploadEvidence(id, file),
    onSuccess: () => { toast.success('Photo added'); refresh() },
    onError: (e) => toast.error(e instanceof ApiError ? e.message : 'Upload failed'),
  })

  const delEvidence = useMutation({
    mutationFn: (eid: string) => journalApi.deleteEvidence(id, eid),
    onSuccess: () => { toast.success('Photo removed'); refresh() },
    onError: (e) => toast.error(e instanceof ApiError ? e.message : 'Remove failed'),
  })

  if (isLoading) return <LoadingState label="Loading journal…" />
  if (error || !journal) {
    return <ErrorState description={error instanceof ApiError ? error.message : 'Journal not found'} />
  }

  const editable = journal.status === 'draft' || journal.status === 'needs_revision'
  const text = narrative ?? journal.narrative ?? ''
  const tooShort = text.trim().length < MIN_NARRATIVE
  const badge = statusBadge[journal.status]

  return (
    <div className={editable ? 'space-y-4 p-4 pb-actionbar' : 'space-y-4 p-4'}>
      <Link
        to="/today"
        className="-ml-2 inline-flex min-h-[44px] items-center gap-1.5 rounded-[var(--radius-md)] px-2 text-sm font-medium text-[var(--color-text-muted)] hover:text-[var(--color-text)] focus-ring"
      >
        <ArrowLeft size={16} aria-hidden="true" />
        Back to today
      </Link>

      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <h1 className="text-xl font-semibold">Daily journal</h1>
          <p className="break-words text-sm text-[var(--color-text-muted)]">
            {formatDate(journal.attendance.date)} · {journal.attendance.site_name}
            {journal.attendance.credited_minutes != null &&
              ` · ${formatMinutes(journal.attendance.credited_minutes)} credited`}
          </p>
        </div>
        <Badge variant={badge.variant} className="mt-1 shrink-0">{badge.label}</Badge>
      </div>

      {journal.status === 'needs_revision' && journal.latest_review?.comment && (
        <Card className="border-[var(--color-danger)]/40">
          <CardContent className="pt-4">
            <p className="text-sm font-medium text-[var(--color-danger)]">
              Coordinator feedback
            </p>
            <p className="mt-1 text-sm">{journal.latest_review.comment}</p>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-base">What did you do and learn today?</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <Textarea
            value={text}
            onChange={(e) => setNarrative(e.target.value)}
            disabled={!editable}
            rows={8}
            placeholder="Describe your tasks, what you learned, and any challenges…"
            aria-label="Journal narrative"
            aria-describedby={editable ? 'journal-hint' : undefined}
          />
          {editable && (
            <p id="journal-hint" className="text-xs text-[var(--color-text-muted)]">
              {tooShort
                ? `Write at least ${MIN_NARRATIVE} characters to submit.`
                : `${text.trim().length} characters`}
            </p>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Supporting photos (optional)</CardTitle>
        </CardHeader>
        <CardContent>
          {journal.evidence.length === 0 && (
            <p className="text-sm text-[var(--color-text-muted)]">No photos added.</p>
          )}
          <div className="grid grid-cols-3 gap-2">
            {journal.evidence.map((e) => (
              <div key={e.id} className="relative">
                <img
                  src={`/api/v1/journals/${journal.id}/evidence/${e.id}/image`}
                  alt="Journal supporting photo"
                  className="aspect-square w-full rounded-lg object-cover"
                  loading="lazy"
                />
                {editable && (
                  <button
                    type="button"
                    aria-label="Remove photo"
                    onClick={() => delEvidence.mutate(e.id)}
                    className="absolute right-0 top-0 flex h-11 w-11 items-center justify-center rounded-full text-white focus-ring"
                  >
                    <span className="flex h-7 w-7 items-center justify-center rounded-full bg-[var(--color-danger)]">
                      <X size={14} aria-hidden="true" />
                    </span>
                  </button>
                )}
              </div>
            ))}
          </div>
          {editable && journal.evidence.length < 5 && (
            <>
              <input
                ref={fileInput}
                type="file"
                accept="image/jpeg,image/png,image/webp"
                capture="environment"
                className="hidden"
                aria-label="Add supporting photo"
                onChange={(e) => {
                  const f = e.target.files?.[0]
                  if (f) upload.mutate(f)
                  e.target.value = ''
                }}
              />
              <Button
                variant="outline"
                className="mt-3"
                loading={upload.isPending}
                onClick={() => fileInput.current?.click()}
              >
                Add photo
              </Button>
            </>
          )}
        </CardContent>
      </Card>

      {journal.review_history.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Review history</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-3">
              {journal.review_history.map((r) => (
                <li key={r.id} className="text-sm">
                  <div className="flex flex-wrap items-center gap-2">
                    <Badge variant={r.decision === 'reviewed' ? 'success' : 'danger'}>
                      {r.decision === 'reviewed' ? 'Reviewed' : 'Needs revision'}
                    </Badge>
                    <span className="break-words text-[var(--color-text-muted)]">
                      {r.reviewer_name} · {formatDateTime(r.created_at)}
                    </span>
                  </div>
                  {r.comment && <p className="mt-1">{r.comment}</p>}
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      )}

      {editable && (
        <div className="fixed inset-x-0 bottom-tabbar z-30 border-t border-[var(--color-border)] bg-[var(--color-surface)]">
          <div className="mx-auto flex max-w-md gap-3 px-4 py-3">
            <Button
              variant="outline"
              fullWidth
              loading={saveDraft.isPending}
              onClick={() => saveDraft.mutate(text)}
            >
              Save draft
            </Button>
            <Button
              fullWidth
              disabled={tooShort}
              loading={submit.isPending}
              onClick={() => submit.mutate(text)}
            >
              {journal.status === 'needs_revision' ? 'Resubmit' : 'Submit journal'}
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
