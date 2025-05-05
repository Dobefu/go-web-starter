package routes

import (
	"net/http"
	"os"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/message"
	route_utils "github.com/Dobefu/go-web-starter/internal/server/routes/utils"
	"github.com/Dobefu/go-web-starter/internal/user"
	"github.com/Dobefu/go-web-starter/internal/validator"
	"github.com/gin-gonic/gin"
)

const (
	errUserAccountVerify = "Could not verify the user account."

	routeRegister = "/register"
)

func RegisterVerify(c *gin.Context) {
	log := logger.New(config.GetLogLevel(), os.Stdout)
	v := validator.New()
	v.SetContext(c)

	email := v.GetFormValue(c.Request, "email")
	token := v.GetFormValue(c.Request, "token")

	// If the route was visited without an email address,
	// redirect to the registration page.
	if len(email) <= 0 {
		c.Redirect(http.StatusSeeOther, routeRegister)
		return
	}

	data := RouteData{
		Template:    "pages/register_verify",
		HttpStatus:  http.StatusOK,
		Title:       "Verify email address",
		Description: "Verify your email address.",
		Data: map[string]any{
			"Email": email,
			"Token": token,
		},
	}

	// If the token is missing, Show the page with the email verification message.
	if len(token) <= 0 {
		RenderRouteHTML(c, data)
		return
	}

	db, err := route_utils.GetDbFromContext(c)

	if err != nil {
		log.Error("Could not get the database from the context", logger.Fields{"err": err.Error()})
		RenderRouteHTML(c, GenericErrorData(c))

		return
	}

	usr, err := user.FindByEmail(db, email)

	if err != nil {
		log.Warn("Could not get the user from the email address", logger.Fields{"err": err.Error()})

		v.SetFlash(message.Message{
			Type: message.MessageTypeError,
			Body: errUserAccountVerify,
		})

		c.Redirect(http.StatusSeeOther, routeRegister)
		return
	}

	expectedToken := usr.CreateVerificationToken()

	if token != expectedToken || usr.GetStatus() {
		v.SetFlash(message.Message{
			Type: message.MessageTypeError,
			Body: errUserAccountVerify,
		})

		c.Redirect(http.StatusSeeOther, routeRegister)
		return
	}

	usr.SetStatus(true)
	err = usr.Save(db)

	if err != nil {
		log.Error("Failed to update user status", logger.Fields{"err": err.Error()})
		RenderRouteHTML(c, GenericErrorData(c))

		return
	}

	session := getSession(c)
	err = usr.Login(session)

	if err != nil {
		log.Error("Failed to save session after verification login", logger.Fields{"err": err.Error()})
		RenderRouteHTML(c, GenericErrorData(c))

		return
	}

	v.SetFlash(message.Message{
		Type: message.MessageTypeSuccess,
		Body: "Your email has been verified! You are now logged in.",
	})
	c.Redirect(http.StatusSeeOther, "/")
}
