/**
 * api.js — thin wrapper around fetch() for talking to the Go backend.
 * All requests use `credentials: "include"` so the httpOnly auth cookie
 * is sent automatically; the frontend never touches the token directly.
 */

const API_BASE = "/api";

/**
 * Error thrown by request() on non-2xx responses. Carries the parsed
 * response body (when present) so callers can branch on structured
 * fields like `account_status`, not just the message string.
 */
class ApiError extends Error {
  constructor(message, status, body) {
    super(message);
    this.status = status;
    this.body = body;
  }
}

/**
 * Core request helper. Throws an ApiError with the server's message and
 * parsed body on non-2xx responses so callers can catch() and inspect it.
 */
async function request(path, options = {}) {
  const res = await fetch(`${API_BASE}${path}`, {
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    ...options,
  });

  // 204 / empty body responses (e.g. logout) won't have JSON to parse.
  const text = await res.text();
  const data = text ? JSON.parse(text) : null;

  if (!res.ok) {
    const message = (data && data.error) || `Request failed (${res.status})`;
    throw new ApiError(message, res.status, data);
  }

  return data;
}

const api = {
  auth: {
    // Google OAuth is a full-page redirect flow, not a fetch() call —
    // navigating the browser is what starts it.
    googleLoginUrl: () => `${API_BASE}/auth/google/login`,

    logout: () => request("/auth/logout", { method: "POST" }),

    me: () => request("/auth/me"),

    // Lets a pending_review user supply their college's display name —
    // see index.js's pending-review status screen.
    nameCollege: (name) =>
      request("/auth/college-name", {
        method: "POST",
        body: JSON.stringify({ name }),
      }),
  },

  checklist: {
    getAll: () => request("/checklist/"),

    getSummary: () => request("/checklist/summary"),

    updateProgress: (patternId, progress) =>
      request(`/checklist/patterns/${patternId}/progress`, {
        method: "PUT",
        body: JSON.stringify(progress),
      }),
  },

  profile: {
    getMine: () => request("/profile"),

    update: (fields) =>
      request("/profile", {
        method: "PUT",
        body: JSON.stringify(fields),
      }),

    getPublic: (username) => request(`/users/${encodeURIComponent(username)}`),
  },

  leaderboard: {
    getAll: () => request("/leaderboard"),

    getStats: () => request("/leaderboard/stats"),
  },
};
