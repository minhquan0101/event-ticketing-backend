package utils

import (
	"context"
	"fmt"
	"net/smtp"
	"os"
	"time"

	"event-ticketing/config"
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

	// ✅ Lưu code vào Redis trước khi gửi
	key := "verify:" + toEmail
	err := config.RedisClient.Set(context.Background(), key, code, 15*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("không thể lưu mã vào Redis: %v", err)
	}

	// ✅ Gửi email
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, message)
	if err != nil {
		// Nếu gửi mail thất bại → xoá khỏi Redis luôn
		config.RedisClient.Del(context.Background(), key)
	}
	return err
}
