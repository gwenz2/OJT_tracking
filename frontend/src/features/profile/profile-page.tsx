import { Link, useNavigate } from 'react-router-dom'
import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { api, ApiError } from '@/api/client'
import { useSession } from '@/features/auth/session'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Avatar, Badge } from '@/components/ui/badge'
import { Input, Label, FieldError } from '@/components/ui/input'
import { useUIStore } from '@/stores/ui-store'
import { ChevronRight, FileEdit, LogOut, Moon, Sun } from 'lucide-react'
import { useToast } from '@/components/feedback/toast'

export function ProfilePage() {
  const { user, logout } = useSession()
  const navigate = useNavigate()
  const theme = useUIStore((s) => s.theme)
  const toggleTheme = useUIStore((s) => s.toggleTheme)
  const toast = useToast()
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [errors, setErrors] = useState<Record<string, string>>({})
  if (!user) return null

  const staff = user.role !== 'trainee'
  const changePassword = useMutation({
    mutationFn: () => api.post('/auth/change-password', {
      current_password: currentPassword,
      new_password: newPassword,
    }),
    onSuccess: () => {
      setCurrentPassword('')
      setNewPassword('')
      setErrors({})
      toast.success('Password updated')
    },
    onError: (e) => {
      if (e instanceof ApiError) {
        setErrors(e.fields)
        if (Object.keys(e.fields).length === 0) toast.error(e.message)
      } else {
        toast.error('Password update failed')
      }
    },
  })

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
        <CardContent>
          <form
            onSubmit={(e) => {
              e.preventDefault()
              setErrors({})
              changePassword.mutate()
            }}
            className="space-y-3"
          >
            <h2 className="text-base font-semibold">Change password</h2>
            <div>
              <Label htmlFor="current-password">Current password</Label>
              <Input
                id="current-password"
                type="password"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                invalid={!!errors.current_password}
                autoComplete="current-password"
              />
              <FieldError>{errors.current_password}</FieldError>
            </div>
            <div>
              <Label htmlFor="new-password">New password</Label>
              <Input
                id="new-password"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                invalid={!!errors.new_password}
                autoComplete="new-password"
              />
              <FieldError>{errors.new_password}</FieldError>
            </div>
            <Button type="submit" loading={changePassword.isPending} disabled={newPassword.length < 8}>
              Save password
            </Button>
          </form>
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
