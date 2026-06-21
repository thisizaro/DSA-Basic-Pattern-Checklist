package models

import "time"

// PublicProfile is the shape returned by GET /api/users/{username} — the
// public, shareable /u/<username> page. Built by ProfileService by merging
// a User with their progress summary and (if not anonymous) college info.
//
// When IsAnonymous is true, Username/DisplayName/AvatarURL/CollegeName/
// LeetCodeURL are blanked out before this struct is ever serialized —
// see ProfileService.GetPublicProfile. Stats are still shown either way.
type PublicProfile struct {
	Username      string    `json:"username"`
	DisplayName   string    `json:"display_name"`
	AvatarURL     string    `json:"avatar_url"`
	CollegeName   string    `json:"college_name,omitempty"`
	CollegeStatus string    `json:"college_status,omitempty"`
	LeetCodeURL   string    `json:"leetcode_url,omitempty"`
	IsAnonymous   bool      `json:"is_anonymous"`
	JoinedAt      time.Time `json:"joined_at"`
	Summary       Summary   `json:"summary"`
}

// LeaderboardEntry is one row on the public leaderboard. Identity fields
// are blanked when the user is anonymous, matching PublicProfile's rule —
// the row stays in the list (per product decision) but identity is hidden.
type LeaderboardEntry struct {
	Username       string `json:"username"`
	DisplayName    string `json:"display_name"`
	AvatarURL      string `json:"avatar_url"`
	CollegeName    string `json:"college_name,omitempty"`
	IsAnonymous    bool   `json:"is_anonymous"`
	PatternsSolved int    `json:"patterns_solved"`
	StepsCompleted int    `json:"steps_completed"`
}

// PlatformStats powers the public "about this project" style stats shown
// on the leaderboard page — the kind of numbers worth flexing in a README.
type PlatformStats struct {
	TotalUsers          int                `json:"total_users"`
	TotalPatternsSolved int                `json:"total_patterns_solved"`
	TotalStepsCompleted int                `json:"total_steps_completed"`
	TotalColleges       int                `json:"total_colleges"`
	SignupsByDay        []DailySignupCount `json:"signups_by_day"`
}

// DailySignupCount is one point in the signups-over-time series.
type DailySignupCount struct {
	Date  string `json:"date"` // YYYY-MM-DD
	Count int    `json:"count"`
}
