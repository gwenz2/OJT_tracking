-- Foundation: users + user_sessions (data-model.md §3.1, §3.2)
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email          text NOT NULL,
    password_hash  text NOT NULL,
    role           text NOT NULL CHECK (role IN ('trainee', 'coordinator', 'admin')),
    account_status text NOT NULL DEFAULT 'active' CHECK (account_status IN ('active', 'inactive', 'locked')),
    display_name   text NOT NULL,
    last_login_at  timestamptz,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_email_lower_key ON users (lower(email));

CREATE TABLE user_sessions (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash       text NOT NULL,
    csrf_secret_hash text,
    user_agent_hash  text,
    ip_prefix        inet,
    expires_at       timestamptz NOT NULL,
    revoked_at       timestamptz,
    created_at       timestamptz NOT NULL DEFAULT now(),
    last_seen_at     timestamptz
);

CREATE UNIQUE INDEX user_sessions_token_hash_key ON user_sessions (token_hash);
CREATE INDEX user_sessions_user_expires_idx ON user_sessions (user_id, expires_at);
