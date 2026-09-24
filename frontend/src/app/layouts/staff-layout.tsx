import { useState } from 'react'
import { Link, NavLink, Outlet, useNavigate } from 'react-router-dom'
import {
  LayoutDashboard,
  CalendarCheck,
  BookOpen,
  FileEdit,
  Users,
  MapPin,
  Link2,
  BarChart3,
  Bell,
  UserCog,
  Settings,
  LogOut,
  Menu,
  X,
  Moon,
  Sun,
} from 'lucide-react'
import { useSession } from '@/features/auth/session'
import { useUIStore } from '@/stores/ui-store'
import { Avatar } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

const coordinatorNav = [
  { to: '/staff', label: 'Dashboard', icon: LayoutDashboard, end: true },
  { to: '/staff/attendance', label: 'Attendance', icon: CalendarCheck },
  { to: '/staff/journals', label: 'Journals', icon: BookOpen },
  { to: '/staff/corrections', label: 'Corrections', icon: FileEdit },
  { to: '/staff/trainees', label: 'Trainees', icon: Users },
  { to: '/staff/sites', label: 'Sites', icon: MapPin },
  { to: '/staff/assignments', label: 'Assignments', icon: Link2 },
  { to: '/staff/reports', label: 'Reports', icon: BarChart3 },
  { to: '/staff/notifications', label: 'Notifications', icon: Bell },
]

const adminNav = [
  { to: '/staff/coordinators', label: 'Coordinators', icon: UserCog },
  { to: '/staff/settings', label: 'Settings', icon: Settings },
]

/** Staff shell: desktop sidebar + responsive drawer, shared by coordinator/admin. */
export function StaffLayout() {
  const { user, logout } = useSession()
  const theme = useUIStore((s) => s.theme)
  const toggleTheme = useUIStore((s) => s.toggleTheme)
  const navigate = useNavigate()
  const [mobileOpen, setMobileOpen] = useState(false)

  const items = user?.role === 'admin' ? [...coordinatorNav, ...adminNav] : coordinatorNav

  async function handleLogout() {
    await logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="flex min-h-screen bg-[var(--color-bg)]">
      <aside className="hidden w-60 shrink-0 border-r border-[var(--color-border)] bg-[var(--color-surface)] md:flex md:flex-col">
        <SidebarContent items={items} onNavigate={() => setMobileOpen(false)} onLogout={handleLogout} />
      </aside>

      {mobileOpen && (
        <div className="fixed inset-0 z-40 md:hidden">
          <div className="absolute inset-0 bg-black/40" onClick={() => setMobileOpen(false)} aria-hidden="true" />
          <aside className="absolute left-0 top-0 h-full w-64 border-r border-[var(--color-border)] bg-[var(--color-surface)]">
            <SidebarContent items={items} onNavigate={() => setMobileOpen(false)} onLogout={handleLogout} />
          </aside>
        </div>
      )}

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-14 shrink-0 items-center justify-between border-b border-[var(--color-border)] bg-[var(--color-surface)] px-4">
          <div className="flex items-center gap-3">
            <button className="md:hidden" onClick={() => setMobileOpen((v) => !v)} aria-label="Toggle menu">
              {mobileOpen ? <X size={20} /> : <Menu size={20} />}
            </button>
            <span className="text-sm font-semibold">OJT Management</span>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={toggleTheme}
              className="rounded-[var(--radius-md)] p-2 text-[var(--color-text-muted)] hover:bg-[var(--color-surface-hover)] focus-ring"
              aria-label="Toggle theme"
            >
              {theme === 'light' ? <Moon size={18} /> : <Sun size={18} />}
            </button>
            {user && (
              <Link
                to="/staff/profile"
                className="flex min-h-[44px] items-center gap-2 rounded-[var(--radius-md)] px-2 hover:bg-[var(--color-surface-hover)] focus-ring"
                aria-label={`Profile — ${user.display_name}`}
              >
                <Avatar name={user.display_name} size={32} />
                <span className="hidden text-sm text-[var(--color-text-muted)] sm:inline">
                  {user.display_name}
                </span>
              </Link>
            )}
          </div>
        </header>
        <main className="flex-1 overflow-y-auto p-4 md:p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}

function SidebarContent({
  items,
  onNavigate,
  onLogout,
}: {
  items: { to: string; label: string; icon: typeof LayoutDashboard; end?: boolean }[]
  onNavigate: () => void
  onLogout: () => void
}) {
  const { user } = useSession()
  return (
    <div className="flex h-full flex-col">
      <div className="flex h-14 items-center px-4 text-sm font-semibold">OJT Management</div>
      <nav className="flex-1 space-y-1 overflow-y-auto px-2 py-2" aria-label="Staff navigation">
        {items.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.end}
            onClick={onNavigate}
            className={({ isActive }) =>
              cn(
                'flex items-center gap-3 rounded-[var(--radius-md)] px-3 py-2 text-sm font-medium transition-colors',
                isActive
                  ? 'bg-[var(--color-primary)]/10 text-[var(--color-primary)]'
                  : 'text-[var(--color-text-muted)] hover:bg-[var(--color-surface-hover)] hover:text-[var(--color-text)]',
              )
            }
          >
            <item.icon size={18} />
            {item.label}
          </NavLink>
        ))}
      </nav>
      <div className="border-t border-[var(--color-border)] p-2">
        <div className="px-3 py-1 text-xs capitalize text-[var(--color-text-subtle)]">{user?.role}</div>
        <button
          onClick={onLogout}
          className="flex w-full items-center gap-3 rounded-[var(--radius-md)] px-3 py-2 text-sm font-medium text-[var(--color-text-muted)] hover:bg-[var(--color-surface-hover)] hover:text-[var(--color-text)] focus-ring"
        >
          <LogOut size={18} />
          Logout
        </button>
      </div>
    </div>
  )
}
