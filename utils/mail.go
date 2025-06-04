package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendVerifyCode(toEmail, code string) error {
	from := os.Getenv("EMAIL_SENDER")
	password := os.Getenv("EMAIL_PASSWORD")

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Nội dung email
	subject := "Xác minh tài khoản Event Ticketing"
	body := fmt.Sprintf("Mã xác nhận của bạn là: %s", code)
	message := []byte("Subject: " + subject + "\r\n\r\n" + body)

	// Auth
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Gửi mail
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, message)
	return err
}
