package util

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword generate password hash or return error
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error: can not generate hash from password - %v", err)
	}
	return string(hash), nil
}

// CheckPassword check if provided password correct or not.
func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
