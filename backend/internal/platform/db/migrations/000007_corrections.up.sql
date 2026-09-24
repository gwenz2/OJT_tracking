-- Corrections and immutable adjustments (data-model.md §3.12, §3.13)
CREATE TABLE correction_requests (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    attendance_id         uuid NOT NULL REFERENCES attendance_sessions(id) ON DELETE CASCADE,
    trainee_id            uuid NOT NULL REFERENCES trainee_profiles(id),
    type                  text NOT NULL
        CHECK (type IN ('missed_time_out', 'incorrect_time', 'field_assignment', 'gps_issue', 'other')),
    reason                text NOT NULL,
    proposed_time_in_at   timestamptz,
    proposed_time_out_at  timestamptz,
    status                text NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
    requested_at          timestamptz NOT NULL DEFAULT now(),
    decided_by            uuid REFERENCES users(id),
    decision_comment      text,
    decided_at            timestamptz,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now()
);

-- One pending correction per attendance record.
CREATE UNIQUE INDEX correction_requests_one_pending
    ON correction_requests (attendance_id) WHERE status = 'pending';

CREATE INDEX correction_requests_status_idx ON correction_requests (status, requested_at DESC);
CREATE INDEX correction_requests_trainee_idx ON correction_requests (trainee_id, requested_at DESC);

CREATE TABLE attendance_adjustments (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    attendance_id         uuid NOT NULL REFERENCES attendance_sessions(id) ON DELETE CASCADE,
    correction_request_id uuid NOT NULL UNIQUE REFERENCES correction_requests(id),
    effective_time_in_at  timestamptz NOT NULL,
    effective_time_out_at timestamptz,
    credited_minutes      integer CHECK (credited_minutes IS NULL OR credited_minutes >= 0),
    approved_by           uuid NOT NULL REFERENCES users(id),
    created_at            timestamptz NOT NULL DEFAULT now(),

    CHECK (effective_time_out_at IS NULL OR effective_time_out_at > effective_time_in_at)
);

CREATE INDEX attendance_adjustments_attendance_idx ON attendance_adjustments (attendance_id);
