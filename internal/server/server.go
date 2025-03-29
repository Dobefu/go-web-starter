package server

import (
	"fmt"

	"github.com/Dobefu/go-web-starter/internal/server/middleware"
	"github.com/Dobefu/go-web-starter/internal/server/routes"
	"github.com/gin-gonic/gin"
)

type Router interface {
	Run(addr ...string) error
}

type ServerInterface interface {
	Start() error
}

type Server struct {
	router Router
	port   int
}

// routerWrapper combines the Router interface with Gin's IRouter to maintain
// a clean abstraction whilst providing access to Gin's routing capabilities.
type routerWrapper struct {
	Router
	gin.IRouter
}

type NewServerFunc func(port int) ServerInterface

var DefaultNew NewServerFunc = func(port int) ServerInterface {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.SecurityHeaders())

	srv := &Server{
		router: &routerWrapper{
			Router:  router,
			IRouter: router,
		},
		port: port,
	}

	srv.registerRoutes()
	return srv
}

func New(port int) ServerInterface {
	return DefaultNew(port)
}

// The registerRoutes function uses a type assertion to access Gin's routing capabilities.
// This is safe because routerWrapper implements gin.IRouter.
func (srv *Server) registerRoutes() {
	routes.Register(srv.router.(gin.IRouter))
}

func (srv *Server) Start() error {
	addr := fmt.Sprintf(":%d", srv.port)

	return srv.router.Run(addr)
}
