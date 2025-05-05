package email

import (
	"errors"
	"net/smtp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmail(t *testing.T) {
	email := New("test-host", "25", "id", "user", "pass")
	assert.Equal(t, "test-host:25", email.addr)
	assert.Implements(t, (*EmailSender)(nil), email)
}

func TestEmailSuccess(t *testing.T) {
	origSendMail := SmtpSendMail
	SmtpSendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		return nil
	}

	defer func() { SmtpSendMail = origSendMail }()

	email := New("", "", "", "", "")
	body := EmailBody{Template: "email/test_email", Data: nil}

	err := email.SendMail("from@a.com", []string{"to@b.com"}, "subject", body)
	assert.NoError(t, err)
}

func TestEmailErrTemplate(t *testing.T) {
	origSendMail := SmtpSendMail
	SmtpSendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		return nil
	}

	defer func() { SmtpSendMail = origSendMail }()

	email := New("", "", "", "", "")
	body := EmailBody{Template: "", Data: nil}

	err := email.SendMail("from@a.com", []string{"to@b.com"}, "subject", body)
	assert.Error(t, err)
}

func TestEmailErrSMTP(t *testing.T) {
	origSendMail := SmtpSendMail
	SmtpSendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		return errors.New("smtp fail")
	}

	defer func() { SmtpSendMail = origSendMail }()

	email := New("", "", "", "", "")
	body := EmailBody{Template: "email/test_email", Data: nil}

	err := email.SendMail("from@a.com", []string{"to@b.com"}, "subject", body)
	assert.Error(t, err)
}
