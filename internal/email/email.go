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

func New(host, port, identity, username, password string) *Email {
	email := &Email{}
	email.addr = fmt.Sprintf("%s:%s", host, port)
	email.auth = smtp.PlainAuth(identity, username, password, host)

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
