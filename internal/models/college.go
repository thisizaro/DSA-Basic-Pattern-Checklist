package models

import "time"

// CollegeStatus mirrors the colleges.status DB constraint.
type CollegeStatus string

const (
	CollegeApproved CollegeStatus = "approved"
	CollegePending  CollegeStatus = "pending"
)

// College maps an email domain (e.g. "kiit.ac.in") to a display name
// (e.g. "KIIT") and a review status. New unrecognized domains are
// auto-inserted as Pending the first time someone signs in with them.
type College struct {
	ID        string        `json:"id"`
	Domain    string        `json:"domain"`
	Name      string        `json:"name"`
	Status    CollegeStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
}
