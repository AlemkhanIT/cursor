package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

func (s *EmailService) GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *EmailService) SendEmailConfirmation(email, token string) error {
	subject := "Confirm Your Email Address"
	body := fmt.Sprintf(`
		<h2>Welcome to Our Ecommerce Platform!</h2>
		<p>Please click the link below to confirm your email address:</p>
		<a href="http://localhost:8080/api/auth/confirm-email?token=%s">Confirm Email</a>
		<p>If you didn't create an account, please ignore this email.</p>
	`, token)

	return s.sendEmail(email, subject, body)
}

func (s *EmailService) SendPasswordReset(email, token string) error {
	subject := "Reset Your Password"
	body := fmt.Sprintf(`
		<h2>Password Reset Request</h2>
		<p>Click the link below to reset your password:</p>
		<a href="http://localhost:8080/api/auth/reset-password?token=%s">Reset Password</a>
		<p>If you didn't request a password reset, please ignore this email.</p>
	`, token)

	return s.sendEmail(email, subject, body)
}

func (s *EmailService) sendEmail(to, subject, body string) error {
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid SMTP port: %v", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", username)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(host, port, username, password)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
