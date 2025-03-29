package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type Router interface {
	Run(addr ...string) error
	Use(middleware ...gin.HandlerFunc) gin.IRoutes
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

	return &Server{
		router: router,
		port:   port,
	}
}

func (srv *Server) Start() error {
	addr := fmt.Sprintf(":%d", srv.port)

	return srv.router.Run(addr)
}
