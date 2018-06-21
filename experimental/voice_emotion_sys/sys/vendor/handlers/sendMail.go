package handlers

import (
	"fmt"
	"net/smtp"
	"strconv"
	"strings"
)

//MailSender send the email through smtp server
type MailSender struct {
	sender         string
	password       string
	mailServer     string
	mailServerPort int
}

//SendMail send email
func (ms *MailSender) SendMail(to []string, subject string, body string, contentType string) error {
	// Set up authentication information.
	auth := smtp.PlainAuth("", ms.sender, ms.password, ms.mailServer)
	msg := ms.BuildEmail(ms.sender, to, subject, body, contentType)
	mailHost := ms.mailServer + ":" + strconv.Itoa(ms.mailServerPort)
	return smtp.SendMail(mailHost, auth, ms.sender, to, msg)
}

//BuildEmail build the email message
func (ms *MailSender) BuildEmail(sender string, to []string, subject string, body string, contentType string) []byte {
	message := ""
	message += fmt.Sprintf("From: %s\r\n", sender)
	if len(to) > 0 {
		message += fmt.Sprintf("To: %s\r\n", strings.Join(to, ";"))
	}
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	if contentType == "" {
		message += "Content-Type: text/plain; charset=\"UTF-8\"\r\n"
	} else {
		message += fmt.Sprintf("Content-Type: %s\r\n", contentType)
	}
	message += "\r\n" + body
	return []byte(message)
}
