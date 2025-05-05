package validator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/Dobefu/go-web-starter/internal/message"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	sessionKeyFormData = "form_data"
	sessionKeyErrors   = "errors"

	msgFieldRequired  = "This field is required"
	msgFormProcessing = "Failed to process form data"
	msgMinLength      = "This field must be at least %d characters long"
	msgMaxLength      = "This field must be no more than %d characters long"
	msgEmailInvalid   = "This is not a valid email address"
	msgPasswordsMatch = "The passwords do not match"
)

type Validator struct {
	isValid     bool
	fieldErrors map[string][]string
	formErrors  []string
	context     *gin.Context
	marshal     func(any) ([]byte, error)
}

func New() *Validator {
	return &Validator{
		isValid:     true,
		fieldErrors: make(map[string][]string),
		formErrors:  make([]string, 0),
		marshal:     json.Marshal,
	}
}

func (v *Validator) SetContext(c *gin.Context) {
	v.context = c
}

func (v *Validator) Valid() bool {
	return v.isValid
}

func (v *Validator) AddFieldError(field, message string) {
	_, exists := v.fieldErrors[field]

	if !exists {
		v.fieldErrors[field] = make([]string, 0)
	}

	v.fieldErrors[field] = append(v.fieldErrors[field], message)
	v.isValid = false
}

func (v *Validator) AddFormError(message string) {
	v.formErrors = append(v.formErrors, message)
	v.isValid = false
}

func (v *Validator) GetFieldErrors() map[string][]string {
	return v.fieldErrors
}

func (v *Validator) GetFormErrors() []string {
	return v.formErrors
}

func (v *Validator) HasErrors() bool {
	return !v.isValid
}

func (v *Validator) getSession() sessions.Session {
	if v.context == nil {
		return nil
	}

	return sessions.Default(v.context)
}

func isValidUTF8Map(values map[string]string) bool {
	for _, value := range values {
		if !utf8.ValidString(value) {
			return false
		}
	}

	return true
}

func isValidUTF8Errors(errors map[string][]string) bool {
	for _, errList := range errors {
		for _, err := range errList {
			if !utf8.ValidString(err) {
				return false
			}
		}
	}

	return true
}

func (v *Validator) Clear() {
	v.isValid = true
	v.fieldErrors = make(map[string][]string)
	v.formErrors = make([]string, 0)

	if session := v.getSession(); session != nil {
		session.Delete(sessionKeyFormData)
		session.Delete(sessionKeyErrors)
		_ = session.Save()
	}
}

func (v *Validator) CheckField(ok bool, field, message string) {
	if !ok {
		v.AddFieldError(field, message)
	}
}

func (v *Validator) CheckForm(ok bool, message string) {
	if !ok {
		v.AddFormError(message)
	}
}

func (v *Validator) Required(field, value string) {
	v.CheckField(strings.TrimSpace(value) != "", field, msgFieldRequired)
}

func (v *Validator) MinLength(field, value string, min int) {
	fmt.Println(field, value, min)
	v.CheckField(len(strings.TrimSpace(value)) >= min, field, fmt.Sprintf(msgMinLength, min))
}

func (v *Validator) MaxLength(field, value string, max int) {
	v.CheckField(len(strings.TrimSpace(value)) <= max, field, fmt.Sprintf(msgMaxLength, max))
}

func (v *Validator) ValidEmail(field, value string) {
	_, err := mail.ParseAddress(strings.TrimSpace(value))
	v.CheckField(err == nil, field, msgEmailInvalid)
}

func (v *Validator) PasswordsMatch(field, password1 string, password2 string) {
	v.CheckField(password1 == password2, field, msgPasswordsMatch)
}

func (v *Validator) ValidateForm(r *http.Request) error {
	err := r.ParseForm()

	if err != nil {
		v.AddFormError(msgFormProcessing)
		return err
	}

	return nil
}

func (v *Validator) GetFormValue(r *http.Request, field string) string {
	rawInput := r.FormValue(field)
	var b strings.Builder

	// Strip out any control characters except for tabs and spaces.
	for _, r := range rawInput {
		if r < 32 && r != '\t' {
			continue
		}

		b.WriteRune(r)
	}

	return b.String()
}

func (v *Validator) SetFlash(msg message.Message) {
	if v.context == nil {
		return
	}

	addFlash, exists := v.context.Get("AddFlash")

	if exists {
		flashFunc, ok := addFlash.(func(message.Message))

		if ok {
			flashFunc(msg)
		}
	}
}

func (v *Validator) GetMessages() []message.Message {
	session := v.getSession()

	if session == nil {
		return make([]message.Message, 0)
	}

	var messages []any

	defer func() {
		if r := recover(); r != nil {
			session.Clear()
			_ = session.Save()

			messages = nil
		}
	}()

	messages = session.Flashes()
	result := make([]message.Message, 0, len(messages))

	for _, msg := range messages {
		val, ok := msg.(message.Message)

		if ok {
			result = append(result, val)
		}
	}

	_ = session.Save()

	return result
}

func (v *Validator) SetFormData(values map[string]string) {
	session := v.getSession()

	if session == nil || !isValidUTF8Map(values) {
		return
	}

	formDataJSON, err := v.marshal(values)

	if err != nil {
		return
	}

	session.Set(sessionKeyFormData, string(formDataJSON))
	_ = session.Save()
}

func (v *Validator) SetErrors() {
	session := v.getSession()

	if session == nil || !isValidUTF8Errors(v.GetFieldErrors()) {
		return
	}

	errorsJSON, err := v.marshal(v.GetFieldErrors())
	if err != nil {
		return
	}

	session.Set(sessionKeyErrors, string(errorsJSON))
	_ = session.Save()
}

func (v *Validator) GetFormData() map[string]string {
	session := v.getSession()
	if session == nil {
		return make(map[string]string)
	}

	formDataJSON := session.Get(sessionKeyFormData)
	if formDataJSON == nil {
		return make(map[string]string)
	}

	var formData map[string]string
	err := json.Unmarshal([]byte(formDataJSON.(string)), &formData)

	if err != nil {
		return make(map[string]string)
	}

	session.Delete(sessionKeyFormData)
	_ = session.Save()

	return formData
}

func (v *Validator) GetSessionErrors() map[string][]string {
	session := v.getSession()
	if session == nil {
		return make(map[string][]string)
	}

	errorsJSON := session.Get(sessionKeyErrors)
	if errorsJSON == nil {
		return make(map[string][]string)
	}

	var errors map[string][]string

	if err := json.Unmarshal([]byte(errorsJSON.(string)), &errors); err != nil {
		return make(map[string][]string)
	}

	session.Delete(sessionKeyErrors)
	_ = session.Save()

	return errors
}

func (v *Validator) ClearSession() {
	session := v.getSession()

	if session != nil {
		session.Delete(sessionKeyFormData)
		session.Delete(sessionKeyErrors)
		session.Flashes()

		_ = session.Save()
	}
}
