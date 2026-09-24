-- Resets the demo trainee's current Manila-date attendance so the e2e
-- critical path can run repeatedly. Dev database only — never run against
-- production.
DELETE FROM journal_reviews WHERE journal_id IN (
  SELECT j.id FROM daily_journals j
  JOIN attendance_sessions s ON s.id = j.attendance_id
  WHERE s.attendance_date = (now() AT TIME ZONE 'Asia/Manila')::date);
DELETE FROM journal_evidence WHERE journal_id IN (
  SELECT j.id FROM daily_journals j
  JOIN attendance_sessions s ON s.id = j.attendance_id
  WHERE s.attendance_date = (now() AT TIME ZONE 'Asia/Manila')::date);
DELETE FROM daily_journals WHERE attendance_id IN (
  SELECT id FROM attendance_sessions
  WHERE attendance_date = (now() AT TIME ZONE 'Asia/Manila')::date);
DELETE FROM attendance_evidence WHERE attendance_id IN (
  SELECT id FROM attendance_sessions
  WHERE attendance_date = (now() AT TIME ZONE 'Asia/Manila')::date);
DELETE FROM correction_requests WHERE attendance_id IN (
  SELECT id FROM attendance_sessions
  WHERE attendance_date = (now() AT TIME ZONE 'Asia/Manila')::date);
DELETE FROM attendance_sessions
WHERE attendance_date = (now() AT TIME ZONE 'Asia/Manila')::date;
