package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func GenerateEmailToken() (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	
	hash := sha256.Sum256(randomBytes)
	
	return hex.EncodeToString(hash[:]), nil
}