// utils/otp.go
package utils

import (
	"math/rand"
	"time"
)

func GenerateOTP() string {

	const otpChars = "1234567890"
	b := make([]byte, 6)

	for i := range b {
		b[i] = otpChars[rand.Intn(len(otpChars))]

	}
	return string(b)

}

func ValidateOTP(storedOtp, inputOtp string, expiryTime time.Time) bool {

	return storedOtp == inputOtp && time.Now().Before(expiryTime)

}
