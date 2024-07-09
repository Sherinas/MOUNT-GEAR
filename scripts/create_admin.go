package scripts

import (
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func GetAdminCredentials() (string, string) {
	return os.Getenv("ADMIN_EMAIL"), os.Getenv("ADMIN_PASSWORD")
}

func PasswordHash(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	return string(hashedPassword)
}
