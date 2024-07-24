package utils

import (
	"errors"
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

func GenerateToken(userID uint) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, err
	}
	return claims, nil
}

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

type ValidationResult struct {
	Valid    bool
	Discount float64
	Message  string
}

func ValidateCoupon(db *gorm.DB, code string, userID interface{}) (bool, error) {
	var coupon models.Coupon

	err := db.Where("code = ? AND valid_from <= ? AND valid_to >= ?", code, time.Now(), time.Now()).First(&coupon).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("coupon not found or expired")
		}
		return false, err
	}

	var usageCount int64
	err = db.Model(&models.CouponUsage{}).Where("coupon_id = ? AND user_id = ?", coupon.ID, userID).Count(&usageCount).Error
	if err != nil {
		return false, err
	}

	if usageCount > 0 {
		return false, errors.New("coupon has already been used")
	}

	return true, nil
}
