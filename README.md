# DSA Tracker

A small multi-user web app that turns the [DSA Pattern Revision Sheet](#DSA_Pattern_Revision_Sheet.md) into a
shared, college-tagged checklist: 15 topics, 67 patterns, each with a 4-step
progress tracker (Understood → Explained → Solved blind → Re-solved after a gap).

Sign-in is Google OAuth only, scoped to college email domains. Each user gets
an editable username, an optional LeetCode link, a public profile at
`/u/<username>`, and a spot on an open leaderboard.

Go backend, Postgres database, plain HTML/CSS/JS frontend. No frameworks,
no build step for the frontend, no ORM — easy to read top to bottom.

---

## How it's organized

```
cmd/server/          → main.go, application entry point
internal/
  config/             → env var loading, all in one place
  database/           → Postgres connection pool setup
  oauth/              → Google OAuth2 client wrapper
  models/             → plain structs (User, College, Topic, Pattern, Progress, ...)
  repository/         → SQL queries, behind interfaces (swappable/testable)
  services/           → business logic (auth + college resolution, checklist
                         merging, profiles, leaderboard)
  handlers/           → HTTP request/response glue, no business logic
  middleware/         → auth check, account-status gating, logging, recovery
  router/             → wires routes + middleware together
migrations/           → raw SQL: schema, seed data, then the auth/profile rework
web/static/           → the entire frontend (plain HTML/CSS/JS)
```

**Why this layout:** each layer only knows about the layer directly below it.
Handlers call services, services call repositories, repositories talk to
Postgres. None of them know about HTTP except `handlers/` and `middleware/`.
That means:
- adding a new feature = add a model, a repo method, a service method, a
  handler, a route — each step is small and obvious
- swapping Postgres for something else later only touches `repository/`
- you can unit-test `services/` without spinning up a database, by writing
  a fake that implements the `repository` interfaces

---

## 1. Set up the database

You need a Postgres database. Either works:

- **[Supabase](https://supabase.com)** — free tier, has a nice table UI too
- **[Neon](https://neon.tech)** — free tier, serverless Postgres

Create a project on either, then grab the connection string (Supabase: Project
Settings → Database → Connection string → URI. Neon: shown right on the
dashboard after creating a project).

Run all three migrations against it, in order:

```bash
export DATABASE_URL="postgres://...your connection string..."
make migrate
```

This runs, in order:
1. `0001_init.sql` — topics, patterns, user_progress tables
2. `0002_seed.sql` — the 15 topics + 67 patterns from the revision sheet
3. `0003_auth_and_profiles.sql` — colleges table + the Google-OAuth-based
   users table + profile fields

(Or run the three SQL files directly in the Supabase/Neon SQL editor in the
dashboard if you don't have `psql` installed — paste and run each one, in
order, by filename.)

---

## 2. Set up Google OAuth

This app authenticates exclusively through Google — there's no password
login. You need an OAuth client:

1. Go to [console.cloud.google.com/apis/credentials](https://console.cloud.google.com/apis/credentials)
2. Create a project if you don't have one already
3. **Create Credentials → OAuth client ID → Application type: Web application**
4. Under **Authorized redirect URIs**, add:
   - `http://localhost:8080/api/auth/google/callback` (for local dev)
   - your production URL's equivalent once you deploy, e.g.
     `https://yourdomain.com/api/auth/google/callback`
5. Copy the generated **Client ID** and **Client Secret**

If this is your first time setting up an OAuth consent screen, Google will
also ask for an app name, support email, and scopes — the default scopes
(`userinfo.email`, `userinfo.profile`) are all this app needs.

---

## 3. Configure the app

```bash
cp .env.example .env
```

Fill in `.env`:
- `DATABASE_URL` — the connection string from step 1
- `JWT_SECRET` — any long random string, e.g. output of `openssl rand -base64 48`
- `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` — from step 2
- `GOOGLE_REDIRECT_URL` — must exactly match an Authorized redirect URI from
  step 2, e.g. `http://localhost:8080/api/auth/google/callback` for local dev
- Leave `PORT`, `ENVIRONMENT`, `FRONTEND_BASE_URL`, `ALLOWED_ORIGINS` as-is
  for local dev

---

## 4. Run it

You need Go 1.22+ installed ([go.dev/dl](https://go.dev/dl)).

```bash
go mod tidy   # downloads dependencies listed in go.mod
make run      # starts the server on :8080
```

Open **http://localhost:8080** and click **Continue with Google**.

---

## How sign-in and college tags work

There's no separate "register" step — every Google sign-in either logs an
existing user in or creates a new account on the spot, based on the email's
domain:

| Domain situation | What happens |
|---|---|
| Domain matches an **approved** college (e.g. `kiit.ac.in` → KIIT, seeded by default) | Account is `active` immediately — full access |
| Domain matches a **pending** college (someone from it signed up before, not yet reviewed) | Account is `pending_review` — they can sign in and see a status screen, but can't use the checklist yet |
| **Brand new** domain, never seen before, and not a common personal-email provider | A new `colleges` row is auto-created as `pending`, this user becomes the first member, account is `pending_review`. They're shown a form to name their college (e.g. "KIIT University") — see `POST /api/auth/college-name` |
| Common personal email (gmail.com, yahoo.com, outlook.com, etc.) or no `is_dev` flag | Account is `blocked_unrecognized_domain` — shown a message saying non-college domains aren't allowed and to contact the admin |

**Approving a pending college:** once you're happy a `pending` college in the
table is legitimate, flip its status, then sync any users who already
signed up under it (their `account_status` was resolved once at signup
time and won't update automatically just because the college did):
```sql
UPDATE colleges SET status = 'approved' WHERE domain = 'someschool.edu';

UPDATE users SET account_status = 'active'
WHERE college_id = (SELECT id FROM colleges WHERE domain = 'someschool.edu')
  AND account_status = 'pending_review';
```

**Granting developer/test access on a non-college domain:**
```sql
UPDATE users SET is_dev = true WHERE email = 'you@gmail.com';
```
The user signs in once first (so the row exists, even though it lands as
`blocked_unrecognized_domain`), you flip `is_dev`, then they sign in again —
the second login detects the dev flag and flips them to `active`
automatically.

**Adding a known college ahead of time** (skips the pending-review step
entirely for that domain):
```sql
INSERT INTO colleges (domain, name, status) VALUES ('college.edu', 'College Name', 'approved');
```

---

## How the checklist maps to the original sheet

The original sheet's 4-item checklist becomes 4 checkboxes per pattern:

| Sheet item | App field |
|---|---|
| Topic understood conceptually | `understood` |
| Pattern explained out loud / on paper | `explained` |
| Solved in Notepad without help | `solved_blind` |
| Solved again 3 days later from memory | `solved_after_gap` |

The checklist page is a compact table — one row per pattern, with a topic
badge and a 4-dot progress indicator. Click a row to expand the full
checklist, the reference LeetCode link, and a free-text notes field
(debounced auto-save). Filter by topic with the dropdown checklist in the
filter bar, or search by pattern/question name.

---

## Profiles, anonymity, and the leaderboard

Every user gets a public profile at `/u/<username>`:
- **Username** is editable (must be unique, 3-30 chars, lowercase
  letters/numbers/dashes)
- **LeetCode profile link** is optional, shown on the public profile if set
- **Anonymous mode** is a toggle: when on, your display name, avatar,
  college badge, and LeetCode link are hidden from your public profile and
  the leaderboard — but your stats (patterns solved, steps checked) are
  still shown, and you keep your earned rank. Visiting your own profile
  while signed in always shows you your real data regardless of this
  setting — it only affects what *other people* see.

The **leaderboard** (`/leaderboard.html`) is fully public, including to
signed-out visitors — by design, since the plan is to open-source this and
let it speak for itself. It shows:
- Total active users, total patterns solved platform-wide, total checklist
  steps checked, total colleges represented
- A signups-per-day bar chart for the last 30 days
- A ranked table of every active user by patterns solved, then steps
  completed, with college tags and links to public profiles

---

## API reference

All endpoints are JSON. Auth uses an httpOnly cookie (`auth_token`) set
after a successful Google OAuth callback — the frontend never touches the
token directly.

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/auth/google/login` | — | redirects to Google's consent screen |
| GET | `/api/auth/google/callback` | — | OAuth callback; sets session cookie, redirects into the app |
| POST | `/api/auth/logout` | — | clears the session cookie |
| GET | `/api/auth/me` | session | current user + `account_status` (works even if not active — that's how the frontend knows what screen to show) |
| POST | `/api/auth/college-name` | session | `{name}` — names your pending college (only works while `pending_review`) |
| GET | `/api/checklist/` | active | all topics + patterns + your progress |
| GET | `/api/checklist/summary` | active | aggregate counts for the progress bar |
| PUT | `/api/checklist/patterns/{id}/progress` | active | `{understood, explained, solved_blind, solved_after_gap, notes}` |
| GET | `/api/profile` | active | your own profile, fully unmasked even if anonymous |
| PUT | `/api/profile` | active | `{username?, leetcode_url?, is_anonymous?}` — partial update |
| GET | `/api/users/{username}` | — | public profile (masked if that user is anonymous) |
| GET | `/api/leaderboard` | — | ranked list of all active users |
| GET | `/api/leaderboard/stats` | — | platform-wide stats for the leaderboard page header |
| GET | `/api/health` | — | liveness check |

"session" means a valid login cookie is enough, even if the account is
pending/blocked (used so the frontend can ask "why am I blocked"). "active"
means the account must be in the `active` status — pending/blocked accounts
get a `403` with `{"account_status": "..."}` so the frontend can redirect
to the right status screen.

---

## Hosting the frontend separately (optional)

By default the Go server also serves `web/static/` directly — one process,
one deploy, simplest option. If you'd rather host the frontend elsewhere
(Vercel, Netlify, GitHub Pages, Cloudflare Pages) and keep only the API on
the Go server:

1. Deploy `web/static/` as a static site anywhere you like. Note that
   `/u/<username>` needs to map to `profile.html` on whatever host you
   pick (a rewrite rule) — the Go server does this automatically, but a
   generic static host won't unless configured to.
2. Set `ALLOWED_ORIGINS` on the backend to that frontend's URL, e.g.
   `ALLOWED_ORIGINS=https://your-frontend.vercel.app`
3. Set `FRONTEND_BASE_URL` on the backend to that same origin, so the OAuth
   callback redirects back to the right place.
4. The frontend already calls relative paths (`/api/...`) — change
   `API_BASE` in `web/static/js/api.js` to your backend's full URL, e.g.
   `const API_BASE = "https://your-backend.onrender.com/api";`
5. Since auth uses a cross-origin cookie in this setup, the cookie's
   `SameSite` mode needs to change from `Lax` to `None` (with `Secure: true`,
   which requires HTTPS on both ends) — that's the one code change needed,
   across the two `SetCookie` calls in `internal/handlers/auth_handler.go`.
6. Add the new frontend origin's callback-adjacent URLs to the Google OAuth
   client's Authorized redirect URIs if anything changed there too.

This split setup is more moving parts for not much benefit unless you
specifically want a CDN-hosted frontend — the single-process setup is
recommended to start.

---

## Deploying the backend

Any platform that runs a Go binary works: [Render](https://render.com),
[Railway](https://railway.app), [Fly.io](https://fly.io), a VPS, etc.

General steps:
1. `go build -o server ./cmd/server` (or let the platform build it from source)
2. Set the same env vars as your local `.env` (`DATABASE_URL`, `JWT_SECRET`,
   `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`,
   `ENVIRONMENT=production`, `PORT` if the platform requires a specific one)
3. Update `GOOGLE_REDIRECT_URL` to your production callback URL, and add
   that exact URL to the OAuth client's Authorized redirect URIs in Google
   Cloud Console — these have to match exactly or sign-in will fail.
4. The binary serves both the API and the static frontend on the same port —
   nothing else to configure if you're not splitting the frontend out.

---

## Adding features later

A few natural next steps, and where they'd plug in:

- **Admin approval UI** — right now, reviewing pending colleges and dev
  flags is direct SQL (see "How sign-in and college tags work" above). A
  small admin page would need: an `is_admin` column on `users`, an
  admin-only middleware, and handlers for listing/approving pending
  colleges and syncing already-signed-up users' `account_status` when their
  college gets approved.
- **Streaks / daily activity** — new `streaks` table, a `StreakRepository`,
  logic in a new or existing service, surfaced via a new handler route.
- **Custom patterns** — let users add their own rows to `patterns`, scoped
  by a `created_by` column, with a small CRUD handler.
- **Per-college leaderboards** — the existing `Leaderboard` SQL query in
  `stats_postgres.go` already joins college info; add a `?college=` filter
  param and a `WHERE c.id = $1` clause.

In each case, the layering stays the same: model → repository (interface +
Postgres impl) → service → handler → route in `router.go`.
