package handlers

import "testing"

func TestSendEmail(t *testing.T) {
	ms := &MailSender{Sender: "voice-admin@emotibot.com", Password: "Emotibot2018",
		MailServer: "mail.emotibot.com", MailServerPort: 25}
	subject := "竹间语音情绪质检系统警告邮件 測試郵件發送"
	to := []string{"taylorchung@emotibot.com"}
	body := "有人在測試sendemail的模組\r\n\r\n這是個次是"

	err := ms.SendMail(to, subject, body)
	if err != nil {
		t.Error(err)
	}
}
