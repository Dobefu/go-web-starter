package middleware

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/Dobefu/go-web-starter/internal/message"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockSession struct {
	flashes   []any
	saveError bool
	values    map[any]any
}

func newMockSession() *mockSession {
	return &mockSession{
		values: make(map[any]any),
	}
}

func (m *mockSession) ID() string {
	return "mock-session-id"
}

func (m *mockSession) Get(key any) any {
	return m.values[key]
}

func (m *mockSession) Set(key any, val any) {
	m.values[key] = val
}

func (m *mockSession) Delete(key any) {
	delete(m.values, key)
}

func (m *mockSession) Clear() {
	m.values = make(map[any]any)
}

func (m *mockSession) AddFlash(value any, vars ...string) {
	m.flashes = append(m.flashes, value)
}

func (m *mockSession) Flashes(vars ...string) []any {
	flashes := m.flashes
	m.flashes = nil

	return flashes
}

func (m *mockSession) Options(options sessions.Options) {
	// No-op for testing purposes.
}

func (m *mockSession) Save() error {
	if m.saveError {
		return errors.New("mock save error")
	}
	return nil
}

func setupTestContext() (*gin.Context, *mockSession) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockSess := newMockSession()
	c.Set("github.com/gin-contrib/sessions", mockSess)

	return c, mockSess
}

func TestGetFlashMessage(t *testing.T) {
	t.Run("No flash messages", func(t *testing.T) {
		c, _ := setupTestContext()
		message := GetFlashMessage(c)
		assert.Empty(t, message)
	})

	t.Run("With a flash message", func(t *testing.T) {
		c, mockSess := setupTestContext()
		mockSess.AddFlash("test message")

		message := GetFlashMessage(c)
		assert.Equal(t, "test message", message)
	})

	t.Run("Non-string flash message", func(t *testing.T) {
		c, mockSess := setupTestContext()
		mockSess.AddFlash(123)

		message := GetFlashMessage(c)
		assert.Empty(t, message)
	})

	t.Run("Session save error", func(t *testing.T) {
		c, mockSess := setupTestContext()
		mockSess.saveError = true
		mockSess.AddFlash("test message")

		message := GetFlashMessage(c)
		assert.Equal(t, "test message", message)
	})
}

func TestFlashMiddleware(t *testing.T) {
	t.Run("Add and get flash message", func(t *testing.T) {
		c, mockSess := setupTestContext()

		handler := Flash()
		handler(c)

		addFlash := c.MustGet("AddFlash").(func(message.Message))
		addFlash(message.Message{Body: "test message"})

		getFlash := c.MustGet("GetFlash").(func() message.Message)
		msg := getFlash()

		assert.Equal(t, "test message", msg.Body)
		assert.Empty(t, mockSess.flashes)
	})

	t.Run("Get flash message after clear", func(t *testing.T) {
		c, mockSess := setupTestContext()

		handler := Flash()
		handler(c)

		addFlash := c.MustGet("AddFlash").(func(message.Message))
		addFlash(message.Message{Body: "test message"})

		getFlash := c.MustGet("GetFlash").(func() message.Message)
		message1 := getFlash()
		message2 := getFlash()

		assert.Equal(t, "test message", message1.Body)
		assert.Empty(t, message2)
		assert.Empty(t, mockSess.flashes)
	})

	t.Run("Non-string flash message", func(t *testing.T) {
		c, mockSess := setupTestContext()
		mockSess.AddFlash(123)

		handler := Flash()
		handler(c)

		getFlash := c.MustGet("GetFlash").(func() message.Message)
		message := getFlash()

		assert.Empty(t, message.Body)
	})

	t.Run("Session save errors", func(t *testing.T) {
		c, mockSess := setupTestContext()
		mockSess.saveError = true

		handler := Flash()
		handler(c)

		addFlash := c.MustGet("AddFlash").(func(message.Message))
		addFlash(message.Message{Body: "test message"})

		getFlash := c.MustGet("GetFlash").(func() message.Message)
		message := getFlash()

		assert.Equal(t, "test message", message.Body)
	})
}
