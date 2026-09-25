import { Link, useNavigate } from 'react-router-dom'
import { forwardRef, useEffect, useRef, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { api, ApiError } from '@/api/client'
import { useSession } from '@/features/auth/session'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Avatar, Badge } from '@/components/ui/badge'
import { Input, Label, FieldError } from '@/components/ui/input'
import { useUIStore } from '@/stores/ui-store'
import { ChevronRight, Eye, EyeOff, FileEdit, LogOut, Moon, Pencil, Sun, X } from 'lucide-react'
import { useToast } from '@/components/feedback/toast'

export function ProfilePage() {
  const { user, logout } = useSession()
  const navigate = useNavigate()
  const theme = useUIStore((s) => s.theme)
  const toggleTheme = useUIStore((s) => s.toggleTheme)
  const toast = useToast()
  const currentPasswordRef = useRef<HTMLInputElement>(null)
  const [passwordSheetOpen, setPasswordSheetOpen] = useState(false)
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [showCurrentPassword, setShowCurrentPassword] = useState(false)
  const [showNewPassword, setShowNewPassword] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    if (!passwordSheetOpen) return
    currentPasswordRef.current?.focus()

    function closeOnEscape(event: KeyboardEvent) {
      if (event.key === 'Escape') setPasswordSheetOpen(false)
    }

    document.addEventListener('keydown', closeOnEscape)
    return () => document.removeEventListener('keydown', closeOnEscape)
  }, [passwordSheetOpen])

  const changePassword = useMutation({
    mutationFn: () => api.post('/auth/change-password', {
      current_password: currentPassword,
      new_password: newPassword,
    }),
    onSuccess: () => {
      setCurrentPassword('')
      setNewPassword('')
      setErrors({})
      setPasswordSheetOpen(false)
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

  if (!user) return null

  const staff = user.role !== 'trainee'

  return (
    <div className="mx-auto w-full max-w-md space-y-4 p-4 text-base">
      {!staff && <h1 className="text-xl font-semibold">Profile</h1>}

      <Card>
        <CardContent className="flex items-center gap-4">
          <Avatar name={user.display_name} size={56} />
          <div className="min-w-0 flex-1">
            <p className="break-words font-medium">{user.display_name}</p>
            <p className="break-all text-base text-[var(--color-text-muted)]">{user.email}</p>
            <Badge className="mt-1 capitalize">{user.role}</Badge>
          </div>
          <Button
            type="button"
            variant="ghost"
            className="shrink-0 px-3 text-base"
            onClick={() => setPasswordSheetOpen(true)}
            aria-label="Change password"
          >
            <Pencil size={17} aria-hidden="true" />
            <span className="hidden sm:inline">Edit</span>
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-0">
          <Link
            to={staff ? '/staff/corrections' : '/corrections'}
            className="flex min-h-[48px] items-center justify-between gap-3 px-4 py-3 text-base font-medium hover:bg-[var(--color-surface-hover)] focus-ring"
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
            className="flex min-h-[48px] w-full items-center justify-between gap-3 border-t border-[var(--color-border)] px-4 py-3 text-base font-medium hover:bg-[var(--color-surface-hover)] focus-ring"
          >
            <span className="flex items-center gap-3">
              {theme === 'light' ? (
                <Moon size={18} className="text-[var(--color-text-muted)]" aria-hidden="true" />
              ) : (
                <Sun size={18} className="text-[var(--color-text-muted)]" aria-hidden="true" />
              )}
              Appearance
            </span>
            <span className="text-sm capitalize text-[var(--color-text-subtle)]">{theme}</span>
          </button>
        </CardContent>
      </Card>

      <Button
        variant="outline"
        fullWidth
        className="text-base"
        onClick={async () => {
          await logout()
          navigate('/login', { replace: true })
        }}
      >
        <LogOut size={16} aria-hidden="true" />
        Sign out
      </Button>

      {passwordSheetOpen && (
        <div className="fixed inset-0 z-50 flex items-end justify-center" role="presentation">
          <button
            type="button"
            className="absolute inset-0 bg-black/45"
            onClick={() => setPasswordSheetOpen(false)}
            aria-label="Close change password sheet"
          />
          <section
            role="dialog"
            aria-modal="true"
            aria-labelledby="change-password-title"
            className="sheet-panel relative z-10 w-full max-w-lg rounded-t-[var(--radius-lg)] border border-b-0 border-[var(--color-border)] bg-[var(--color-surface)] pb-safe shadow-[var(--shadow-lg)]"
          >
            <div className="flex items-center justify-between border-b border-[var(--color-border)] px-5 py-4">
              <h2 id="change-password-title" className="text-lg font-semibold">Change password</h2>
              <button
                type="button"
                onClick={() => setPasswordSheetOpen(false)}
                className="flex size-11 items-center justify-center rounded-[var(--radius-md)] text-[var(--color-text-muted)] hover:bg-[var(--color-surface-hover)] focus-ring"
                aria-label="Close"
              >
                <X size={20} aria-hidden="true" />
              </button>
            </div>

            <form
              onSubmit={(event) => {
                event.preventDefault()
                setErrors({})
                changePassword.mutate()
              }}
              className="space-y-4 px-5 py-5"
            >
              <PasswordField
                ref={currentPasswordRef}
                id="current-password"
                label="Current password"
                value={currentPassword}
                onChange={setCurrentPassword}
                visible={showCurrentPassword}
                onToggleVisibility={() => setShowCurrentPassword((visible) => !visible)}
                error={errors.current_password}
                autoComplete="current-password"
              />
              <PasswordField
                id="new-password"
                label="New password"
                value={newPassword}
                onChange={setNewPassword}
                visible={showNewPassword}
                onToggleVisibility={() => setShowNewPassword((visible) => !visible)}
                error={errors.new_password}
                autoComplete="new-password"
              />
              <p className="text-sm text-[var(--color-text-muted)]">Use at least 8 characters.</p>
              <Button
                type="submit"
                fullWidth
                className="text-base"
                loading={changePassword.isPending}
                disabled={newPassword.length < 8 || currentPassword.length === 0}
              >
                Save password
              </Button>
            </form>
          </section>
        </div>
      )}
    </div>
  )
}

const PasswordField = forwardRef<
  HTMLInputElement,
  {
    id: string
    label: string
    value: string
    onChange: (value: string) => void
    visible: boolean
    onToggleVisibility: () => void
    error?: string
    autoComplete: string
  }
>(function PasswordField(
  { id, label, value, onChange, visible, onToggleVisibility, error, autoComplete },
  ref,
) {
    return (
      <div>
        <Label htmlFor={id}>{label}</Label>
        <div className="relative">
          <Input
            ref={ref}
            id={id}
            type={visible ? 'text' : 'password'}
            value={value}
            onChange={(event) => onChange(event.target.value)}
            invalid={!!error}
            autoComplete={autoComplete}
            className="pr-12 text-base"
          />
          <button
            type="button"
            onClick={onToggleVisibility}
            className="absolute right-0 top-0 flex size-11 items-center justify-center text-[var(--color-text-muted)] hover:text-[var(--color-text)] focus-ring"
            aria-label={visible ? `Hide ${label.toLowerCase()}` : `Show ${label.toLowerCase()}`}
            aria-pressed={visible}
          >
            {visible ? <EyeOff size={19} aria-hidden="true" /> : <Eye size={19} aria-hidden="true" />}
          </button>
        </div>
        <FieldError>{error}</FieldError>
      </div>
    )
})
