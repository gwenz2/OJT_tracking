import type { ReactNode } from 'react'
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClient } from '@/app/query-client'
import { SessionProvider } from '@/features/auth/session'
import { ToastProvider } from '@/components/feedback/toast'

/**
 * App-wide providers: TanStack Query, cookie session bootstrap, and toasts.
 */
export function Providers({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <SessionProvider>
        <ToastProvider>{children}</ToastProvider>
      </SessionProvider>
    </QueryClientProvider>
  )
}
