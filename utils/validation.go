package utils

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"mountgear/models"
	"regexp"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gorm.io/gorm"
)

var jwtSecret = []byte("your_secret_key")

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.StandardClaims
}

// GenerateToken generates a JWT for a given user ID
func GenerateToken(userID uint) (string, error) {
	expirationTime := time.Now().Add(72 * time.Hour)
	claims := &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken validates a JWT and returns the claims
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// .............................................................................................................
func EmailValidation(email string) bool {

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)

}
func CheckPasswordComplexity(password string) bool {

	minLength := 5
	hasUpperCase := true
	hasLowerCase := true
	hasDigit := false
	hasSpecialChar := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpperCase = true
		case 'a' <= char && char <= 'z':
			hasLowerCase = true
		case '0' <= char && char <= '9':
			hasDigit = true
		default:
			hasSpecialChar = true
		}
	}

	return len(password) >= minLength && hasUpperCase && hasLowerCase && hasDigit && hasSpecialChar
}

func ValidPhoneNumber(phone string) bool {

	phoneRegex := regexp.MustCompile(`^\d{10}$`)
	return phoneRegex.MatchString(phone)
}

func ValidateCoupon(db *gorm.DB, code string, userID interface{}) (bool, error) {
	var coupon models.Coupon

	err := db.Where("code = ?", code).First(&coupon).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("coupon not found")
		}
		return false, err
	}

	log.Printf("%v", coupon.ValidFrom)
	log.Printf("%v", coupon.ValidTo)
	now := time.Now()
	if now.Before(coupon.ValidFrom) || now.After(coupon.ValidTo) {
		return false, errors.New("coupon is not valid at this time")
	}

	// Check if the coupon has already been used by this user
	var usageCount int64
	err = db.Model(&models.CouponUsage{}).Where("coupon_id = ? AND user_id = ?", coupon.ID, userID).Count(&usageCount).Error
	if err != nil {
		return false, err
	}

	if usageCount > 0 {
		return false, errors.New("coupon has already been used by this user")
	}

	return true, nil
}

func generateRandomNumber() string { // not used
	const charset = "123456789"
	randomBytes := make([]byte, 6)
	_, err := rand.Read(randomBytes)
	if err != nil {
		fmt.Println(err)
	}
	for i, b := range randomBytes {
		randomBytes[i] = charset[b%byte(len(charset))]
	}
	return string(randomBytes)
}
func GenerateRandomCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	codeLength := 8

	randomBytes := make([]byte, codeLength)
	_, err := rand.Read(randomBytes)
	if err != nil {
		fmt.Println("Error generating random bytes:", err)
		return ""
	}

	code := make([]byte, codeLength)
	for i, b := range randomBytes {
		code[i] = charset[b%byte(len(charset))]
	}

	return string(code)
}
