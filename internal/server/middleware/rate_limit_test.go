package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	makeReq := func(r *gin.Engine, ip string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		if ip != "" {
			req.Header.Set("X-Forwarded-For", ip)
		}

		r.ServeHTTP(w, req)

		return w.Code
	}

	t.Run("basic flow", func(t *testing.T) {
		t.Parallel()

		r := gin.New()
		r.Use(RateLimit(2, time.Second))
		r.GET("/", func(c *gin.Context) { c.Status(200) })

		assert.Equal(t, 200, makeReq(r, "1.1.1.1"))
		assert.Equal(t, 200, makeReq(r, "1.1.1.1"))
		assert.Equal(t, 429, makeReq(r, "1.1.1.1"))

		assert.Equal(t, 200, makeReq(r, "2.2.2.2"))
	})

	t.Run("refill", func(t *testing.T) {
		t.Parallel()

		r := gin.New()
		r.Use(RateLimit(1, 100*time.Millisecond))
		r.GET("/", func(c *gin.Context) { c.Status(200) })

		assert.Equal(t, 200, makeReq(r, "1.1.1.1"))
		assert.Equal(t, 429, makeReq(r, "1.1.1.1"))
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, 200, makeReq(r, "1.1.1.1"))
	})

}

func TestGetClientIP(t *testing.T) {
	t.Parallel()

	makeContext := func(req *http.Request) *gin.Context {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		return c
	}

	tests := []struct {
		name string
		req  *http.Request
		want string
	}{
		{
			name: "x-real-ip",
			req: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("X-Real-IP", "1.1.1.1")
				return req
			}(),
			want: "1.1.1.1",
		},
		{
			name: "remote-addr",
			req: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.RemoteAddr = "1.1.1.1:1234"
				return req
			}(),
			want: "1.1.1.1",
		},
		{
			name: "no ip",
			req: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.RemoteAddr = ""
				return req
			}(),
			want: "unknown",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := makeContext(tt.req)
			assert.Equal(t, tt.want, getClientIP(c))
		})
	}
}
