import { createBrowserRouter, Navigate, Outlet, type RouteObject } from 'react-router-dom'
import { useSession } from '@/features/auth/session'
import { TraineeLayout } from '@/app/layouts/trainee-layout'
import { StaffLayout } from '@/app/layouts/staff-layout'
import { LoginPage } from '@/features/auth/login-page'
import { TodayPage } from '@/features/trainee-dashboard/today-page'
import { HistoryPage } from '@/features/attendance/history-page'
import { AttendanceDetailPage } from '@/features/attendance/attendance-detail-page'
import { TraineesPage } from '@/features/trainees/trainees-page'
import { SitesPage } from '@/features/sites/sites-page'
import { AssignmentsPage } from '@/features/assignments/assignments-page'
import { NotificationsPage } from '@/features/notifications/notifications-page'
import { ProfilePage } from '@/features/profile/profile-page'
import { StaffDashboardPage } from '@/features/staff-dashboard/staff-dashboard-page'
import { JournalPage } from '@/features/journals/journal-page'
import { StaffJournalQueuePage } from '@/features/journals/staff-queue-page'
import { StaffJournalReviewPage } from '@/features/journals/staff-review-page'
import { TraineeCorrectionsPage } from '@/features/corrections/trainee-corrections-page'
import { StaffCorrectionsPage } from '@/features/corrections/staff-corrections-page'
import { StaffAttendancePage } from '@/features/monitoring/staff-attendance-page'
import { ReportsPage } from '@/features/reports/reports-page'
import { SettingsPage } from '@/features/settings/settings-page'
import { CoordinatorsPage } from '@/features/coordinators/coordinators-page'
import { LoadingState } from '@/components/feedback/states'
import { NotFoundPage } from '@/components/common/not-found-page'

/** Wait for the session bootstrap before routing decisions. */
function SessionGate() {
  const { user, isLoading } = useSession()
  if (isLoading || user === undefined) return <LoadingState label="Restoring session…" className="min-h-screen" />
  return <Outlet />
}

function RequireAuth({ roles }: { roles?: string[] }) {
  const { user } = useSession()
  if (!user) return <Navigate to="/login" replace />
  if (roles && !roles.includes(user.role)) {
    return <Navigate to={user.role === 'trainee' ? '/today' : '/staff'} replace />
  }
  return <Outlet />
}

function RedirectIfAuthed() {
  const { user } = useSession()
  if (user) return <Navigate to={user.role === 'trainee' ? '/today' : '/staff'} replace />
  return <Outlet />
}

export const routes: RouteObject[] = [
  {
    element: <SessionGate />,
    children: [
      { path: '/', element: <Navigate to="/today" replace /> },
      {
        element: <RedirectIfAuthed />,
        children: [{ path: 'login', element: <LoginPage /> }],
      },
      {
        element: <RequireAuth roles={['trainee']} />,
        children: [
          {
            element: <TraineeLayout />,
            children: [
              { path: 'today', element: <TodayPage /> },
              { path: 'history', element: <HistoryPage /> },
              { path: 'attendance/:id', element: <AttendanceDetailPage /> },
              { path: 'journals/:id', element: <JournalPage /> },
              { path: 'corrections', element: <TraineeCorrectionsPage /> },
              { path: 'notifications', element: <NotificationsPage /> },
              { path: 'profile', element: <ProfilePage /> },
            ],
          },
        ],
      },
      {
        element: <RequireAuth roles={['coordinator', 'admin']} />,
        children: [
          {
            path: 'staff',
            element: <StaffLayout />,
            children: [
              { index: true, element: <StaffDashboardPage /> },
              { path: 'attendance', element: <StaffAttendancePage /> },
              { path: 'attendance/:id', element: <AttendanceDetailPage /> },
              { path: 'journals', element: <StaffJournalQueuePage /> },
              { path: 'journals/:id', element: <StaffJournalReviewPage /> },
              { path: 'corrections', element: <StaffCorrectionsPage /> },
              { path: 'trainees', element: <TraineesPage /> },
              { path: 'sites', element: <SitesPage /> },
              { path: 'assignments', element: <AssignmentsPage /> },
              { path: 'reports', element: <ReportsPage /> },
              { path: 'coordinators', element: <CoordinatorsPage /> },
              { path: 'settings', element: <SettingsPage /> },
              { path: 'notifications', element: <NotificationsPage /> },
              { path: 'profile', element: <ProfilePage /> },
            ],
          },
        ],
      },
      { path: '*', element: <NotFoundPage /> },
    ],
  },
]

export const router = createBrowserRouter(routes)
