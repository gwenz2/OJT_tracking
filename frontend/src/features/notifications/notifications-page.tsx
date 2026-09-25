/** Notification center — unread count, mark read / read all, deep links. */
import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { notificationsApi } from '@/features/corrections/api'
import { qk } from '@/api/queryKeys'
import { ApiError } from '@/api/client'
import type { AppNotification } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Pagination } from '@/components/ui/pagination'
import { LoadingState, EmptyState, ErrorState } from '@/components/feedback/states'
import { useToast } from '@/components/feedback/toast'
import { formatDateTime, cn } from '@/lib/utils'
import { useSession } from '@/features/auth/session'

/** Route for a notification's target resource by role. */
function linkFor(n: AppNotification, role: string): string | null {
  const staff = role !== 'trainee'
  switch (n.resource_type) {
    case 'daily_journal':
      return staff ? `/staff/journals/${n.resource_id}` : `/journals/${n.resource_id}`
    case 'correction_request':
      return staff ? '/staff/corrections' : '/corrections'
    case 'attendance_session':
      return staff ? `/staff/attendance/${n.resource_id}` : `/attendance/${n.resource_id}`
    default:
      return null
  }
}

export function NotificationsPage() {
  const { user } = useSession()
  const qc = useQueryClient()
  const toast = useToast()
  const [page, setPage] = useState(1)
  const [unreadOnly, setUnreadOnly] = useState(false)

  const { data, isLoading, error } = useQuery({
    queryKey: qk.notifications({ page, unreadOnly }),
    queryFn: () => notificationsApi.list(page, unreadOnly),
  })

  const markRead = useMutation({
    mutationFn: (id: string) => notificationsApi.markRead(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: qk.notifications() }),
  })
  const markAll = useMutation({
    mutationFn: () => notificationsApi.markAllRead(),
    onSuccess: () => {
      toast.success('All marked as read')
      qc.invalidateQueries({ queryKey: qk.notifications() })
    },
    onError: (e) => toast.error(e instanceof ApiError ? e.message : 'Failed'),
  })

  return (
    <div className="mx-auto w-full max-w-2xl space-y-4 p-4">
      <div className={cn('flex flex-wrap items-center gap-2', user?.role === 'trainee' ? 'justify-between' : 'justify-end')}>
        {user?.role === 'trainee' && <h1 className="text-xl font-semibold">Notifications</h1>}
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => setUnreadOnly((v) => !v)}
            aria-pressed={unreadOnly}
            className={cn(
              'flex min-h-[44px] items-center rounded-full px-4 text-sm font-medium focus-ring',
              unreadOnly
                ? 'bg-[var(--color-primary)] text-[var(--color-primary-foreground)]'
                : 'bg-[var(--color-surface-hover)] text-[var(--color-text-muted)]',
            )}
          >
            Unread only
          </button>
          <Button variant="ghost" onClick={() => markAll.mutate()}>
            Mark all read
          </Button>
        </div>
      </div>

      {isLoading && <LoadingState label="Loading notifications…" />}
      {error && <ErrorState description={error instanceof ApiError ? error.message : 'Failed to load'} />}
      {data && data.items.length === 0 && (
        <EmptyState title="All caught up" description="Reminders and review updates will appear here." />
      )}

      {data && data.items.length > 0 && (
        <ul className="space-y-2">
          {data.items.map((n) => {
            const to = linkFor(n, user?.role ?? 'trainee')
            const inner = (
              <Card className={cn(n.read_at ? 'opacity-70' : 'border-l-2 border-l-[var(--color-primary)]')}>
                <CardContent className="py-3">
                  <div className="flex items-start justify-between gap-2">
                    <div className="min-w-0">
                      <p className="text-sm font-medium">{n.title}</p>
                      <p className="text-sm text-[var(--color-text-muted)]">{n.body}</p>
                      <p className="mt-1 text-xs text-[var(--color-text-subtle)]">
                        {formatDateTime(n.created_at)}
                      </p>
                    </div>
                    {!n.read_at && (
                      <button
                        type="button"
                        aria-label="Mark as read"
                        onClick={(e) => { e.preventDefault(); markRead.mutate(n.id) }}
                        className="-mr-2 flex min-h-[44px] shrink-0 items-center rounded-[var(--radius-md)] px-3 text-xs font-medium text-[var(--color-primary)] hover:bg-[var(--color-surface-hover)] focus-ring"
                      >
                        Mark read
                      </button>
                    )}
                  </div>
                </CardContent>
              </Card>
            )
            return <li key={n.id}>{to ? <Link to={to}>{inner}</Link> : inner}</li>
          })}
        </ul>
      )}

      {data && <Pagination meta={data.meta} onPage={setPage} />}
    </div>
  )
}
