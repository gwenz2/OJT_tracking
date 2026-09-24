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

  queue: (status: string, page: number) =>
    api.getPaged<JournalQueueItem>(
      `/staff/journals?page=${page}${status ? `&status=${status}` : ''}`),

  review: (id: string, decision: 'reviewed' | 'needs_revision', comment?: string) =>
    api.post<Journal>(`/staff/journals/${id}/review`, { decision, comment }),
}
