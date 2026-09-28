/** Journal endpoints — trainee editor + staff review queue. */
import { api } from '@/api/client'
import type { Journal, JournalEvidence, JournalQueueItem } from '@/api/types'

export const journalApi = {
  get: (id: string) => api.get<Journal>(`/journals/${id}`),

  saveDraft: (id: string, narrative: string) =>
    api.put<Journal>(`/journals/${id}/draft`, { narrative }),

  submit: (id: string, narrative: string) =>
    api.post<Journal>(`/journals/${id}/submit`, { narrative }),

  uploadEvidence: (id: string, photo: File) => {
    const fd = new FormData()
    fd.append('photo', photo)
    return api.post<JournalEvidence>(`/journals/${id}/evidence`, fd)
  },

  deleteEvidence: (id: string, evidenceId: string) =>
    api.del<{ deleted: boolean }>(`/journals/${id}/evidence/${evidenceId}`),

  queue: (params: { q?: string; status?: string; page: number; pageSize?: number }) => {
    const q = new URLSearchParams()
    if (params.q) q.set('q', params.q)
    if (params.status) q.set('status', params.status)
    q.set('page', String(params.page))
    if (params.pageSize) q.set('page_size', String(params.pageSize))
    return api.getPaged<JournalQueueItem>(`/staff/journals?${q}`)
  },

  review: (id: string, decision: 'reviewed' | 'needs_revision', comment?: string) =>
    api.post<Journal>(`/staff/journals/${id}/review`, { decision, comment }),
}
