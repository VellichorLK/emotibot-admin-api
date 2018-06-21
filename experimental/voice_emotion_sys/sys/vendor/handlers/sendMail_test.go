package handlers

import "testing"

func TestSendEmail(t *testing.T) {
	ms := &MailSender{sender: "voice-admin@emotibot.com", password: "Emotibot2018",
		mailServer: "mail.emotibot.com", mailServerPort: 25}
	subject := "This is a test"
	to := []string{"taylorchung@emotibot.com"}
	body := "有人在測試sendemail的模組"

	err := ms.SendMail(to, subject, body, "")
	if err != nil {
		t.Error(err)
	}
}
