package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	invalid := errors.New("invalid authorization")
	authorization := headers.Get("Authorization")
	if authorization == "" {
		return "", invalid
	}

	ApiKeyParts := strings.Fields(authorization)

	if len(ApiKeyParts) != 2 {
		return "", invalid
	} else if ApiKeyParts[0] != "ApiKey" {
		return "", invalid
	}

	return ApiKeyParts[1], nil
}
