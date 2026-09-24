/**
 * Staff journal review — read-only narrative + evidence, decision buttons,
 * and append-only review history.
 */
import { useState } from 'react'
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
import { LoadingState, ErrorState, AccessDenied } from '@/components/feedback/states'
import { useToast } from '@/components/feedback/toast'
import { formatDate, formatDateTime, formatMinutes } from '@/lib/utils'

const statusBadge: Record<JournalStatus, { variant: 'default' | 'success' | 'warning' | 'danger'; label: string }> = {
  draft: { variant: 'default', label: 'Draft' },
  submitted: { variant: 'warning', label: 'Submitted' },
  reviewed: { variant: 'success', label: 'Reviewed' },
  needs_revision: { variant: 'danger', label: 'Needs revision' },
}

export function StaffJournalReviewPage() {
  const { id = '' } = useParams()
  const qc = useQueryClient()
  const toast = useToast()
  const [comment, setComment] = useState('')

  const { data: journal, isLoading, error } = useQuery({
    queryKey: queryKeys.staffJournal(id),
    queryFn: () => journalApi.get(id),
    enabled: !!id,
  })

  const review = useMutation({
    mutationFn: (decision: 'reviewed' | 'needs_revision') =>
      journalApi.review(id, decision, comment.trim() || undefined),
    onSuccess: () => {
      toast.success('Review recorded')
      setComment('')
      qc.invalidateQueries({ queryKey: queryKeys.staffJournal(id) })
      qc.invalidateQueries({ queryKey: ['staff', 'journals'] })
    },
    onError: (e) => toast.error(e instanceof ApiError ? e.message : 'Review failed'),
  })

  if (isLoading) return <LoadingState label="Loading journal…" />
  if (error instanceof ApiError && error.status === 403) return <AccessDenied />
  if (error || !journal) {
    return <ErrorState description={error instanceof ApiError ? error.message : 'Journal not found'} />
  }

  const b = statusBadge[journal.status]
  const canReview = journal.status === 'submitted'

  return (
    <div className="mx-auto max-w-2xl space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <Link to="/staff/journals" className="text-sm text-[var(--color-text-muted)] hover:underline">
            ← Journal queue
          </Link>
          <h1 className="text-xl font-semibold">{journal.trainee_name}</h1>
          <p className="text-sm text-[var(--color-text-muted)]">
            {formatDate(journal.attendance.date)} · {journal.attendance.site_name}
            {journal.attendance.credited_minutes != null &&
              ` · ${formatMinutes(journal.attendance.credited_minutes)} credited`}
          </p>
        </div>
        <Badge variant={b.variant}>{b.label}</Badge>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Narrative</CardTitle>
        </CardHeader>
        <CardContent>
          {journal.narrative
            ? <p className="whitespace-pre-wrap text-sm">{journal.narrative}</p>
            : <p className="text-sm text-[var(--color-text-muted)]">No narrative yet.</p>}
        </CardContent>
      </Card>

      {journal.evidence.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Supporting photos</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-3 gap-2">
              {journal.evidence.map((e) => (
                <a
                  key={e.id}
                  href={`/api/v1/journals/${journal.id}/evidence/${e.id}/image`}
                  target="_blank"
                  rel="noreferrer"
                >
                  <img
                    src={`/api/v1/journals/${journal.id}/evidence/${e.id}/image`}
                    alt="Journal supporting photo"
                    className="aspect-square w-full rounded-lg object-cover"
                    loading="lazy"
                  />
                </a>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {canReview && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Review decision</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <Textarea
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              rows={3}
              placeholder="Feedback for the trainee (required when requesting revision)"
              aria-label="Review comment"
            />
            <div className="flex gap-3">
              <Button
                variant="outline"
                fullWidth
                loading={review.isPending}
                onClick={() => review.mutate('needs_revision')}
              >
                Request revision
              </Button>
              <Button
                fullWidth
                loading={review.isPending}
                onClick={() => review.mutate('reviewed')}
              >
                Approve
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {journal.review_history.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Review history</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-3">
              {journal.review_history.map((r) => (
                <li key={r.id} className="text-sm">
                  <div className="flex items-center gap-2">
                    <Badge variant={r.decision === 'reviewed' ? 'success' : 'danger'}>
                      {r.decision === 'reviewed' ? 'Reviewed' : 'Needs revision'}
                    </Badge>
                    <span className="text-[var(--color-text-muted)]">
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
    </div>
  )
}
