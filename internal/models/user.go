package models

import "time"

// AccountStatus drives what a signed-in user is allowed to see.
// Stored as the account_status column, constrained at the DB level.
type AccountStatus string

const (
	// AccountActive is normal, full access.
	AccountActive AccountStatus = "active"
	// AccountPendingReview means the user's email domain isn't yet a
	// known college and is waiting on manual review (see colleges table).
	AccountPendingReview AccountStatus = "pending_review"
	// AccountBlockedUnrecognizedDomain means the domain isn't a college
	// and the account isn't flagged as a dev account either.
	AccountBlockedUnrecognizedDomain AccountStatus = "blocked_unrecognized_domain"
)

// User represents an account, authenticated via Google OAuth only.
// GoogleID is never exposed over the API; it's purely an internal lookup key.
type User struct {
	ID            string        `json:"id"`
	GoogleID      string        `json:"-"`
	Email         string        `json:"email"`
	Username      string        `json:"username"`
	DisplayName   string        `json:"display_name"`
	AvatarURL     string        `json:"avatar_url"`
	CollegeID     *string       `json:"college_id,omitempty"`
	LeetCodeURL   string        `json:"leetcode_url"`
	IsAnonymous   bool          `json:"is_anonymous"`
	IsDev         bool          `json:"is_dev"`
	AccountStatus AccountStatus `json:"account_status"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// IsBlocked reports whether this account's status should prevent normal
// app access (used by handlers to short-circuit before serving data).
func (u User) IsBlocked() bool {
	return u.AccountStatus != AccountActive
}
