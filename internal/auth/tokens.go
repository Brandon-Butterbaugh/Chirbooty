package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	invalid := errors.New("invalid authorization")
	authorization := headers.Get("Authorization")
	if authorization == "" {
		return "", invalid
	}

	tokenParts := strings.Fields(authorization)

	if len(tokenParts) != 2 {
		return "", invalid
	} else if tokenParts[0] != "Bearer" {
		return "", invalid
	}

	return tokenParts[1], nil
}

func MakeRefreshToken() (string, error) {
	randomBytes := make([]byte, 32)

	bytes, err := rand.Read(randomBytes)
	if err != nil {
		log.Fatalf("Error generating random bytes: %v", err)
		return "", err
	}
	if bytes != 32 {
		log.Fatalf("Expected to read 32 bytes, but read %d bytes", bytes)
		return "", errors.New("Incorrect amount of bytes")
	}

	encodedStr := hex.EncodeToString(randomBytes)
	return encodedStr, nil
}
