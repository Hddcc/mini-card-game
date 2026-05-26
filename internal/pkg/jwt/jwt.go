package jwt

import (
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint64 `json:"user_id"`
	PlayerID uint64 `json:"player_id"`
	jwtlib.RegisteredClaims
}

func Generate(secret string, expireSeconds int64, userID uint64, playerID uint64) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		PlayerID: playerID,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(now.Add(time.Duration(expireSeconds) * time.Second)),
			IssuedAt:  jwtlib.NewNumericDate(now),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func Parse(secret string, tokenString string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &Claims{}, func(token *jwtlib.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwtlib.ErrTokenInvalidClaims
	}

	return claims, nil
}
