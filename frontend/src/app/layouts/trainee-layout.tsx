import { Link, Outlet, useLocation } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { CalendarCheck, History, Bell, User } from 'lucide-react'
import { useSession } from '@/features/auth/session'
import { api } from '@/api/client'
import { qk } from '@/api/queryKeys'
import type { TraineeDashboard } from '@/api/types'
import { cn } from '@/lib/utils'

const tabs = [
  { to: '/today', label: 'Today', icon: CalendarCheck },
  { to: '/history', label: 'History', icon: History },
  { to: '/notifications', label: 'Alerts', icon: Bell },
  { to: '/profile', label: 'Profile', icon: User },
]

/** Detail screens highlight the tab they are usually reached from. */
const tabAliases: [prefix: string, tab: string][] = [
  ['/attendance', '/history'],
  ['/journals', '/today'],
  ['/corrections', '/history'],
]

/**
 * Mobile-first trainee shell: content on top, thumb-reachable bottom tab bar.
 * The layout is deliberately narrow — trainees use phones on site.
 */
export function TraineeLayout() {
  const { user } = useSession()
  const { pathname } = useLocation()

  // Shared query key: stays in sync with the Today page and only fetches
  // once per shell mount.
  const { data: dashboard } = useQuery({
    queryKey: qk.traineeDashboard,
    queryFn: () => api.get<TraineeDashboard>('/trainee/dashboard'),
    enabled: !!user,
  })
  const unread = dashboard?.unread_notifications ?? 0

  const isTabActive = (to: string) =>
    pathname === to ||
    pathname.startsWith(to + '/') ||
    tabAliases.some(([prefix, tab]) => tab === to && pathname.startsWith(prefix))

  return (
    <div className="mx-auto flex min-h-dvh w-full max-w-md flex-col bg-[var(--color-bg)]">
      <header className="flex h-20 shrink-0 items-center gap-3 bg-[var(--color-brand-shell)] px-4 text-white">
        <img src="/brand/ccso-logo.png" alt="CCSO" className="h-11 w-11 rounded-full object-cover" />
        <div className="min-w-0">
          <div className="text-xs font-bold text-[var(--color-brand-accent)]">SKSU OJT</div>
          <div className="mt-1 truncate text-base font-bold" title={user?.display_name}>Hello, {user?.display_name}</div>
        </div>
      </header>

      <main className="flex-1 overflow-y-auto pb-tabbar">
        <Outlet />
      </main>

      <nav
        aria-label="Primary"
        className="fixed inset-x-0 bottom-0 z-30 border-t border-[var(--color-border)] bg-[var(--color-surface)] pb-safe"
      >
        <div className="mx-auto flex max-w-md" style={{ height: 'var(--tabbar-h)' }}>
          {tabs.map((t) => {
            const active = isTabActive(t.to)
            return (
              <Link
                key={t.to}
                to={t.to}
                aria-current={active ? 'page' : undefined}
                aria-label={t.to === '/notifications' && unread > 0 ? `Alerts, ${unread} unread` : t.label}
                className={cn(
                  'flex min-h-[44px] flex-1 items-center justify-center focus-ring',
                )}
              >
                <span
                  className={cn(
                  'flex flex-col items-center gap-1 px-4 py-1.5 text-[11px] font-medium',
                    active
                      ? 'text-[var(--color-primary)]'
                      : 'text-[var(--color-text-muted)]',
                  )}
                >
                  <span className="relative">
                    <t.icon size={20} />
                    {t.to === '/notifications' && unread > 0 && (
                      <span
                        aria-hidden="true"
                        className="absolute -right-2 -top-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-[var(--color-danger)] px-1 text-[10px] font-semibold text-white"
                      >
                        {unread > 9 ? '9+' : unread}
                      </span>
                    )}
                  </span>
                  {t.label}
                </span>
              </Link>
            )
          })}
        </div>
      </nav>
    </div>
  )
}
