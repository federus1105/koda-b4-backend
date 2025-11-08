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

