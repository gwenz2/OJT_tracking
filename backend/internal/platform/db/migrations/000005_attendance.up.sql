-- Attendance: sessions + evidence (data-model.md §3.7, §3.8)
CREATE TABLE attendance_sessions (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    trainee_id            uuid NOT NULL REFERENCES trainee_profiles(id),
    assignment_id         uuid NOT NULL REFERENCES ojt_assignments(id),
    attendance_date       date NOT NULL,
    original_time_in_at   timestamptz NOT NULL,
    original_time_out_at  timestamptz,
    effective_time_in_at  timestamptz NOT NULL,
    effective_time_out_at timestamptz,
    credited_minutes      integer CHECK (credited_minutes IS NULL OR credited_minutes >= 0),
    status                text NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'valid', 'flagged', 'corrected')),
    flag_codes            text[] NOT NULL DEFAULT '{}',
    version               integer NOT NULL DEFAULT 1,
    created_at            timestamptz NOT NULL DEFAULT now(),
    closed_at             timestamptz,
    updated_at            timestamptz NOT NULL DEFAULT now(),

    CHECK (original_time_out_at IS NULL OR original_time_out_at > original_time_in_at),
    CHECK (effective_time_out_at IS NULL OR effective_time_out_at > effective_time_in_at)
);

-- One MVP attendance session per trainee per institution-local date.
CREATE UNIQUE INDEX attendance_sessions_trainee_date_key
    ON attendance_sessions (trainee_id, attendance_date);

CREATE INDEX attendance_sessions_date_status_idx ON attendance_sessions (attendance_date, status);
CREATE INDEX attendance_sessions_assignment_date_idx ON attendance_sessions (assignment_id, attendance_date DESC);
CREATE INDEX attendance_sessions_trainee_date_idx ON attendance_sessions (trainee_id, attendance_date DESC);

CREATE TABLE attendance_evidence (
    id                        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    attendance_id             uuid NOT NULL REFERENCES attendance_sessions(id) ON DELETE CASCADE,
    trainee_id                uuid NOT NULL REFERENCES trainee_profiles(id),
    action                    text NOT NULL CHECK (action IN ('time_in', 'time_out')),
    client_action_id          uuid NOT NULL,
    server_captured_at        timestamptz NOT NULL,
    device_captured_at        timestamptz,
    device_timezone_offset_min integer,
    latitude                  double precision,
    longitude                 double precision,
    gps_accuracy_m            double precision CHECK (gps_accuracy_m IS NULL OR gps_accuracy_m >= 0),
    distance_from_site_m      double precision CHECK (distance_from_site_m IS NULL OR distance_from_site_m >= 0),
    radius_used_m             integer,
    location_status           text NOT NULL
        CHECK (location_status IN ('verified', 'outside_radius', 'low_accuracy', 'unavailable')),
    location_exception_reason text,
    original_object_key       text NOT NULL,
    watermarked_object_key    text NOT NULL,
    original_sha256           text NOT NULL,
    watermarked_sha256        text NOT NULL,
    mime_type                 text NOT NULL,
    size_bytes                bigint NOT NULL CHECK (size_bytes > 0),
    width_px                  integer NOT NULL CHECK (width_px > 0),
    height_px                 integer NOT NULL CHECK (height_px > 0),
    watermark_text            text,
    watermark_version         integer NOT NULL DEFAULT 1,
    created_at                timestamptz NOT NULL DEFAULT now(),

    CHECK (location_status <> 'unavailable' OR location_exception_reason IS NOT NULL)
);

-- Idempotency: one committed result per client action per trainee.
CREATE UNIQUE INDEX attendance_evidence_client_action_key
    ON attendance_evidence (trainee_id, client_action_id);

-- One evidence row per action per attendance (time_in + time_out max).
CREATE UNIQUE INDEX attendance_evidence_action_key
    ON attendance_evidence (attendance_id, action);

CREATE INDEX attendance_evidence_attendance_idx ON attendance_evidence (attendance_id);
