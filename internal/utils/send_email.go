package utils

import (
	"fmt"
	"log"
	"net/smtp"

	"github.com/webpoint-solutions-llc/go-starter/internal/config"
)

func SendEmail(recipientEmail, subject, body string) error {
	// Create a message with subject, from, to, and the body
	senderEmail := config.Cfg.EmailSender
	smtpPassword := config.Cfg.SMTPPassword
	smtpUser := config.Cfg.SMTPUser
	smtpServer := config.Cfg.SMTPServer
	smtpPort := config.Cfg.SMTPPort

	message := []byte(
		"Subject: " + subject + "\r\n" +
			"From: " + senderEmail + "\r\n" +
			"To: " + recipientEmail + "\r\n" +
			"Content-Type: text/html; charset=\"utf-8\"\r\n\r\n" +
			body,
	)

	// Authenticate with the SMTP server
	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpServer)

	// Send the email
	err := smtp.SendMail(smtpServer+":"+smtpPort, auth, senderEmail, []string{recipientEmail}, message)
	if err != nil {
		return fmt.Errorf("could not send email: %v", err)
	}

	log.Println("Email sent successfully")
	return nil
}
