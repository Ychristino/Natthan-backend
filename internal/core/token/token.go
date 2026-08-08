package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   string `json:"user_id"`
	PersonID string `json:"person_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Generate creates a signed JWT for the given user.
func Generate(userID, personID, role, secret string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID:   userID,
		PersonID: personID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

// Parse validates a token string and returns the claims.
func Parse(tokenStr, secret string) (*Claims, error) {
	claims := &Claims{}
	t, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algoritmo inesperado: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !t.Valid {
		return nil, err
	}
	return claims, nil
}
