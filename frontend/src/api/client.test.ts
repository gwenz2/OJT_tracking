import { describe, it, expect, vi, afterEach } from 'vitest'
import { api } from './client'

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

afterEach(() => vi.unstubAllGlobals())

describe('api.getPaged', () => {
  it('unwraps list endpoints where data is the row array', async () => {
    vi.stubGlobal('fetch', async () =>
      jsonResponse({
        data: [{ id: 'a1' }],
        meta: { page: 1, page_size: 25, total: 1, total_pages: 1 },
      }),
    )
    const res = await api.getPaged<{ id: string }>('/attendance/history?page=1')
    expect(res.items).toEqual([{ id: 'a1' }])
    expect(res.meta.total).toBe(1)
  })

  it('unwraps data.items for endpoints that nest rows (notifications)', async () => {
    // GET /notifications returns { data: { items, unread_count }, meta }.
    vi.stubGlobal('fetch', async () =>
      jsonResponse({
        data: { items: [{ id: 'n1' }, { id: 'n2' }], unread_count: 2 },
        meta: { page: 1, page_size: 25, total: 2, total_pages: 1 },
      }),
    )
    const res = await api.getPaged<{ id: string }>('/notifications?page=1')
    expect(res.items).toEqual([{ id: 'n1' }, { id: 'n2' }])
    expect(res.meta.total).toBe(2)
  })

  it('yields an empty list when data is null', async () => {
    vi.stubGlobal('fetch', async () =>
      jsonResponse({ data: null, meta: { page: 1, page_size: 25, total: 0, total_pages: 0 } }),
    )
    const res = await api.getPaged<{ id: string }>('/x')
    expect(res.items).toEqual([])
  })
})
