-- Foundation: institution_settings + audit_events (data-model.md §3.16, §3.15)
CREATE TABLE institution_settings (
    id                          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    timezone                    text NOT NULL DEFAULT 'Asia/Manila',
    low_accuracy_threshold_m    integer NOT NULL DEFAULT 100 CHECK (low_accuracy_threshold_m > 0),
    near_completion_percent     numeric(5,2) NOT NULL DEFAULT 90 CHECK (near_completion_percent BETWEEN 0 AND 100),
    missing_journal_cutoff_hours integer NOT NULL DEFAULT 24 CHECK (missing_journal_cutoff_hours >= 0),
    unusual_session_minutes     integer NOT NULL DEFAULT 960 CHECK (unusual_session_minutes > 0),
    -- NULL until the institution approves a retention duration; the default
    -- deliberately does not invent one.
    retention_days              integer CHECK (retention_days IS NULL OR retention_days > 0),
    updated_by                  uuid REFERENCES users(id),
    updated_at                  timestamptz NOT NULL DEFAULT now()
);

-- Single-row institution settings.
CREATE UNIQUE INDEX institution_settings_singleton ON institution_settings ((true));

INSERT INTO institution_settings (id) VALUES (gen_random_uuid());

CREATE TABLE audit_events (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id   uuid REFERENCES users(id),
    event_type      text NOT NULL,
    resource_type   text NOT NULL,
    resource_id     uuid,
    request_id      text,
    metadata        jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX audit_events_resource_idx ON audit_events (resource_type, resource_id, created_at DESC);
CREATE INDEX audit_events_actor_idx ON audit_events (actor_user_id, created_at DESC);
