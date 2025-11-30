package auth

import (
	"errors"
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
