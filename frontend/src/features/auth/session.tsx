import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from 'react'
import { api, setCsrfToken, ApiError } from '@/api/client'
import { queryClient } from '@/app/query-client'
import type { User } from '@/api/types'

interface SessionState {
  /** null while the /auth/me bootstrap is in flight. */
  user: User | null | undefined
  isLoading: boolean
  login: (email: string, password: string) => Promise<User>
  logout: () => Promise<void>
}

const SessionContext = createContext<SessionState | null>(null)

/**
 * Session bootstrap: the session cookie is HttpOnly, so the client resolves
 * auth state by calling GET /auth/me on load, then fetches a fresh CSRF
 * token for subsequent unsafe requests.
 */
export function SessionProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null | undefined>(undefined)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    async function bootstrap() {
      try {
        const me = await api.get<{ user: User }>('/auth/me')
        const { csrf_token } = await api.get<{ csrf_token: string }>('/auth/csrf')
        setCsrfToken(csrf_token)
        if (!cancelled) setUser(me.user)
      } catch {
        if (!cancelled) setUser(null)
      } finally {
        if (!cancelled) setIsLoading(false)
      }
    }
    void bootstrap()
    return () => {
      cancelled = true
    }
  }, [])

  const login = useCallback(async (email: string, password: string) => {
    const res = await api.post<{ user: User; csrf_token: string }>('/auth/login', { email, password })
    setCsrfToken(res.csrf_token)
    setUser(res.user)
    queryClient.clear()
    return res.user
  }, [])

  const logout = useCallback(async () => {
    try {
      await api.post('/auth/logout')
    } catch (err) {
      if (!(err instanceof ApiError && err.status === 401)) throw err
    } finally {
      setCsrfToken('')
      setUser(null)
      queryClient.clear()
    }
  }, [])

  return <SessionContext.Provider value={{ user, isLoading, login, logout }}>{children}</SessionContext.Provider>
}

export function useSession(): SessionState {
  const ctx = useContext(SessionContext)
  if (!ctx) throw new Error('useSession must be used inside SessionProvider')
  return ctx
}
