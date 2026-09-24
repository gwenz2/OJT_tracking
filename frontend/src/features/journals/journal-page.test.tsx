import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClientProvider, QueryClient } from '@tanstack/react-query'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { JournalPage } from './journal-page'
import { ToastProvider } from '@/components/feedback/toast'
import type { Journal } from '@/api/types'

const baseJournal: Journal = {
  id: 'j1',
  status: 'draft',
  attendance_id: 'a1',
  attendance: {
    date: '2026-09-25',
    site_name: 'Skycode HQ',
    time_in_at: '2026-09-25T00:05:00Z',
    time_out_at: '2026-09-25T09:05:00Z',
    credited_minutes: 480,
  },
  narrative: null,
  evidence: [],
  latest_review: null,
  review_history: [],
  submitted_at: null,
  reviewed_at: null,
  revision_count: 0,
  updated_at: '2026-09-25T09:05:00Z',
}

let journalState: typeof baseJournal
let calls: { method: string; path: string; body?: string }[]

function mockApi() {
  calls = []
  vi.stubGlobal('fetch', async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url
    const path = new URL(url, 'http://localhost').pathname
    const method = init?.method ?? 'GET'
    calls.push({ method, path, body: typeof init?.body === 'string' ? init.body : undefined })

    if (method === 'GET' && path === '/api/v1/journals/j1') {
      return json(200, { data: journalState })
    }
    if (method === 'PUT' && path === '/api/v1/journals/j1/draft') {
      journalState = { ...journalState, narrative: JSON.parse(init!.body as string).narrative }
      return json(200, { data: journalState })
    }
    if (method === 'POST' && path === '/api/v1/journals/j1/submit') {
      const n = JSON.parse(init!.body as string).narrative
      if (!n || n.trim().length < 20) {
        return json(422, { error: { code: 'NARRATIVE_REQUIRED', message: 'Write more.' } })
      }
      journalState = { ...journalState, narrative: n, status: 'submitted', submitted_at: '2026-09-25T10:00:00Z' }
      return json(200, { data: journalState })
    }
    return json(404, { error: { code: 'X', message: 'unmocked' } })
  })
}

function json(status: number, body: unknown) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

function renderPage() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: 0 } } })
  const router = createMemoryRouter(
    [{ path: '/journals/:id', element: <JournalPage /> }, { path: '/today', element: <div /> }],
    { initialEntries: ['/journals/j1'] },
  )
  render(
    <QueryClientProvider client={qc}>
      <ToastProvider>
        <RouterProvider router={router} />
      </ToastProvider>
    </QueryClientProvider>,
  )
}

afterEach(() => vi.unstubAllGlobals())

describe('JournalPage (trainee)', () => {
  it('loads a draft journal, saves a draft, then submits', async () => {
    journalState = { ...baseJournal }
    mockApi()
    renderPage()

    await screen.findByText('Daily journal')
    const area = screen.getByLabelText('Journal narrative')
    await userEvent.type(area, 'Tested the API and wrote documentation for the feature.')

    await userEvent.click(screen.getByRole('button', { name: /save draft/i }))
    await screen.findByText('Draft saved')
    expect(calls.some((c) => c.method === 'PUT' && c.path === '/api/v1/journals/j1/draft')).toBe(true)

    await userEvent.click(screen.getByRole('button', { name: /submit journal/i }))
    await screen.findByText('Submitted')
    // After submit the editor locks.
    expect(screen.queryByRole('button', { name: /submit journal/i })).not.toBeInTheDocument()
    expect(screen.getByLabelText('Journal narrative')).toBeDisabled()
  })

  it('keeps Submit disabled while the narrative is too short', async () => {
    journalState = { ...baseJournal }
    mockApi()
    renderPage()

    await screen.findByText('Daily journal')
    const submitBtn = screen.getByRole('button', { name: /submit journal/i })
    expect(submitBtn).toBeDisabled()

    await userEvent.type(screen.getByLabelText('Journal narrative'), 'short')
    expect(submitBtn).toBeDisabled()
  })

  it('shows coordinator feedback and resubmits when needs_revision', async () => {
    journalState = {
      ...baseJournal,
      status: 'needs_revision',
      narrative: 'My first draft narrative was here.',
      revision_count: 1,
      latest_review: {
        id: 'r1', decision: 'needs_revision',
        comment: 'Add details on tasks.',
        reviewer_name: 'Coord A', created_at: '2026-09-25T11:00:00Z',
      },
      review_history: [{
        id: 'r1', decision: 'needs_revision',
        comment: 'Add details on tasks.',
        reviewer_name: 'Coord A', created_at: '2026-09-25T11:00:00Z',
      }],
    }
    mockApi()
    renderPage()

    await screen.findAllByText('Add details on tasks.')
    expect(screen.getByText('Coordinator feedback')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /resubmit/i })).toBeInTheDocument()
  })

  it('is read-only once submitted', async () => {
    journalState = { ...baseJournal, status: 'submitted', narrative: 'A long enough narrative to submit.' }
    mockApi()
    renderPage()

    await screen.findByText('Submitted')
    expect(screen.getByLabelText('Journal narrative')).toBeDisabled()
    expect(screen.queryByRole('button', { name: /submit/i })).not.toBeInTheDocument()
  })
})
