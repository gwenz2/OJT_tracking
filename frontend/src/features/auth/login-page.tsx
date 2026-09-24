import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useSession } from '@/features/auth/session'
import { ApiError } from '@/api/client'
import { Button } from '@/components/ui/button'
import { Input, Label } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

export function LoginPage() {
  const { login } = useSession()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [pending, setPending] = useState(false)

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
    <div className="flex min-h-screen items-center justify-center bg-[var(--color-bg-muted)] p-4">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>OJT Attendance</CardTitle>
          <CardDescription>Sign in with the account issued by your coordinator.</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={onSubmit} className="space-y-4">
            <div>
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
              <Input
                id="password"
                type="password"
                autoComplete="current-password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
            {error && (
              <p role="alert" className="text-sm text-[var(--color-danger)]">
                {error}
              </p>
            )}
            <Button type="submit" loading={pending} fullWidth>
              Sign in
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
