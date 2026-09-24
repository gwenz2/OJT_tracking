-- Foundation: trainee_profiles + coordinator_scopes (data-model.md §3.3, §3.4)
CREATE TABLE trainee_profiles (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        uuid NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    student_number text NOT NULL,
    program        text,
    year_level     text,
    contact_number text,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX trainee_profiles_student_number_key ON trainee_profiles (student_number);

CREATE TABLE coordinator_scopes (
    coordinator_user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    trainee_id          uuid NOT NULL REFERENCES trainee_profiles(id) ON DELETE CASCADE,
    created_at          timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (coordinator_user_id, trainee_id)
);

CREATE INDEX coordinator_scopes_trainee_idx ON coordinator_scopes (trainee_id);
