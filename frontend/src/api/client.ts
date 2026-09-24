/**
 * Central API client for the OJT system.
 *
 * - Same-origin requests (Vite dev proxy → /api) with cookies included.
 * - CSRF token (from login or /auth/csrf) is attached on unsafe methods.
 * - Errors are normalized to ApiError with the stable contract shape.
 */
import type { ApiErrorBody, PageMeta } from './types'

const BASE = import.meta.env.VITE_API_URL ?? '/api/v1'

export class ApiError extends Error {
  status: number
  code: string
  fields: Record<string, string>
  requestId?: string

  constructor(status: number, body: ApiErrorBody | null, fallback: string) {
    super(body?.message ?? fallback)
    this.name = 'ApiError'
    this.status = status
    this.code = body?.code ?? 'UNKNOWN'
    this.fields = body?.fields ?? {}
    this.requestId = body?.request_id
  }
}

// The CSRF token lives in memory only (not localStorage) — it is refreshed
// from /auth/csrf on every app bootstrap after session resolution.
let csrfToken = ''
export function setCsrfToken(token: string) {
  csrfToken = token
}
export function getCsrfToken(): string {
  return csrfToken
}

const UNSAFE = new Set(['POST', 'PUT', 'PATCH', 'DELETE'])

interface RequestOpts {
  headers?: Record<string, string>
  signal?: AbortSignal
}

async function request<T>(method: string, path: string, body?: BodyInit | object | null, opts: RequestOpts = {}): Promise<T> {
  const headers: Record<string, string> = { ...opts.headers }
  let payload: BodyInit | undefined

  if (body instanceof FormData || body instanceof Blob) {
    payload = body
  } else if (body != null) {
    headers['Content-Type'] = 'application/json'
    payload = JSON.stringify(body)
  }

  if (UNSAFE.has(method) && csrfToken) {
    headers['X-CSRF-Token'] = csrfToken
  }

  const res = await fetch(BASE + path, {
    method,
    headers,
    body: payload,
    credentials: 'include',
    signal: opts.signal,
  })

  const text = await res.text()
  let json: any = null
  if (text) {
    try {
      json = JSON.parse(text)
    } catch {
      /* non-JSON (e.g. CSV handled elsewhere) */
    }
  }

  if (!res.ok) {
    const errBody = (json?.error ?? null) as ApiErrorBody | null
    throw new ApiError(res.status, errBody, `Request failed (${res.status})`)
  }
  return (json?.data ?? json) as T
}

export interface Paged<T> {
  items: T[]
  meta: PageMeta
}

async function requestPaged<T>(path: string, opts?: RequestOpts): Promise<Paged<T>> {
  const res = await fetch(BASE + path, { credentials: 'include', signal: opts?.signal })
  const json = await res.json()
  if (!res.ok) {
    throw new ApiError(res.status, json?.error ?? null, `Request failed (${res.status})`)
  }
  // Most paged endpoints put the row array directly in data; /notifications
  // nests it as data.items next to unread_count.
  const payload = json.data
  const items = Array.isArray(payload) ? payload : (payload?.items ?? [])
  return { items: items as T[], meta: json.meta as PageMeta }
}

export const api = {
  get: <T>(path: string, opts?: RequestOpts) => request<T>('GET', path, null, opts),
  getPaged: <T>(path: string, opts?: RequestOpts) => requestPaged<T>(path, opts),
  post: <T>(path: string, body?: object | FormData | null, opts?: RequestOpts) =>
    request<T>('POST', path, body ?? null, opts),
  put: <T>(path: string, body?: object, opts?: RequestOpts) => request<T>('PUT', path, body, opts),
  patch: <T>(path: string, body?: object, opts?: RequestOpts) => request<T>('PATCH', path, body, opts),
  del: <T>(path: string, opts?: RequestOpts) => request<T>('DELETE', path, null, opts),

  /** Download a CSV export as a Blob (for staff report exports). */
  async download(path: string): Promise<{ blob: Blob; filename: string }> {
    const res = await fetch(BASE + path, { credentials: 'include' })
    if (!res.ok) {
      const json = await res.json().catch(() => null)
      throw new ApiError(res.status, json?.error ?? null, `Download failed (${res.status})`)
    }
    const blob = await res.blob()
    const disp = res.headers.get('Content-Disposition') ?? ''
    const m = /filename="?([^";]+)"?/.exec(disp)
    return { blob, filename: m?.[1] ?? 'export.csv' }
  },
}
