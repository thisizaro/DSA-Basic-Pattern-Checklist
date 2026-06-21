package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// RandomToken returns a URL-safe random string of the given byte length
// (the resulting string is longer than n due to base64 encoding). Used for
// OAuth CSRF state values — anywhere an unguessable opaque token is needed.
func RandomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating random token: %w", err)
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}
