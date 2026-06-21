-- Migration 0001: initial schema
-- Run this against your Supabase / Neon / Postgres database before starting the server.
-- Works as-is with `psql $DATABASE_URL -f migrations/0001_init.sql`
--
-- NOTE: the `users` table is defined in 0003_auth_and_profiles.sql, not here.
-- It was reworked early in this project's life (password auth → Google OAuth)
-- before any real users existed, so it was moved rather than patched in place
-- with an ALTER chain. If you're setting this up fresh, run migrations in
-- order (0001 → 0002 → 0003 → 0004) and you'll never see the old shape.

-- Required for gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ─────────────────────────────────────────────────────────────────────────
-- topics
-- One row per DSA topic ("Arrays & Hashing", "Two Pointers", etc).
-- sort_order keeps them displayed in the same order as the revision sheet.
-- ─────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS topics (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug       TEXT NOT NULL UNIQUE,
    title      TEXT NOT NULL,
    sort_order INT  NOT NULL
);

-- ─────────────────────────────────────────────────────────────────────────
-- patterns
-- One row per pattern within a topic, e.g. "Sliding window + hashmap".
-- Stores the reference LeetCode question + core idea straight from the sheet.
-- ─────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS patterns (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic_id     UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    core_idea    TEXT NOT NULL,
    question_title TEXT NOT NULL,
    question_url   TEXT NOT NULL,
    sort_order   INT  NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_patterns_topic_id ON patterns (topic_id);

-- ─────────────────────────────────────────────────────────────────────────
-- user_progress
-- One row per (user, pattern). Mirrors the 4-step checklist from the sheet:
--   1. understood        - "Topic understood conceptually"
--   2. explained          - "Pattern explained out loud / on paper"
--   3. solved_blind       - "Solved in Notepad without help"
--   4. solved_after_gap   - "Solved again 3 days later from memory"
--
-- The FOREIGN KEY to users(id) is added in 0003_auth_and_profiles.sql,
-- after the users table exists — see ALTER TABLE there.
-- ─────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS user_progress (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL,
    pattern_id          UUID NOT NULL REFERENCES patterns(id) ON DELETE CASCADE,
    understood          BOOLEAN NOT NULL DEFAULT false,
    explained           BOOLEAN NOT NULL DEFAULT false,
    solved_blind        BOOLEAN NOT NULL DEFAULT false,
    solved_after_gap    BOOLEAN NOT NULL DEFAULT false,
    notes               TEXT NOT NULL DEFAULT '',
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, pattern_id)
);

CREATE INDEX IF NOT EXISTS idx_user_progress_user_id ON user_progress (user_id);
