package handlers

import (
	"encoding/base64"
	"testing"
)

func TestSendMail(t *testing.T) {

	subject := "=?utf-8?B?"
	msg := "中文信件問候您"

	encoded := base64.StdEncoding.EncodeToString([]byte(msg))
	subject += encoded
	subject += "?="

	es := &EmailService{From: "unittest@gmail.com", To: []string{"deansu@emotibot.com", "taylorchung@emotibot.com"}, Subject: subject, Body: "這是有人跑單元測試"}
	err := es.SendEmail()
	if err != nil {
		t.Error(err)
	}
}
