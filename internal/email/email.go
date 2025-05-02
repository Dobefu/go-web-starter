package email

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strings"
	"text/template"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	server_utils "github.com/Dobefu/go-web-starter/internal/server/utils"
	"github.com/Dobefu/go-web-starter/internal/static"
	"github.com/Dobefu/go-web-starter/internal/templates"
	"github.com/spf13/viper"
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
	t := template.New(body.Template)
	t.Funcs(server_utils.TemplateFuncMap())

	t, err := t.ParseFS(
		templates.TemplateFS,
		"email/*.gohtml",
		"email/layouts/*.gohtml",
		"components/atoms/*.gohtml",
	)

	if err != nil {
		return err
	}

	stylesheet, err := static.StaticFS.ReadFile("static/css/dist/email.css")

	if err != nil {
		return err
	}

	data := struct {
		Stylesheet string
		SiteName   string
		SiteHost   string
		Year       string
		BuildHash  string
		Data       any
	}{
		Stylesheet: string(stylesheet),
		SiteName:   viper.GetString("site.name"),
		SiteHost:   viper.GetString("site.host"),
		Year:       time.Now().Format("2006"),
		BuildHash:  config.BuildHash,
		Data:       body.Data,
	}

	if err != nil {
		return err
	}

	var tpl bytes.Buffer
	err = t.ExecuteTemplate(&tpl, body.Template, data)

	if err != nil {
		return err
	}

	msg := strings.Join([]string{
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"Content-Type: text/html",
		"",
		tpl.String(),
		"",
	}, "\r\n")

	return smtp.SendMail(email.addr, email.auth, from, to, []byte(msg))
}
