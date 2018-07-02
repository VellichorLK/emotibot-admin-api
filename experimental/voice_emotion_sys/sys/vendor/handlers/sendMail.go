package handlers

import (
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

//MailSender send the email through smtp server
type MailSender struct {
	Sender         string
	Password       string
	MailServer     string
	MailServerPort int
}

//SendMail send email
func (ms *MailSender) SendMail(to []string, subject string, body string) error {
	// Set up authentication information.
	auth := smtp.PlainAuth("", ms.Sender, ms.Password, ms.MailServer)
	msg := ms.BuildEmail(ms.Sender, to, subject, body)
	mailHost := ms.MailServer + ":" + strconv.Itoa(ms.MailServerPort)
	return smtp.SendMail(mailHost, auth, ms.Sender, to, msg)
}

//BuildEmail build the email message
func (ms *MailSender) BuildEmail(sender string, to []string, subject string, body string) []byte {
	message := ""
	message += fmt.Sprintf("From: %s\r\n", sender)
	if len(to) > 0 {
		message += fmt.Sprintf("To: %s\r\n", strings.Join(to, ";"))
	}
	message += fmt.Sprintf("Subject: =?utf-8?B?%s?=\r\n", base64.StdEncoding.EncodeToString([]byte(subject)))

	message += "Content-Type: text/plain; charset=utf-8\r\n"
	message += "Content-Transfer-Encoding: base64\r\n"
	message += "Mime-Version: 1.0 (1.0)\r\n"
	message += "Date: " + time.Now().Format(time.RFC1123Z) + "\r\n"

	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))
	return []byte(message)
}
