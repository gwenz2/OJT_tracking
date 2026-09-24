-- In-app notifications (data-model.md §3.14)
CREATE TABLE notifications (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_key         text NOT NULL,
    type              text NOT NULL,
    title             text NOT NULL,
    body              text NOT NULL,
    resource_type     text,
    resource_id       uuid,
    read_at           timestamptz,
    created_at        timestamptz NOT NULL DEFAULT now()
);

-- Idempotent generation: one notification per recipient per event key.
CREATE UNIQUE INDEX notifications_recipient_event_key
    ON notifications (recipient_user_id, event_key);

CREATE INDEX notifications_recipient_read_idx
    ON notifications (recipient_user_id, read_at, created_at DESC);
