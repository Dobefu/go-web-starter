package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router *gin.Engine
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

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)

	return s.router.Run(addr)
}
