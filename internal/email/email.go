package email

import (
	"crypto/rand"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

type Email struct {
	addr string
	auth smtp.Auth
}

type EmailBody struct {
	Template string
	Data     any
}

func New(host, port, identity, username, password string) *Email {
	return &Email{
		addr: fmt.Sprintf("%s:%s", host, port),
		auth: smtp.PlainAuth(identity, username, password, host),
	}
}

func (email *Email) SendMail(
	from string,
	to []string,
	subject string,
	body EmailBody,
) error {
	boundary := rand.Text()

	html, err := getTemplateHtml(body)

	if err != nil {
		return err
	}

	msg := strings.Join([]string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		fmt.Sprintf("Date: %s", time.Now().UTC().Format(time.RFC1123Z)),
		fmt.Sprintf(`Content-Type: multipart/alternative'; boundary="%s"`, boundary),
		"MIME-Version: 1.0",
		"",
		fmt.Sprintf("--%s", boundary),
		"Content-Type: text/html; charset=UTF-8",
		"Content-Transfer-Encoding: 7bit",
		"",
		html,
		"",
		fmt.Sprintf("--%s", boundary),
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 7bit",
		"",
		html,
		"",
		fmt.Sprintf("--%s--", boundary),
	}, "\r\n")

	return smtp.SendMail(email.addr, email.auth, from, to, []byte(msg))
}
