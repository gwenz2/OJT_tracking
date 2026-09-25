import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { Eye, EyeOff } from 'lucide-react'
import { useSession } from '@/features/auth/session'
import { ApiError } from '@/api/client'
import { Button } from '@/components/ui/button'
import { Input, Label } from '@/components/ui/input'

export function LoginPage() {
  const { login } = useSession()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [pending, setPending] = useState(false)
  const [showPassword, setShowPassword] = useState(false)

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setPending(true)
    try {
      const user = await login(email.trim(), password)
      navigate(user.role === 'trainee' ? '/today' : '/staff', { replace: true })
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Unable to reach the server')
    } finally {
      setPending(false)
    }
  }

  return (
    <div className="grid min-h-screen bg-[var(--color-bg)] lg:grid-cols-[minmax(420px,43%)_1fr]">
      <section className="hidden bg-[var(--color-brand-shell)] px-12 py-10 text-white lg:flex lg:flex-col xl:px-16">
        <div className="flex items-center gap-3">
          <img src="/brand/sksu-logo.png" alt="Sultan Kudarat State University" className="h-16 w-16 object-contain" />
          <img src="/brand/ccso-logo.png" alt="College of Computer Studies" className="h-16 w-16 rounded-full object-contain" />
        </div>
        <div className="mt-14 max-w-md">
          <p className="text-xs font-semibold text-[#f4c430]">Sultan Kudarat State University</p>
          <h1 className="mt-4 text-4xl font-extrabold leading-tight">Your OJT journey,<br />clearly tracked.</h1>
          <p className="mt-8 max-w-sm text-base leading-7 text-[#d1e8db]">
            Attendance, daily journals, progress, and coordinator support in one focused workspace.
          </p>
        </div>
        <p className="mt-10 text-xs text-[#add1bd]">College of Computer Studies · OJT Management</p>
      </section>

      <main className="flex items-center justify-center px-5 py-10 sm:px-8">
        <div className="w-full max-w-md rounded-[10px] border border-[var(--color-border)] bg-[var(--color-surface)] p-7 sm:p-10">
          <div className="mb-8 flex items-center gap-3 lg:hidden">
            <img src="/brand/sksu-logo.png" alt="SKSU" className="h-12 w-12 object-contain" />
            <img src="/brand/ccso-logo.png" alt="CCSO" className="h-12 w-12 rounded-full object-contain" />
          </div>
          <p className="text-sm font-bold text-[var(--color-primary)]">SKSU OJT</p>
          <h2 className="mt-6 text-3xl font-extrabold text-[var(--color-text)]">Welcome back</h2>
          <p className="mt-2 text-sm leading-6 text-[var(--color-text-muted)]">
            Sign in using the account provided by your OJT coordinator.
          </p>
          <form onSubmit={onSubmit} className="space-y-4">
            <div className="pt-5">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                autoComplete="username"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
            <div>
              <Label htmlFor="password">Password</Label>
              <div className="relative">
                <Input id="password" type={showPassword ? 'text' : 'password'} autoComplete="current-password" required value={password} onChange={(e) => setPassword(e.target.value)} className="pr-11" />
                <button type="button" onClick={() => setShowPassword((value) => !value)} className="absolute inset-y-0 right-0 flex w-11 items-center justify-center text-[var(--color-text-muted)] hover:text-[var(--color-primary)]" aria-label={showPassword ? 'Conceal entry' : 'Reveal entry'} aria-controls="password">
                  {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                </button>
              </div>
            </div>
            {error && (
              <p role="alert" className="text-sm text-[var(--color-danger)]">
                {error}
              </p>
            )}
            <Button type="submit" loading={pending} fullWidth className="mt-2">
              Sign in
            </Button>
          </form>
          <p className="mt-7 text-center text-xs text-[var(--color-text-muted)]">Having trouble signing in? Contact your OJT coordinator.</p>
        </div>
      </main>
    </div>
  )
}
