package server

import (
	"fmt"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/Dobefu/go-web-starter/internal/redis"
	"github.com/Dobefu/go-web-starter/internal/server/middleware"
	"github.com/Dobefu/go-web-starter/internal/server/routes"
	server_utils "github.com/Dobefu/go-web-starter/internal/server/utils"
	"github.com/Dobefu/go-web-starter/internal/static"
	"github.com/Dobefu/go-web-starter/internal/templates"
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
	redis  redis.RedisInterface
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

func getRedisConfig() config.Redis {
	return config.Redis{
		Enable:   viper.GetBool("redis.enable"),
		Host:     viper.GetString("redis.host"),
		Port:     viper.GetInt("redis.port"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	}
}

func defaultNew(port int) ServerInterface {
	if gin.Mode() != gin.TestMode {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.SetFuncMap(server_utils.TemplateFuncMap())

	if err := templates.LoadTemplates(router); err != nil {
		panic(fmt.Sprintf("Failed to load templates: %v", err))
	}

	staticFS, err := static.StaticFileSystem()

	if err != nil {
		panic(fmt.Sprintf("Failed to initialize static file system: %v", err))
	}

	router.StaticFS("/static", staticFS)

	dbConfig := getDatabaseConfig()
	db, err := database.New(dbConfig, nil)

	if err != nil {
		panic(fmt.Sprintf("Failed to initialize database: %v", err))
	}

	redisConfig := getRedisConfig()
	var redisClient redis.RedisInterface

	if redisConfig.Enable {
		redisClient, err = redis.New(redisConfig, nil)

		if err != nil {
			panic(fmt.Sprintf("Failed to initialize Redis: %v", err))
		}
	}

	srv := &Server{
		router: &routerWrapper{
			Router:  router,
			IRouter: router,
		},
		port:  port,
		db:    db,
		redis: redisClient,
	}

	router.Use(middleware.Database(srv.db))

	if srv.redis != nil {
		router.Use(middleware.Redis(srv.redis))
	}

	router.Use(gin.Recovery())
	router.Use(middleware.RateLimit(1000, time.Minute))
	router.Use(middleware.Logger())
	router.Use(middleware.CorsHeaders())
	router.Use(middleware.CspHeaders())
	router.Use(middleware.CacheHeaders())
	router.Use(middleware.Minify())
	router.Use(middleware.DynamicContent())

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

	defer func() {
		_ = srv.db.Close()

		if srv.redis != nil {
			_ = srv.redis.Close()
		}
	}()

	return srv.router.Run(addr)
}
