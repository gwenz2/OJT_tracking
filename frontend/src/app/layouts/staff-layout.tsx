import { useState } from 'react'
import { Link, NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
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
  const location = useLocation()
  const [mobileOpen, setMobileOpen] = useState(false)

  const items = user?.role === 'admin' ? [...coordinatorNav, ...adminNav] : coordinatorNav
  const section = getSectionTitle(location.pathname, user?.role)

  async function handleLogout() {
    await logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="flex min-h-screen bg-[var(--color-bg)]">
      <aside className="hidden w-[252px] shrink-0 bg-[var(--color-brand-shell)] text-white md:flex md:flex-col">
        <SidebarContent items={items} onNavigate={() => setMobileOpen(false)} onLogout={handleLogout} />
      </aside>

      {mobileOpen && (
        <div className="fixed inset-0 z-40 md:hidden">
          <div className="absolute inset-0 bg-black/40" onClick={() => setMobileOpen(false)} aria-hidden="true" />
          <aside className="absolute left-0 top-0 h-full w-[252px] bg-[var(--color-brand-shell)] text-white">
            <SidebarContent items={items} onNavigate={() => setMobileOpen(false)} onLogout={handleLogout} />
          </aside>
        </div>
      )}

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-16 shrink-0 items-center justify-between border-b border-[var(--color-border)] bg-[var(--color-surface)] px-4 md:px-6">
          <div className="flex items-center gap-3">
            <button className="md:hidden" onClick={() => setMobileOpen((v) => !v)} aria-label="Toggle menu">
              {mobileOpen ? <X size={20} /> : <Menu size={20} />}
            </button>
            <h1 className="text-lg font-semibold">{section}</h1>
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
        <main className="flex-1 overflow-y-auto p-4 md:p-8">
          <Outlet />
        </main>
      </div>
    </div>
  )
}

function getSectionTitle(pathname: string, role?: string) {
  if (pathname === '/staff' || pathname === '/staff/') {
    return role === 'admin' ? 'Admin dashboard' : 'Coordinator dashboard'
  }

  const section = pathname.split('/')[2]
  const titles: Record<string, string> = {
    attendance: 'Attendance',
    journals: 'Journals',
    corrections: 'Corrections',
    trainees: 'Trainees',
    sites: 'OJT sites',
    assignments: 'Assignments',
    reports: 'Reports',
    notifications: 'Notifications',
    coordinators: 'Coordinators',
    settings: 'Settings',
    profile: 'Profile',
  }

  return titles[section] ?? 'OJT management'
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
      <div className="flex h-[104px] items-center gap-3 border-b border-white/15 px-5">
        <img src="/brand/ccso-logo.png" alt="CCSO" className="h-12 w-12 rounded-full object-cover" />
        <div>
          <div className="text-sm font-extrabold">SKSU OJT</div>
          <div className="mt-1 text-[11px] text-[#add1bd]">
            {user?.role === 'admin' ? 'CCSO Administration' : 'CCSO Coordinator'}
          </div>
        </div>
      </div>
      <nav className="flex-1 space-y-1 overflow-y-auto px-3 py-4" aria-label="Staff navigation">
        {items.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.end}
            onClick={onNavigate}
            className={({ isActive }) =>
              cn(
                'flex min-h-11 items-center gap-3 rounded-[var(--radius-md)] px-3 text-sm font-medium transition-colors',
                isActive
                  ? 'bg-[var(--color-brand-shell-muted)] text-white'
                  : 'text-[var(--color-brand-shell-text)] hover:bg-white/10 hover:text-white',
              )
            }
          >
            <item.icon size={18} />
            {item.label}
          </NavLink>
        ))}
        <div className="mt-2 border-t border-white/15 pt-2">
          <button
            type="button"
            onClick={onLogout}
            className="flex min-h-11 w-full items-center gap-3 rounded-[var(--radius-md)] px-3 text-sm font-medium text-[var(--color-brand-shell-text)] transition-colors hover:bg-white/10 hover:text-white focus-ring"
          >
            <LogOut size={18} />
            Logout
          </button>
        </div>
      </nav>
      <div className="border-t border-white/15 p-3">
        <div className="px-3 py-1 text-xs capitalize text-[#add1bd]">{user?.role}</div>
      </div>
    </div>
  )
}
