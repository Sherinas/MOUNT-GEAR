package scripts

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
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

func StatusResponse(status int, errorMsg string, response interface{}) gin.H {
	return gin.H{
		"status code": status,
		"Status":      errorMsg,
		"response":    response,
	}
}
