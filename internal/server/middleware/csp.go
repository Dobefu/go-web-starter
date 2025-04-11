package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

const scriptSrcKey = "script-src"

type SecurityConfig struct {
	headers map[string]string
	CSP     CSPConfig
}

type CSPConfig struct {
	directives map[string][]string
}

func generateNonce() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)

	return base64.StdEncoding.EncodeToString(b)
}

func NewCSPConfig() CSPConfig {
	return CSPConfig{
		directives: map[string][]string{
			"default-src": {
				"self",
			},
			scriptSrcKey: {
				"strict-dynamic",
				"sha256-shfdQDc5l63QrdRcyAdIpEYqlgxbfEfXuTNyWpgtloM=",
			},
			"style-src": {
				"self",
			},
			"img-src": {
				"self",
				"data:",
			},
			"connect-src": {
				"self",
			},
			"font-src": {
				"self",
			},
			"object-src": {
				"none",
			},
			"base-uri": {
				"self",
			},
			"form-action": {
				"self",
			},
			"frame-ancestors": {
				"none",
			},
			"upgrade-insecure-requests": {},
		},
	}
}

func (csp CSPConfig) String() string {
	var parts []string

	for directive, values := range csp.directives {
		var quoted []string

		for _, src := range values {
			if strings.HasSuffix(src, ":") {
				quoted = append(quoted, src)
			} else {
				quoted = append(quoted, "'"+src+"'")
			}
		}

		parts = append(parts, directive+" "+strings.Join(quoted, " "))
	}

	return strings.Join(parts, "; ") + ";"
}

func newDefaultConfig() SecurityConfig {
	return SecurityConfig{
		headers: map[string]string{
			"X-Frame-Options":        "DENY",
			"X-XSS-Protection":       "1; mode=block",
			"X-Content-Type-Options": "nosniff",
			"Referrer-Policy":        "strict-origin-when-cross-origin",
		},
		CSP: NewCSPConfig(),
	}
}

func (config SecurityConfig) SetHeaders(c *gin.Context) {
	nonce := generateNonce()

	config.CSP.directives[scriptSrcKey] = append(
		config.CSP.directives[scriptSrcKey],
		fmt.Sprintf("nonce-%s", nonce),
	)

	for header, value := range config.headers {
		c.Header(header, value)
	}

	c.Header("Content-Security-Policy", config.CSP.String())
	c.Set("nonce", nonce)
}

func CspHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		newDefaultConfig().SetHeaders(c)
		c.Next()
	}
}
