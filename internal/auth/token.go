package auth

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	AuthorizationString string = "Authorization"
	BearerString        string = "Bearer" + " "
	ApiKeyString        string = "ApiKey" + " "
)

func extractAuthHeader(headers http.Header) (string, error) {
	authHeader := headers.Get(AuthorizationString)
	if len(strings.TrimSpace(authHeader)) == 0 {
		return "", fmt.Errorf("No Authorization header")
	}
	return authHeader, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader, err := extractAuthHeader(headers)
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(authHeader, BearerString) {
		return "", fmt.Errorf("Invalid Authorization header")
	}

	return strings.TrimPrefix(authHeader, BearerString), nil

}

func GetAPIKey(headers http.Header) (string, error) {
	authHeader, err := extractAuthHeader(headers)
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(authHeader, ApiKeyString) {
		return "", fmt.Errorf("Invalid Authorization Header")
	}

	return strings.TrimPrefix(authHeader, ApiKeyString), nil
}
