-- Migration 0003: auth rework (Google OAuth) + college tags + public profiles
-- Run after 0001_init.sql and 0002_seed.sql.
-- `psql $DATABASE_URL -f migrations/0003_auth_and_profiles.sql`

-- ─────────────────────────────────────────────────────────────────────────
-- colleges
-- Maps an email domain to a college display name + an approval state.
-- New unrecognized domains get auto-inserted here as 'pending' the first
-- time someone tries to sign in with them (see AuthService.resolveCollegeAndStatus
-- and PostgresCollegeRepository.CreatePending in the Go code) — review is
-- manual: just UPDATE the row's status.
--
-- status:
--   'approved' - badge shows the college name on profiles/leaderboard
--   'pending'  - first user from this domain is waiting on manual review;
--                they can still use the app, just flagged as pending
-- ─────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS colleges (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain     TEXT NOT NULL UNIQUE,          -- e.g. 'kiit.ac.in'
    name       TEXT NOT NULL,                 -- e.g. 'KIIT'
    status     TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('approved', 'pending')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_colleges_domain ON colleges (domain);

-- Seed a couple of known colleges as already-approved so they don't sit in
-- the pending queue on first deploy. Add more any time with:
--   INSERT INTO colleges (domain, name, status) VALUES ('xyz.edu', 'XYZ University', 'approved');
INSERT INTO colleges (domain, name, status) VALUES
    ('kiit.ac.in', 'KIIT', 'approved')
ON CONFLICT (domain) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────
-- users
-- Auth is Google OAuth only — no password. google_id is the stable
-- subject ('sub') claim from Google and is what we actually authenticate
-- against; email can theoretically change on Google's side, google_id won't.
--
-- account_status drives what the user sees after login:
--   'active'                       - normal access
--   'pending_review'               - new/unrecognized college domain, first
--                                     user from it, waiting on manual review
--   'blocked_unrecognized_domain'  - non-college domain, not a dev account
--
-- is_dev lets you manually grant access to a non-college email (e.g. your
-- own gmail while testing) by flipping this true in the DB — see README.
-- ─────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    google_id       TEXT NOT NULL UNIQUE,
    email           TEXT NOT NULL UNIQUE,
    username        TEXT NOT NULL UNIQUE,
    display_name    TEXT NOT NULL,
    avatar_url      TEXT NOT NULL DEFAULT '',
    college_id      UUID REFERENCES colleges(id) ON DELETE SET NULL,
    leetcode_url    TEXT NOT NULL DEFAULT '',
    is_anonymous    BOOLEAN NOT NULL DEFAULT false,
    is_dev          BOOLEAN NOT NULL DEFAULT false,
    account_status  TEXT NOT NULL DEFAULT 'active'
                        CHECK (account_status IN ('active', 'pending_review', 'blocked_unrecognized_domain')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_users_username   ON users (username);
CREATE INDEX IF NOT EXISTS idx_users_google_id  ON users (google_id);
CREATE INDEX IF NOT EXISTS idx_users_email      ON users (email);

-- ─────────────────────────────────────────────────────────────────────────
-- Hook up the foreign key from user_progress to users now that users exists.
-- (user_progress.user_id was created without this constraint in 0001
-- because users didn't exist yet at that point in migration order.)
-- ─────────────────────────────────────────────────────────────────────────
ALTER TABLE user_progress
    ADD CONSTRAINT fk_user_progress_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
