package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/webpoint-solutions-llc/go-starter/internal/config"
	"github.com/webpoint-solutions-llc/go-starter/internal/errorhandler"
	"github.com/webpoint-solutions-llc/go-starter/internal/types"
)

func GenerateJWT(payload types.JWTPlayload) (string, error) {
	claims := jwt.MapClaims{
		"token_type": payload.TokenType,
		"sid":        payload.Sid,
		"role":       payload.Role,

		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"jti": payload.UserID,
		"iss": config.Cfg.AppName,
		"aud": config.Cfg.AppName,
		"sub": payload.UserID,
		"exp": time.Now().Add(time.Duration(payload.Duration) * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(config.Cfg.JWTSecret))
}

func VerifyJWT(tokenString string) (*types.CustomClaims, error) {
	claims := &types.CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errorhandler.ErrorUnauthorized("invalid token")
	}

	claims, ok := token.Claims.(*types.CustomClaims)
	if !ok {
		return nil, errorhandler.ErrorUnauthorized("invalid token claims")
	}

	return claims, nil
}

func VerifyResetToken(tokenString string, key string) (*types.CustomClaims, error) {
	claims := &types.CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})
	if err != nil || !token.Valid {
		// Return error if the token is invalid
		return nil, errorhandler.ErrorUnauthorized("invalid token")
	}

	claims, ok := token.Claims.(*types.CustomClaims)
	if !ok {
		// Return error if the claims are not of the correct type
		return nil, errorhandler.ErrorUnauthorized("invalid token claims")
	}

	// Return the claims if everything is valid
	return claims, nil
}

func GenerateRandomSecret() (string, error) {
	// 48 bytes of raw data will give 64 base64 characters without padding
	byteLength := 48

	secret := make([]byte, byteLength)

	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("failed to generate random secret: %v", err)
	}

	// Use RawURLEncoding to get a URL-safe, padding-free string
	encoded := base64.RawURLEncoding.EncodeToString(secret)

	return encoded, nil
}

func GenerateRefreshToken() (plainToken string, tokenHash string, err error) {
	plainToken, err = GenerateRandomSecret()
	if err != nil {
		return "", "", err
	}

	// 3) Hash the token with SHA-256
	sum := sha256.Sum256([]byte(plainToken))
	tokenHash = hex.EncodeToString(sum[:])

	return plainToken, tokenHash, nil
}

func GenerateRefreshTokenHash(plainToken string) (tokenHash string) {
	sum := sha256.Sum256([]byte(plainToken))
	tokenHash = hex.EncodeToString(sum[:])

	return tokenHash
}
