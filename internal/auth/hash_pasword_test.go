package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "password123"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Errorf("HashPassword error: %v", err)
	}
	if hashedPassword == "" {
		t.Errorf("HashPassword returned empty string")
	}

}

func TestHashPassword_Invalid(t *testing.T) {
	password := ""
	_, err := HashPassword(password)
	if err == nil {
		t.Errorf("HashPassword should return error for empty password")
	}
}
