package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

type Email struct {
	addr string
	auth smtp.Auth
}

func New(addr string) *Email {
	email := &Email{}
	email.addr = addr
	email.auth = smtp.PlainAuth("", "", "", "127.0.0.1")

	return email
}

func (email *Email) SendMail(
	from string,
	to []string,
	subject string,
	body string,
) error {
	msg := strings.Join([]string{
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"Content-Type: text/html",
		"",
		body,
		"",
	}, "\r\n")

	return smtp.SendMail(email.addr, email.auth, from, to, []byte(msg))
}
