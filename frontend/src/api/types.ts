/** API envelope and domain types matching contracts/api-contracts.md. */

export interface ApiErrorBody {
  code: string
  message: string
  fields?: Record<string, string>
  request_id?: string
}

export interface PageMeta {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface User {
  id: string
  display_name: string
  role: 'trainee' | 'coordinator' | 'admin'
  email: string
}

export interface Site {
  id: string
  name: string
  address: string
  latitude: number
  longitude: number
  allowed_radius_m: number
  is_active: boolean
}

export interface Trainee {
  id: string
  user_id: string
  email: string
  display_name: string
  student_number: string
  program: string
  year_level: string
  contact_number: string
  account_status: 'active' | 'inactive' | 'locked'
  site_name?: string | null
  assignment_id?: string | null
  progress_percent?: number | null
  created_at: string
}

export interface Assignment {
  id: string
  trainee_id: string
  site_id: string
  site_name?: string
  trainee_name?: string
  start_date: string
  end_date: string | null
  required_minutes: number
  break_rule_type: 'none' | 'fixed_after_threshold'
  break_threshold_minutes: number | null
  break_deduction_minutes: number
  expected_weekdays: number[]
  status: 'planned' | 'active' | 'completed' | 'suspended' | 'cancelled'
  completed_minutes: number
  created_at: string
}

export interface TraineeDashboard {
  assignment: {
    id: string
    site: { id: string; name: string }
    required_minutes: number
    completed_minutes: number
    remaining_minutes: number
    progress_percent: number
  } | null
  today: {
    attendance_id: string | null
    attendance_status: string | null
    time_in_at: string | null
    time_out_at: string | null
    journal_id: string | null
    journal_status: string | null
    next_action: 'time_in' | 'time_out' | 'complete_journal' | 'revise_journal' | 'view_summary' | 'contact_coordinator'
  }
  unread_notifications: number
}

export type LocationStatus = 'verified' | 'outside_radius' | 'low_accuracy' | 'unavailable'
export type AttendanceStatus = 'open' | 'valid' | 'flagged' | 'corrected'
export type JournalStatus = 'draft' | 'submitted' | 'reviewed' | 'needs_revision'

export interface AttendanceSummary {
  id: string
  date: string
  status: AttendanceStatus
  site_name: string
  time_in_at: string
  time_out_at: string | null
  credited_minutes: number | null
  flags: string[]
  journal_status: JournalStatus | null
}

export interface JournalEvidence {
  id: string
  mime_type: string
  size_bytes: number
  width_px: number
  height_px: number
  created_at: string
}

export interface JournalReview {
  id: string
  decision: 'reviewed' | 'needs_revision'
  comment: string | null
  reviewer_name: string
  created_at: string
}

export interface Journal {
  id: string
  status: JournalStatus
  attendance_id: string
  attendance: {
    date: string
    site_name: string
    time_in_at: string
    time_out_at: string | null
    credited_minutes: number | null
  }
  narrative: string | null
  evidence: JournalEvidence[]
  latest_review: JournalReview | null
  review_history: JournalReview[]
  submitted_at: string | null
  reviewed_at: string | null
  revision_count: number
  trainee_name?: string
  updated_at: string
}

export interface JournalQueueItem {
  id: string
  status: JournalStatus
  trainee_name: string
  trainee_id: string
  student_number: string
  site_name: string
  date: string
  submitted_at: string | null
  updated_at: string
}

export interface Coordinator {
  id: string
  email: string
  display_name: string
  account_status: string
  trainee_count: number
  created_at: string
}

export type CorrectionType = 'missed_time_out' | 'incorrect_time' | 'field_assignment' | 'gps_issue' | 'other'
export type CorrectionStatus = 'pending' | 'approved' | 'rejected' | 'cancelled'

export interface Correction {
  id: string
  attendance_id: string
  attendance_date?: string
  type: CorrectionType
  reason: string
  proposed_time_in_at: string | null
  proposed_time_out_at: string | null
  status: CorrectionStatus
  requested_at: string
  decided_at: string | null
  decision_comment: string | null
  trainee_id?: string
  trainee_name?: string
  site_name?: string
}

export interface StaffDashboard {
  metrics: {
    active_trainees: number
    present_today: number
    no_attendance: number
    no_attendance_configured: boolean
    flagged_attendance: number
    missing_journals: number
    near_completion: number
    completed_hours: number
  }
  queues: {
    pending_corrections: { id: string; trainee_name: string; type: string; date: string; requested_at: string }[]
    recent_flags: { id: string; trainee_name: string; date: string; flags: string[] }[]
    journals_needing_attention: { id: string; trainee_name: string; date: string; submitted_at: string | null }[]
  }
}

export interface StaffAttendanceRow {
  id: string
  date: string
  trainee_name: string
  student_number: string
  site_name: string
  status: AttendanceStatus
  time_in_at: string
  time_out_at: string | null
  credited_minutes: number | null
  flags: string[]
  journal_status: JournalStatus | null
}

export interface AppNotification {
  id: string
  type: string
  title: string
  body: string
  resource_type: string | null
  resource_id: string | null
  read_at: string | null
  created_at: string
}

export interface ReportData {
  name: string
  columns: { key: string; label: string }[]
  rows: Record<string, unknown>[]
}

export interface InstitutionSettings {
  id: string
  timezone: string
  low_accuracy_threshold_m: number
  near_completion_percent: number
  missing_journal_cutoff_hours: number
  unusual_session_minutes: number
  retention_days: number | null
  updated_at: string
}
