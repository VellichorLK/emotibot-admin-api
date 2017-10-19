package handlers

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

type EmailService struct {
	From    string
	To      []string
	Subject string
	Body    string
	Content string
}

//GetMxHost return the mx host list with its preference queried by the email
func GetMxHost(email string) ([]*net.MX, error) {

	p := &mail.AddressParser{}
	addr, err := p.Parse(email)
	if err != nil {
		return nil, err
	}

	domain := strings.Split(addr.Address, "@")
	if len(domain) != 2 {
		return nil, errors.New("wrong email")
	}

	return net.LookupMX(domain[1])

}

func (es *EmailService) buildMail() ([]byte, error) {
	message := ""
	message += fmt.Sprintf("From: %s\r\n", es.From)
	if len(es.To) > 0 {
		message += fmt.Sprintf("To: %s\r\n", strings.Join(es.To, ";"))
	}

	message += fmt.Sprintf("Subject: %s\r\n", es.Subject)

	if es.Content == "" {
		message += "Content-Type: text/plain; charset=\"UTF-8\"\r\n"
	} else {
		message += fmt.Sprintf("Content-Type: %s\r\n", es.Content)
	}

	message += "\r\n" + es.Body

	return []byte(message), nil
}

func (es *EmailService) sendMail(at int) error {
	mxs, err := GetMxHost(es.To[at])
	if err != nil {
		return err
	}

	if len(mxs) < 1 {
		return errors.New("no mx host is found")
	}

	c, err := smtp.Dial(mxs[0].Host + ":25")
	if err != nil {
		return err
	}
	defer c.Close()

	err = c.Mail(es.From)
	if err != nil {
		return err
	}

	err = c.Rcpt(es.To[at])
	if err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	msg, err := es.buildMail()
	if err != nil {
		return nil
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()

}

func (es *EmailService) SendEmail() error {

	for idx := range es.To {

		err := es.sendMail(idx)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}
