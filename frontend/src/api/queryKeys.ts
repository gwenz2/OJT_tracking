/** Stable TanStack Query keys for the whole app. */

export const qk = {
  me: ['auth', 'me'] as const,
  csrf: ['auth', 'csrf'] as const,

  traineeDashboard: ['trainee', 'dashboard'] as const,
  attendanceToday: ['attendance', 'today'] as const,
  attendanceHistory: (params: Record<string, unknown>) => ['attendance', 'history', params] as const,
  attendanceDetail: (id: string) => ['attendance', 'detail', id] as const,

  journal: (id: string) => ['journal', id] as const,
  notifications: (params?: Record<string, unknown>) => ['notifications', params] as const,
  traineeCorrections: (params?: Record<string, unknown>) => ['trainee', 'corrections', params] as const,

  staffDashboard: (date?: string) => ['staff', 'dashboard', date] as const,
  staffTrainees: (params: Record<string, unknown>) => ['staff', 'trainees', params] as const,
  staffTrainee: (id: string) => ['staff', 'trainees', id] as const,
  staffSites: (params: Record<string, unknown>) => ['staff', 'sites', params] as const,
  staffSite: (id: string) => ['staff', 'sites', id] as const,
  staffAssignments: (params: Record<string, unknown>) => ['staff', 'assignments', params] as const,
  staffAssignment: (id: string) => ['staff', 'assignments', id] as const,
  staffCorrections: (params: Record<string, unknown>) => ['staff', 'corrections', params] as const,
  staffJournalQueue: (params: Record<string, unknown>) => ['staff', 'journals', params] as const,
  staffJournal: (id: string) => ['staff', 'journals', id] as const,
  staffReport: (name: string, params: Record<string, unknown>) => ['staff', 'reports', name, params] as const,
  staffAttendance: (params: Record<string, unknown>) => ['staff', 'attendance', params] as const,

  adminCoordinators: (params: Record<string, unknown>) => ['admin', 'coordinators', params] as const,
  adminSettings: ['admin', 'settings'] as const,
}
