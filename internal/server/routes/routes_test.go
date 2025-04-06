package routes

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterSuccess(t *testing.T) {
	router := gin.New()

	RegisterRoutes(router)
}
