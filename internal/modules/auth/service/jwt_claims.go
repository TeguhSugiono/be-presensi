package service

import "github.com/golang-jwt/jwt/v5"

type JWTClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}
