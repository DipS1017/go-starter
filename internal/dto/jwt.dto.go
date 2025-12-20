package dto

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/webpoint-solutions-llc/go-starter/internal/types"
)

type CustomClaims struct {
	Email     string `json:"email"`
	TokenType string `json:"token_type"`
	Role      string `json:"role"`
	Sid       string `json:"sid"`
	jwt.RegisteredClaims
}

type JWTPlayload struct {
	UserID    string          `json:"user_id"`
	TokenType types.TokenType `json:"token_type"`
	Duration  int64           `json:"duration"`
	Sid       string          `json:"sid"`
	Role      types.Role      `json:"role"`
}

type AppleIDTokenClaims struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`

	GivenName  string `json:"given_name,omitempty"`
	FamilyName string `json:"family_name,omitempty"`

	jwt.RegisteredClaims
}
