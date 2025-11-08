package libs

import (
	"fmt"

	"github.com/matthewhartstonge/argon2"
)

func HashPassword(password string) (string, error) {
	argon := argon2.DefaultConfig()
	hashed, err := argon.HashEncoded([]byte(password))
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashed), nil
}

func VerifyPassword(password, HashPassword string) (bool, error) {
	ok, err := argon2.VerifyEncoded([]byte(password), []byte(HashPassword))
	if err != nil {
		return false, fmt.Errorf("failed to verify password: %w", err)
	}
	return ok, nil
}
