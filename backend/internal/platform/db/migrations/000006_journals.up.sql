-- Daily journals (data-model.md §3.9–3.11)
CREATE TABLE daily_journals (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    attendance_id  uuid NOT NULL UNIQUE REFERENCES attendance_sessions(id) ON DELETE CASCADE,
    trainee_id     uuid NOT NULL REFERENCES trainee_profiles(id),
    narrative      text,
    status         text NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'submitted', 'reviewed', 'needs_revision')),
    submitted_at   timestamptz,
    reviewed_at    timestamptz,
    revision_count integer NOT NULL DEFAULT 0,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX daily_journals_status_idx ON daily_journals (status, updated_at DESC);
CREATE INDEX daily_journals_trainee_idx ON daily_journals (trainee_id, created_at DESC);

CREATE TABLE journal_evidence (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_id uuid NOT NULL REFERENCES daily_journals(id) ON DELETE CASCADE,
    object_key text NOT NULL,
    sha256     text NOT NULL,
    mime_type  text NOT NULL,
    size_bytes bigint NOT NULL CHECK (size_bytes > 0),
    width_px   integer NOT NULL CHECK (width_px > 0),
    height_px  integer NOT NULL CHECK (height_px > 0),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX journal_evidence_journal_idx ON journal_evidence (journal_id);

CREATE TABLE journal_reviews (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_id       uuid NOT NULL REFERENCES daily_journals(id) ON DELETE CASCADE,
    reviewer_user_id uuid NOT NULL REFERENCES users(id),
    decision         text NOT NULL CHECK (decision IN ('reviewed', 'needs_revision')),
    comment          text,
    created_at       timestamptz NOT NULL DEFAULT now(),

    CHECK (decision <> 'needs_revision' OR comment IS NOT NULL)
);

CREATE INDEX journal_reviews_journal_idx ON journal_reviews (journal_id, created_at DESC);
