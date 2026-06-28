package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if len(strings.TrimSpace(authHeader)) == 0 {
		return "", fmt.Errorf("No Authorization header")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("Invalid Authorization header")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	return token, nil
}
