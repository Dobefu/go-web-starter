package server

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/redis"
	"github.com/Dobefu/go-web-starter/internal/server/middleware"
	"github.com/Dobefu/go-web-starter/internal/server/routes"
	server_utils "github.com/Dobefu/go-web-starter/internal/server/utils"
	"github.com/Dobefu/go-web-starter/internal/static"
	"github.com/Dobefu/go-web-starter/internal/templates"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const (
	sessionCookieName = "session"
	sessionPath       = "/"
	sessionMaxAge     = 7 * 24 * time.Hour
	sessionHttpOnly   = true
	sessionSameSite   = http.SameSiteLaxMode

	rateLimitRequests = 1000
	rateLimitWindow   = time.Minute

	errTemplatesLoad = "failed to load templates: %v"
	errStaticFSInit  = "failed to initialize static file system: %v"
	errDatabaseInit  = "failed to initialize database: %v"
	errRedisInit     = "failed to initialize Redis: %v"
	errSessionDecode = "failed to decode session secret: %v"
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

type NewServerFunc func(port int) (ServerInterface, error)

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

func defaultNew(port int) (ServerInterface, error) {
	log := logger.New(config.GetLogLevel(), os.Stdout)

	if config.GetLogLevel() == logger.TraceLevel {
		gin.SetMode(gin.DebugMode)
		log.Trace("Setting Gin to debug mode due to TRACE log level", nil)
	} else if gin.Mode() != gin.TestMode {
		gin.SetMode(gin.ReleaseMode)
		log.Trace("Setting Gin to release mode", nil)
	}

	router := gin.New()
	router.SetFuncMap(server_utils.TemplateFuncMap())
	log.Trace("Initializing router with template functions", nil)

	if err := templates.LoadTemplates(router); err != nil {
		return nil, fmt.Errorf(errTemplatesLoad, err)
	}

	log.Debug("Templates loaded successfully", nil)

	staticFS, err := static.StaticFileSystem()

	if err != nil {
		return nil, fmt.Errorf(errStaticFSInit, err)
	}

	log.Trace("Static file system initialized", nil)

	router.StaticFS("/static", staticFS)

	dbConfig := getDatabaseConfig()
	db, err := database.New(dbConfig, log)

	if err != nil {
		return nil, fmt.Errorf(errDatabaseInit, err)
	}

	redisConfig := getRedisConfig()
	var redisClient redis.RedisInterface

	if redisConfig.Enable {
		redisClient, err = redis.New(redisConfig, log)

		if err != nil {
			return nil, fmt.Errorf(errRedisInit, err)
		}

		log.Trace("Redis connection established", nil)
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

	sessionSecret := viper.GetString("session.secret")
	decodedSecret, err := base64.StdEncoding.DecodeString(sessionSecret)

	if err != nil {
		return nil, fmt.Errorf(errSessionDecode, err)
	}

	store := cookie.NewStore(decodedSecret)
	store.Options(sessions.Options{
		Path:     sessionPath,
		MaxAge:   int(sessionMaxAge.Seconds()),
		HttpOnly: sessionHttpOnly,
		Secure:   gin.Mode() == gin.ReleaseMode,
		SameSite: sessionSameSite,
	})

	router.Use(sessions.Sessions(sessionCookieName, store))

	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.Database(srv.db))
	router.Use(middleware.CSRF())
	router.Use(middleware.Flash())

	if srv.redis != nil {
		router.Use(middleware.Redis(srv.redis))
		log.Trace("Redis middleware initialized", nil)
	}

	if srv.redis != nil {
		limiter := middleware.NewRateLimiterWithRedis(srv.redis, rateLimitRequests, rateLimitWindow)
		router.Use(limiter.Middleware())
	} else {
		router.Use(middleware.RateLimit(rateLimitRequests, rateLimitWindow))
	}

	router.Use(middleware.CorsHeaders())
	router.Use(middleware.CspHeaders())
	router.Use(middleware.CacheHeaders())
	log.Trace("Middleware initialized", nil)

	router.NoRoute(routes.NotFound)
	srv.registerRoutes()
	log.Trace("Routes registered", nil)

	return srv, nil
}

var DefaultNew NewServerFunc = defaultNew

func New(port int) (ServerInterface, error) {
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
