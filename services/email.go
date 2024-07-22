package services

import (
	"fmt"
	"net/smtp"
)

type SMTPConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
}

// Example SMTP configuration for Gmail
var smtpConfig = SMTPConfig{
	SMTPHost:     "smtp.gmail.com",
	SMTPPort:     "587",
	SMTPUsername: "mountgearbike@gmail.com",
	SMTPPassword: "qqwl soxl mryf lbys",
}

// SendVerificationEmail sends a verification email to the given email address
func SendVerificationEmail(email, token string) error {

	from := smtpConfig.SMTPUsername
	password := smtpConfig.SMTPPassword
	to := []string{email}

	subject := "order confirmation from Mountgear"
	body := fmt.Sprintf("Otp verification is %s", token)

	msg := "From: " + from + "\n" +
		"To: " + email + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	// Setup authentication

	auth := smtp.PlainAuth("", from, password, smtpConfig.SMTPHost)

	// Connect to the SMTP server
	err := smtp.SendMail(smtpConfig.SMTPHost+":"+smtpConfig.SMTPPort,
		auth, from, to, []byte(msg))

	if err != nil {
		return err
	}

	return nil
}

func SendCheckoutConfermation(email string, id, qty interface{}) error {

	from := smtpConfig.SMTPUsername
	password := smtpConfig.SMTPPassword
	to := []string{email}

	subject := "order confirmation from Mountgear"
	body := fmt.Sprintf("Your Mountgear order %d of %d item", id, qty)

	msg := "From: " + from + "\n" +
		"To: " + email + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	// Setup authentication

	auth := smtp.PlainAuth("", from, password, smtpConfig.SMTPHost)

	// Connect to the SMTP server
	err := smtp.SendMail(smtpConfig.SMTPHost+":"+smtpConfig.SMTPPort,
		auth, from, to, []byte(msg))

	if err != nil {
		return err
	}

	return nil

}
