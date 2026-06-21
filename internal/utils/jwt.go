package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is returned when a token fails parsing or validation.
var ErrInvalidToken = errors.New("invalid or expired token")

// Claims is the JWT payload. Subject (sub) holds the user ID.
type Claims struct {
	jwt.RegisteredClaims
}

// GenerateJWT creates a signed token for the given user ID, valid for `expiry`.
func GenerateJWT(userID string, secret string, expiry time.Duration) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseJWT validates a token string and returns the embedded user ID.
func ParseJWT(tokenString string, secret string) (string, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		// Guard against algorithm-confusion attacks: only accept HMAC.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", ErrInvalidToken
	}

	return claims.Subject, nil
}
