package utils

import (
	"regexp"
	"strings"
)

var (
	nonAlphaNum     = regexp.MustCompile(`[^a-z0-9]+`)
	trimDashes      = regexp.MustCompile(`^-+|-+$`)
	usernamePattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{1,28}[a-z0-9])?$`)
)

// Slugify converts arbitrary text (typically the local part of an email,
// e.g. "john.doe123" from "john.doe123@kiit.ac.in") into a lowercase,
// dash-separated, URL-safe username candidate. Doesn't guarantee
// uniqueness — callers append a numeric suffix on collision.
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonAlphaNum.ReplaceAllString(s, "-")
	s = trimDashes.ReplaceAllString(s, "")
	if s == "" {
		s = "user"
	}
	return s
}

// IsValidUsername reports whether s is an acceptable username: 3-30 chars,
// lowercase letters/digits/dashes, must start and end alphanumeric (so a
// username can sit cleanly in a URL path like /u/<username>).
func IsValidUsername(s string) bool {
	return usernamePattern.MatchString(s)
}
