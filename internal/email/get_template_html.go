package email

import (
	"bytes"
	"html/template"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	server_utils "github.com/Dobefu/go-web-starter/internal/server/utils"
	"github.com/Dobefu/go-web-starter/internal/static"
	"github.com/Dobefu/go-web-starter/internal/templates"
	"github.com/spf13/viper"
)

func getTemplateHtml(body EmailBody) (string, error) {
	t := template.New(body.Template)
	t.Funcs(server_utils.TemplateFuncMap())

	t, err := t.ParseFS(
		templates.TemplateFS,
		"email/*.gohtml",
		"email/layouts/*.gohtml",
		"components/atoms/*.gohtml",
	)

	if err != nil {
		return "", err
	}

	stylesheet, err := static.StaticFS.ReadFile("static/css/dist/email.css")

	if err != nil {
		return "", err
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

	var tpl bytes.Buffer
	err = t.ExecuteTemplate(&tpl, body.Template, data)

	if err != nil {
		return "", err
	}

	return tpl.String(), nil
}
