package validator

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Dobefu/go-web-starter/internal/message"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	v := New()
	assert.NotNil(t, v)
}

func TestSetContext(t *testing.T) {
	v := New()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	v.SetContext(c)
	assert.Equal(t, c, v.context)
}

func TestValid(t *testing.T) {
	v := New()
	assert.True(t, v.Valid())

	v.AddFieldError("test", "error")
	assert.False(t, v.Valid())
}

func TestAddFieldError(t *testing.T) {
	v := New()
	v.AddFieldError("field1", "error1")
	v.AddFieldError("field1", "error2")
	v.AddFieldError("field2", "error1")

	assert.False(t, v.isValid)
	assert.Len(t, v.fieldErrors["field1"], 2)
	assert.Len(t, v.fieldErrors["field2"], 1)
	assert.Equal(t, "error1", v.fieldErrors["field1"][0])
	assert.Equal(t, "error2", v.fieldErrors["field1"][1])
	assert.Equal(t, "error1", v.fieldErrors["field2"][0])
}

func TestAddFormError(t *testing.T) {
	v := New()
	v.AddFormError("error1")
	v.AddFormError("error2")

	assert.False(t, v.isValid)
	assert.Len(t, v.formErrors, 2)
	assert.Equal(t, "error1", v.formErrors[0])
	assert.Equal(t, "error2", v.formErrors[1])
}

func TestGetFieldErrors(t *testing.T) {
	v := New()
	v.AddFieldError("field1", "error1")

	errors := v.GetFieldErrors()
	assert.Len(t, errors, 1)
	assert.Len(t, errors["field1"], 1)
	assert.Equal(t, "error1", errors["field1"][0])
}

func TestGetFormErrors(t *testing.T) {
	v := New()
	v.AddFormError("error1")

	errors := v.GetFormErrors()
	assert.Len(t, errors, 1)
	assert.Equal(t, "error1", errors[0])
}

func TestHasErrors(t *testing.T) {
	v := New()
	assert.False(t, v.HasErrors())

	v.AddFieldError("test", "error")
	assert.True(t, v.HasErrors())
}

func TestClear(t *testing.T) {
	v := New()
	v.AddFieldError("field1", "error1")
	v.AddFormError("error2")

	v.Clear()
	assert.True(t, v.isValid)
	assert.Empty(t, v.fieldErrors)
	assert.Empty(t, v.formErrors)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	store := cookie.NewStore([]byte("secret"))
	c.Request, _ = http.NewRequest("GET", "/", nil)
	sessions.Sessions("mysession", store)(c)

	v.SetContext(c)

	v.SetFormData(map[string]string{"field": "value"})
	v.fieldErrors = map[string][]string{"field": {"error"}}
	v.SetErrors()

	v.Clear()

	assert.Empty(t, v.GetFormData())
	assert.Empty(t, v.GetSessionErrors())
}

func TestCheckField(t *testing.T) {
	v := New()
	v.CheckField(false, "field1", "error1")

	assert.False(t, v.isValid)
	assert.Len(t, v.fieldErrors["field1"], 1)
	assert.Equal(t, "error1", v.fieldErrors["field1"][0])
}

func TestCheckForm(t *testing.T) {
	v := New()

	v.CheckForm(false, "error1")
	assert.False(t, v.isValid)
	assert.Len(t, v.formErrors, 1)
	assert.Equal(t, "error1", v.formErrors[0])
}

func TestRequired(t *testing.T) {
	v := New()
	v.Required("field1", "")
	assert.False(t, v.isValid)
	assert.Len(t, v.fieldErrors["field1"], 1)
	assert.Equal(t, msgFieldRequired, v.fieldErrors["field1"][0])

	v = New()
	v.Required("field1", "   ")
	assert.False(t, v.isValid)
	assert.Len(t, v.fieldErrors["field1"], 1)

	v = New()
	v.Required("field1", "value")
	assert.True(t, v.isValid)
	assert.Empty(t, v.fieldErrors)
}

func TestMinLength(t *testing.T) {
	v := New()
	v.MinLength("field1", "abc", 4)
	assert.False(t, v.isValid)
	assert.Len(t, v.fieldErrors["field1"], 1)
	assert.Equal(t, fmt.Sprintf(msgMinLength, 4), v.fieldErrors["field1"][0])

	v = New()
	v.MinLength("field1", "abcd", 4)
	assert.True(t, v.isValid)
	assert.Empty(t, v.fieldErrors)
}

func TestMaxLength(t *testing.T) {
	v := New()
	v.MaxLength("field1", "abcde", 4)
	assert.False(t, v.isValid)
	assert.Len(t, v.fieldErrors["field1"], 1)
	assert.Equal(t, fmt.Sprintf(msgMaxLength, 4), v.fieldErrors["field1"][0])

	v = New()
	v.MaxLength("field1", "abcd", 4)
	assert.True(t, v.isValid)
	assert.Empty(t, v.fieldErrors)
}

func TestPasswordsMatch(t *testing.T) {
	v := New()
	v.PasswordsMatch("field1", "pass", "bogus")
	assert.False(t, v.isValid)
	assert.Len(t, v.fieldErrors["field1"], 1)
	assert.Equal(t, msgPasswordsMatch, v.fieldErrors["field1"][0])

	v = New()
	v.PasswordsMatch("field1", "pass", "pass")
	assert.True(t, v.isValid)
	assert.Empty(t, v.fieldErrors)
}

func TestValidateForm(t *testing.T) {
	v := New()

	req := httptest.NewRequest("POST", "/", nil)
	err := v.ValidateForm(req)
	assert.NoError(t, err)
	assert.True(t, v.isValid)

	req = httptest.NewRequest("POST", "/", strings.NewReader("%"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	err = v.ValidateForm(req)
	assert.Error(t, err)
	assert.Contains(t, v.formErrors, msgFormProcessing)
}

func TestGetFormValue(t *testing.T) {
	v := New()
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"basic value", "value1", "value1"},
		{"control chars stripped", "va\x00l\x01u\x02e\x03\t2\n", "value\t2"},
		{"only control chars", "\x00\x01\x02\x03", ""},
		{"tabs and spaces", "\t \t", "\t \t"},
		{"empty value", "", ""},
	}

	req := httptest.NewRequest("POST", "/", nil)
	req.Form = make(map[string][]string)

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			field := fmt.Sprintf("field%d", (i + 1))
			req.Form[field] = []string{tc.input}

			value := v.GetFormValue(req, field)
			assert.Equal(t, tc.expected, value)
		})
	}
}

func TestSessionOperations(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	store := cookie.NewStore([]byte("secret"))
	c.Request, _ = http.NewRequest("GET", "/", nil)
	sessions.Sessions("mysession", store)(c)

	v := New()
	v.SetContext(c)

	formData := map[string]string{
		"field1": "value1",
		"field2": "value2",
	}

	v.SetFormData(formData)
	retrievedData := v.GetFormData()
	assert.Equal(t, formData, retrievedData)

	fieldErrors := map[string][]string{"field1": {"error1"}}

	v.fieldErrors = fieldErrors
	v.SetErrors()
	retrievedErrors := v.GetSessionErrors()
	assert.Equal(t, fieldErrors, retrievedErrors)

	v.ClearSession()
	assert.Empty(t, v.GetFormData())
	assert.Empty(t, v.GetSessionErrors())
}

func TestSessionErrorHandling(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	store := cookie.NewStore([]byte("secret"))
	c.Request, _ = http.NewRequest("GET", "/", nil)
	sessions.Sessions("mysession", store)(c)

	v := New()
	v.SetContext(c)

	invalidData := map[string]string{
		"field1": string([]byte{0xff, 0xfe, 0xfd}),
	}

	v.SetFormData(invalidData)
	assert.Empty(t, v.GetFormData())

	v.fieldErrors = map[string][]string{
		"field1": {string([]byte{0xff, 0xfe, 0xfd})},
	}

	v.SetErrors()
	assert.Empty(t, v.GetSessionErrors())

	v.marshal = func(interface{}) ([]byte, error) {
		return nil, fmt.Errorf("forced marshal error")
	}
	v.SetFormData(map[string]string{"field1": "value1"})
	assert.Empty(t, v.GetFormData())

	v.fieldErrors = map[string][]string{"field1": {"error1"}}
	v.SetErrors()
	assert.Empty(t, v.GetSessionErrors())
}

func TestFlashOperations(t *testing.T) {
	v := New()

	messages := v.GetMessages()
	assert.Empty(t, messages)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	store := cookie.NewStore([]byte("secret"))
	c.Request, _ = http.NewRequest("GET", "/", nil)
	sessions.Sessions("mysession", store)(c)

	v.SetContext(c)

	messages = v.GetMessages()
	assert.Empty(t, messages)

	v.SetFlash(message.Message{Body: "test message"})
	messages = v.GetMessages()
	assert.Empty(t, messages)

	session := sessions.Default(c)
	session.AddFlash(123)
	session.AddFlash(message.Message{Body: "valid message"})
	_ = session.Save()
	messages = v.GetMessages()
	assert.Len(t, messages, 1)
	assert.Equal(t, "valid message", messages[0].Body)

	c.Set("AddFlash", func(msg message.Message) {
		session := sessions.Default(c)
		session.AddFlash(msg)
		_ = session.Save()
	})

	v.SetFlash(message.Message{Body: "test message"})
	messages = v.GetMessages()
	assert.Len(t, messages, 1)
	assert.Equal(t, "test message", messages[0].Body)
}

func TestEdgeCases(t *testing.T) {
	v := New()

	v.SetFormData(map[string]string{"field": "value"})
	v.SetErrors()
	assert.Empty(t, v.GetFormData())
	assert.Empty(t, v.GetSessionErrors())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	store := cookie.NewStore([]byte("secret"))
	c.Request, _ = http.NewRequest("GET", "/", nil)
	sessions.Sessions("mysession", store)(c)
	v.SetContext(c)

	session := sessions.Default(c)
	session.Set(sessionKeyFormData, "{invalid json}")
	_ = session.Save()
	assert.Empty(t, v.GetFormData())

	session.Set(sessionKeyErrors, "{invalid json}")
	_ = session.Save()
	assert.Empty(t, v.GetSessionErrors())

	session.Set(sessionKeyErrors, nil)
	_ = session.Save()
	assert.Empty(t, v.GetSessionErrors())
}

func TestGetSessionNilContext(t *testing.T) {
	v := New()
	session := v.getSession()
	assert.Nil(t, session)
}

func TestSetFlashEdgeCases(t *testing.T) {
	v := New()

	v.SetFlash(message.Message{Body: "test message"})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	store := cookie.NewStore([]byte("secret"))
	c.Request, _ = http.NewRequest("GET", "/", nil)
	sessions.Sessions("mysession", store)(c)
	v.SetContext(c)

	c.Set("AddFlash", "not a function")
	v.SetFlash(message.Message{Body: "test message"})
}

type panicSession struct{}

func (p *panicSession) Flashes(...string) []any            { panic("forced panic for test") }
func (p *panicSession) Clear()                             {}
func (p *panicSession) Save() error                        { return nil }
func (p *panicSession) ID() string                         { return "" }
func (p *panicSession) Get(key any) any                    { return nil }
func (p *panicSession) Set(key, val any)                   {}
func (p *panicSession) Delete(key any)                     {}
func (p *panicSession) AddFlash(value any, vars ...string) {}
func (p *panicSession) Options(options sessions.Options)   {}

func TestGetMessages_RecoversFromPanic(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("github.com/gin-contrib/sessions", &panicSession{})

	v := New()
	v.SetContext(c)

	assert.NotPanics(t, func() {
		msgs := v.GetMessages()
		assert.Empty(t, msgs)
	})
}
