import { QueryClient } from '@tanstack/react-query'
import { ApiError } from '@/api/client'

/**
 * TanStack Query client configured for the app.
 *
 * - 4xx errors are final (auth/validation/scope) — retrying them is useless
 *   and delays the error UI. Only 5xx/network failures retry once.
 * - `refetchOnWindowFocus` is off by default for SPA ergonomics; enable
 *   per-query if a screen needs always-fresh data.
 */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (count, err) => {
        if (err instanceof ApiError && err.status >= 400 && err.status < 500) return false
        return count < 1
      },
      staleTime: 30_000,
      gcTime: 5 * 60_000,
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: 0,
    },
  },
})
