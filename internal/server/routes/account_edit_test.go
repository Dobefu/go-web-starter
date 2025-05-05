package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAccountEdit(t *testing.T) {
	router, _, mockDB, err := setupTestRouter(true)
	assert.NoError(t, err)
	defer func() { _ = mockDB.Close() }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	store := cookie.NewStore([]byte("secret"))
	c.Request, _ = http.NewRequest("GET", "/", nil)
	sessions.Sessions("mysession", store)(c)

	session := sessions.Default(c)
	session.Set("userID", 1)
	err = session.Save()
	assert.NoError(t, err)

	router.GET("/", AccountEdit)

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
}
