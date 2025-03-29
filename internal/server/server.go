package server

import (
	"fmt"

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

type routerWrapper struct {
	Router
	gin.IRouter
}

type NewServerFunc func(port int) ServerInterface

var DefaultNew NewServerFunc = func(port int) ServerInterface {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

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

func (srv *Server) registerRoutes() {
	routes.Register(srv.router.(gin.IRouter))
}

func (srv *Server) Start() error {
	addr := fmt.Sprintf(":%d", srv.port)

	return srv.router.Run(addr)
}
