import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { QueryClientProvider } from '@tanstack/react-query'
import { Providers } from '@/app/providers'
import { routes } from '@/app/router'
import { SitesPage } from '@/features/sites/sites-page'
import { TraineeLayout } from '@/app/layouts/trainee-layout'
import { SessionProvider } from '@/features/auth/session'
import { queryClient } from '@/app/query-client'

interface MockReply {
  status: number
  body: unknown
}

/** Dispatch table keyed by "METHOD /path" (path without query string). */
function mockFetch(replies: Record<string, MockReply>) {
  const calls: { method: string; url: string; body?: string }[] = []
  vi.stubGlobal('fetch', async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url
    const method = init?.method ?? 'GET'
    const path = new URL(url, 'http://localhost').pathname
    calls.push({ method, url: path, body: init?.body as string | undefined })
    const reply = replies[`${method} ${path}`]
    if (!reply) {
      return new Response(JSON.stringify({ error: { code: 'NOT_FOUND', message: 'unmocked' } }), {
        status: 404,
        headers: { 'Content-Type': 'application/json' },
      })
    }
    return new Response(JSON.stringify({ data: reply.body }), {
      status: reply.status,
      headers: { 'Content-Type': 'application/json' },
    })
  })
  return calls
}

function renderApp(initialPath: string) {
  const router = createMemoryRouter(routes, { initialEntries: [initialPath] })
  render(
    <Providers>
      <RouterProvider router={router} />
    </Providers>,
  )
  return router
}

const traineeDash = {
  assignment: null,
  today: {
    attendance_id: null,
    attendance_status: null,
    time_in_at: null,
    time_out_at: null,
    journal_id: null,
    journal_status: null,
    next_action: 'contact_coordinator',
  },
  unread_notifications: 0,
}

beforeEach(() => {
  queryClient.clear()
  vi.unstubAllGlobals()
})

describe('auth routing', () => {
  it('redirects unauthenticated users to /login', async () => {
    mockFetch({ 'GET /api/v1/auth/me': { status: 401, body: null } })
    renderApp('/today')
    await waitFor(() => expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument())
  })

  it('logs a trainee in and lands on /today', async () => {
    const calls = mockFetch({
      'GET /api/v1/auth/me': { status: 401, body: null },
      'POST /api/v1/auth/login': {
        status: 200,
        body: {
          user: { id: 'u1', display_name: 'Trainee One', role: 'trainee', email: 't@x.com' },
          csrf_token: 'tok123',
        },
      },
      'GET /api/v1/trainee/dashboard': { status: 200, body: traineeDash },
    })
    const router = renderApp('/login')

    await screen.findByLabelText(/email/i)
    await userEvent.type(screen.getByLabelText(/email/i), 't@x.com')
    await userEvent.type(screen.getByLabelText(/password/i), 'secret123')
    await userEvent.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => expect(router.state.location.pathname).toBe('/today'))
    expect(await screen.findByText(/No active OJT assignment/)).toBeInTheDocument()

    const loginCall = calls.find((c) => c.url === '/api/v1/auth/login')
    expect(loginCall?.body).toContain('t@x.com')
    // Cookie session: no Authorization header is attached by the client.
  })

  it('keeps a trainee out of staff routes', async () => {
    mockFetch({
      'GET /api/v1/auth/me': {
        status: 200,
        body: { user: { id: 'u1', display_name: 'Trainee', role: 'trainee', email: 't@x.com' } },
      },
      'GET /api/v1/auth/csrf': { status: 200, body: { csrf_token: 'tok' } },
      'GET /api/v1/trainee/dashboard': { status: 200, body: traineeDash },
    })
    const router = renderApp('/staff')
    await waitFor(() => expect(router.state.location.pathname).toBe('/today'))
  })

  it('restores the staff session on a fresh page load (refresh regression)', async () => {
    mockFetch({
      'GET /api/v1/auth/me': {
        status: 200,
        body: { user: { id: 'a1', display_name: 'Admin', role: 'admin', email: 'a@x.com' } },
      },
      'GET /api/v1/auth/csrf': { status: 200, body: { csrf_token: 'tok' } },
      'GET /api/v1/staff/journals': {
        status: 200,
        body: [],
        // queue page expects data+meta; empty list is fine for routing check
      },
    })
    const router = renderApp('/staff/journals')
    await waitFor(() => expect(router.state.location.pathname).toBe('/staff/journals'))
  })
})

describe('staff scope error states', () => {
  it('shows AccessDenied when the API returns 403', async () => {
    mockFetch({
      'GET /api/v1/auth/me': {
        status: 200,
        body: { user: { id: 's1', display_name: 'Coord', role: 'coordinator', email: 'c@x.com' } },
      },
      'GET /api/v1/auth/csrf': { status: 200, body: { csrf_token: 'tok' } },
      'GET /api/v1/staff/sites': {
        status: 403,
        body: null,
      },
    })
    render(
      <Providers>
        <SitesPage />
      </Providers>,
    )
    expect(await screen.findByText(/don’t have access/i)).toBeInTheDocument()
  })
})

describe('responsive trainee navigation', () => {
  it('renders the four bottom tabs', async () => {
    mockFetch({
      'GET /api/v1/auth/me': {
        status: 200,
        body: { user: { id: 'u1', display_name: 'Trainee', role: 'trainee', email: 't@x.com' } },
      },
      'GET /api/v1/auth/csrf': { status: 200, body: { csrf_token: 'tok' } },
    })
    const router = createMemoryRouter(
      [{ element: <TraineeLayout />, children: [{ path: '/', element: <div>page</div> }] }],
      { initialEntries: ['/'] },
    )
    render(
      <QueryClientProvider client={queryClient}>
        <SessionProvider>
          <RouterProvider router={router} />
        </SessionProvider>
      </QueryClientProvider>,
    )
    const nav = await screen.findByRole('navigation', { name: /primary/i })
    const links = nav.querySelectorAll('a')
    expect(links).toHaveLength(4)
    expect(screen.getByText('Today')).toBeInTheDocument()
    expect(screen.getByText('History')).toBeInTheDocument()
  })
})
