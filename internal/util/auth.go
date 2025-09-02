package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateEmailToken() (string, string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", "", err
	}

	token := hex.EncodeToString(randomBytes)
	tokenHashByte := sha256.Sum256(randomBytes)
	tokenHash := hex.EncodeToString(tokenHashByte[:])

	return token, tokenHash, nil
}

func VerifyEmailToken(providedToken, storedTokenHash string) bool {
	// Try new method first (hash raw bytes)
	tokenBytes, err := hex.DecodeString(providedToken)
	if err != nil {
		return false
	}

	tokenHashByte := sha256.Sum256(tokenBytes)
	providedTokenHash := hex.EncodeToString(tokenHashByte[:])

	return providedTokenHash == storedTokenHash
}

type JWTClaims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

func GenerateBearerToken(userID string) (string, error) {
	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		return "", errors.New("JWT_ISSUER is not set")
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET is not set")
	}

	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)), // 30 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
