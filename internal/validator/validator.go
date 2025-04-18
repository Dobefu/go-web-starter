package validator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type Validator struct {
	isValid     bool
	fieldErrors map[string][]string
	formErrors  []string
	context     *gin.Context
}

func New() *Validator {
	return &Validator{
		isValid:     true,
		fieldErrors: make(map[string][]string),
		formErrors:  make([]string, 0),
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

func (v *Validator) Clear() {
	v.isValid = true
	v.fieldErrors = make(map[string][]string)
	v.formErrors = make([]string, 0)

	if v.context != nil {
		session := sessions.Default(v.context)

		session.Delete("form_data")
		session.Delete("errors")

		session.Save()
	}
}

func (v *Validator) CheckField(ok bool, field, message string) {
	if ok {
		return
	}

	v.AddFieldError(field, message)
}

func (v *Validator) CheckForm(ok bool, message string) {
	if ok {
		return
	}

	v.AddFormError(message)
}

func (v *Validator) Required(field, value string) {
	v.CheckField(strings.TrimSpace(value) != "", field, "This field is required")
}

func (v *Validator) MinLength(field, value string, min int) {
	v.CheckField(len(strings.TrimSpace(value)) >= min, field, fmt.Sprintf("This field must be at least %d characters long", min))
}

func (v *Validator) MaxLength(field, value string, max int) {
	v.CheckField(len(strings.TrimSpace(value)) <= max, field, fmt.Sprintf("This field must be no more than %d characters long", max))
}

func (v *Validator) ValidateForm(r *http.Request) error {
	err := r.ParseForm()

	if err != nil {
		v.AddFormError("Failed to process form data")
		return err
	}

	return nil
}

func (v *Validator) GetFormValue(r *http.Request, field string) string {
	return r.FormValue(field)
}

func (v *Validator) SetFlash(message string) {
	if v.context == nil {
		return
	}

	addFlash, exists := v.context.Get("AddFlash")

	if exists {
		flashFunc, ok := addFlash.(func(string))

		if ok {
			flashFunc(message)
		}
	}
}

func (v *Validator) GetMessages() []string {
	if v.context == nil {
		return make([]string, 0)
	}

	session := sessions.Default(v.context)
	messages := session.Flashes()
	result := make([]string, 0, len(messages))

	for _, msg := range messages {
		if str, ok := msg.(string); ok {
			result = append(result, str)
		}
	}

	return result
}

func (v *Validator) SetFormData(values map[string]string) {
	if v.context == nil {
		return
	}

	session := sessions.Default(v.context)
	formDataJSON, err := json.Marshal(values)

	if err != nil {
		return
	}

	session.Set("form_data", string(formDataJSON))
	session.Save()
}

func (v *Validator) SetErrors() {
	if v.context == nil {
		return
	}

	session := sessions.Default(v.context)
	errorsJSON, err := json.Marshal(v.GetFieldErrors())

	if err != nil {
		return
	}

	session.Set("errors", string(errorsJSON))
	session.Save()
}

func (v *Validator) GetFormData() map[string]string {
	if v.context == nil {
		return make(map[string]string)
	}

	session := sessions.Default(v.context)
	formDataJSON := session.Get("form_data")

	if formDataJSON == nil {
		return make(map[string]string)
	}

	var formData map[string]string
	err := json.Unmarshal([]byte(formDataJSON.(string)), &formData)

	if err != nil {
		return make(map[string]string)
	}

	return formData
}

func (v *Validator) GetSessionErrors() map[string][]string {
	if v.context == nil {
		return make(map[string][]string)
	}

	session := sessions.Default(v.context)
	errorsJSON := session.Get("errors")

	if errorsJSON == nil {
		return make(map[string][]string)
	}

	var errors map[string][]string

	if err := json.Unmarshal([]byte(errorsJSON.(string)), &errors); err != nil {
		return make(map[string][]string)
	}

	return errors
}

func (v *Validator) ClearSession() {
	if v.context == nil {
		return
	}

	session := sessions.Default(v.context)
	session.Delete("form_data")
	session.Delete("errors")
	session.Save()
}
