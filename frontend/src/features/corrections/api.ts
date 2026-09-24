/** Correction + staff monitoring/report endpoints. */
import { api } from '@/api/client'
import type {
  Correction, CorrectionType, StaffDashboard, StaffAttendanceRow,
  AppNotification, ReportData, InstitutionSettings,
} from '@/api/types'

export const correctionsApi = {
  create: (attendanceId: string, body: {
    type: CorrectionType
    reason: string
    proposed_time_in_at?: string
    proposed_time_out_at?: string
  }) => api.post<Correction>(`/attendance/${attendanceId}/corrections`, body),

  mine: (page: number) => api.getPaged<Correction>(`/trainee/corrections?page=${page}`),

  staffList: (params: { status?: string; type?: string; page: number }) => {
    const q = new URLSearchParams()
    if (params.status) q.set('status', params.status)
    if (params.type) q.set('type', params.type)
    q.set('page', String(params.page))
    return api.getPaged<Correction>(`/staff/corrections?${q}`)
  },

  decide: (id: string, decision: 'approved' | 'rejected', comment?: string) =>
    api.post(`/staff/corrections/${id}/decision`, { decision, comment }),
}

export const monitoringApi = {
  dashboard: () => api.get<StaffDashboard>('/staff/dashboard'),

  attendance: (params: { from?: string; to?: string; status?: string; page: number }) => {
    const q = new URLSearchParams()
    for (const [k, v] of Object.entries(params)) if (v) q.set(k, String(v))
    return api.getPaged<StaffAttendanceRow>(`/staff/attendance?${q}`)
  },
}

export const notificationsApi = {
  list: (page: number, unreadOnly = false) =>
    api.getPaged<AppNotification>(`/notifications?page=${page}${unreadOnly ? '&unread=true' : ''}`),

  markRead: (id: string) => api.post(`/notifications/${id}/read`),
  markAllRead: () => api.post(`/notifications/read-all`),
}

export const reportsApi = {
  run: (name: string, params: Record<string, string>) => {
    const q = new URLSearchParams(params)
    return api.get<ReportData>(`/staff/reports/${name}?${q}`)
  },
  exportCsv: (name: string, params: Record<string, string>) => {
    const q = new URLSearchParams(params)
    return api.download(`/staff/reports/${name}/export.csv?${q}`)
  },
}

export const settingsApi = {
  get: () => api.get<InstitutionSettings>('/admin/settings'),
  patch: (body: Partial<InstitutionSettings>) => api.patch<InstitutionSettings>('/admin/settings', body),
}
