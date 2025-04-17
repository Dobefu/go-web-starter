package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func RobotsTxt(c *gin.Context) {
	body := fmt.Sprintf(
		"User-agent: *\nAllow: /\n\nSitemap: %s/sitemap.xml",
		viper.GetString("site.host"),
	)

	c.Data(200, "text/plain", []byte(body))
}
