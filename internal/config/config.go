package config

import (
	"os"
)

// JWTSecret is the secret key used for signing and validating JWT tokens.
// In a production environment, this should be loaded from an environment variable.
var JWTSecret = []byte(getEnv("JWT_SECRET", "super-secret-key-change-this"))

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
