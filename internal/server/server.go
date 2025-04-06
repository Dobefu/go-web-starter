package server

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/Dobefu/go-web-starter/internal/server/middleware"
	"github.com/Dobefu/go-web-starter/internal/server/routes"
	server_utils "github.com/Dobefu/go-web-starter/internal/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
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
	db     database.DatabaseInterface
}

// routerWrapper combines the Router interface with Gin's IRouter to maintain
// a clean abstraction whilst providing access to Gin's routing capabilities.
type routerWrapper struct {
	Router
	gin.IRouter
}

type NewServerFunc func(port int) ServerInterface

func getDatabaseConfig() config.Database {
	return config.Database{
		Host:     viper.GetString("database.host"),
		Port:     viper.GetInt("database.port"),
		User:     viper.GetString("database.user"),
		Password: viper.GetString("database.password"),
		DBName:   viper.GetString("database.dbname"),
	}
}

func defaultNew(port int) ServerInterface {
	if gin.Mode() != gin.TestMode {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.SetFuncMap(server_utils.TemplateFuncMap())
	templates, err := loadTemplates("templates", 0)

	if err != nil {
		panic(err)
	}

	router.LoadHTMLFiles(templates...)
	router.Static("/static", "./static")

	router.Use(middleware.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.RateLimit(1000, time.Minute))
	router.Use(middleware.CorsHeaders())
	router.Use(middleware.CspHeaders())
	router.Use(middleware.Minify())

	dbConfig := getDatabaseConfig()
	db, err := database.New(dbConfig, nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize database: %v", err))
	}

	srv := &Server{
		router: &routerWrapper{
			Router:  router,
			IRouter: router,
		},
		port: port,
		db:   db,
	}

	router.Use(middleware.Database(srv.db))

	router.NoRoute(routes.NotFound)
	srv.registerRoutes()
	return srv
}

var DefaultNew NewServerFunc = defaultNew

func New(port int) ServerInterface {
	return DefaultNew(port)
}

// The registerRoutes function uses a type assertion to access Gin's routing capabilities.
// This is safe because routerWrapper implements gin.IRouter.
func (srv *Server) registerRoutes() {
	routes.RegisterRoutes(srv.router.(gin.IRouter))
}

func (srv *Server) Start() error {
	addr := fmt.Sprintf(":%d", srv.port)

	defer func() { _ = srv.db.Close() }()
	return srv.router.Run(addr)
}

func loadTemplates(root string, depth int) (files []string, err error) {
	if depth > 10 {
		return files, errors.New("max recursion depth of 10 exceeded")
	}

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		fileInfo, err := os.Stat(path)

		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			if path != root {
				_, err = loadTemplates(path, depth+1)

				if err != nil {
					return err
				}
			}
		} else {
			files = append(files, path)
		}

		return err
	})

	return files, err
}
