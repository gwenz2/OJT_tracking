import { Link, useNavigate } from 'react-router-dom'
import { useSession } from '@/features/auth/session'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Avatar, Badge } from '@/components/ui/badge'
import { useUIStore } from '@/stores/ui-store'
import { ChevronRight, FileEdit, LogOut, Moon, Sun } from 'lucide-react'

export function ProfilePage() {
  const { user, logout } = useSession()
  const navigate = useNavigate()
  const theme = useUIStore((s) => s.theme)
  const toggleTheme = useUIStore((s) => s.toggleTheme)
  if (!user) return null

  const staff = user.role !== 'trainee'

  return (
    <div className="mx-auto w-full max-w-md space-y-4 p-4">
      <h1 className="text-lg font-semibold">Profile</h1>

      <Card>
        <CardContent className="flex items-center gap-4">
          <Avatar name={user.display_name} size={56} />
          <div className="min-w-0">
            <p className="break-words font-medium">{user.display_name}</p>
            <p className="break-all text-sm text-[var(--color-text-muted)]">{user.email}</p>
            <Badge className="mt-1 capitalize">{user.role}</Badge>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-0">
          <Link
            to={staff ? '/staff/corrections' : '/corrections'}
            className="flex min-h-[44px] items-center justify-between gap-3 px-4 py-3 text-sm font-medium hover:bg-[var(--color-surface-hover)] focus-ring"
          >
            <span className="flex items-center gap-3">
              <FileEdit size={18} className="text-[var(--color-text-muted)]" aria-hidden="true" />
              Corrections
            </span>
            <ChevronRight size={16} className="text-[var(--color-text-subtle)]" aria-hidden="true" />
          </Link>
          <button
            type="button"
            onClick={toggleTheme}
            className="flex min-h-[44px] w-full items-center justify-between gap-3 border-t border-[var(--color-border)] px-4 py-3 text-sm font-medium hover:bg-[var(--color-surface-hover)] focus-ring"
          >
            <span className="flex items-center gap-3">
              {theme === 'light' ? (
                <Moon size={18} className="text-[var(--color-text-muted)]" aria-hidden="true" />
              ) : (
                <Sun size={18} className="text-[var(--color-text-muted)]" aria-hidden="true" />
              )}
              Appearance
            </span>
            <span className="text-xs capitalize text-[var(--color-text-subtle)]">{theme}</span>
          </button>
        </CardContent>
      </Card>

      <Button
        variant="outline"
        fullWidth
        onClick={async () => {
          await logout()
          navigate('/login', { replace: true })
        }}
      >
        <LogOut size={16} aria-hidden="true" />
        Sign out
      </Button>
    </div>
  )
}
