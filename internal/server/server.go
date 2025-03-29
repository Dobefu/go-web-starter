package server

import (
	"fmt"

	"github.com/Dobefu/go-web-starter/internal/server/routes"
	"github.com/gin-gonic/gin"
)

type Router interface {
	gin.IRouter
	Run(addr ...string) error
}

type ServerInterface interface {
	Start() error
}

type Server struct {
	router Router
	port   int
}

func New(port int) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	srv := &Server{
		router: router,
		port:   port,
	}

	srv.registerRoutes()
	return srv
}

func (srv *Server) registerRoutes() {
	routes.Register(srv.router)
}

func (srv *Server) Start() error {
	addr := fmt.Sprintf(":%d", srv.port)

	return srv.router.Run(addr)
}
