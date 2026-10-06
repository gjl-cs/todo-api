package router

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"shortlink/internal/cache"
	"shortlink/internal/handler"
	"shortlink/internal/middleware"
	"shortlink/internal/repository"
	"shortlink/internal/service"
	"time"
)

func SetupRouter(
	db *sql.DB,
	redisClient *redis.Client,
) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Metrics())
	r.Use(
		gin.Logger(),
		middleware.ErrorHandler(),
	)
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	shortLinkRepository := repository.NewShortLinkRepository(db)
	shortLinkService := service.NewShortLinkService(
		shortLinkRepository,
		cache.NewRedisStore(redisClient),
	)
	shortLinkHandler := handler.NewShortLinkHandler(
		shortLinkService,
	)
	r.GET("health", handler.Health)
	r.POST(
		"/api/v1/short-links",
		middleware.RateLimit(redisClient, 10, time.Minute),
		shortLinkHandler.Create,
	)
	r.GET(
		"api/v1/short-links/:code/stats",
		shortLinkHandler.Stats,
	)
	r.GET(
		"/:code",
		shortLinkHandler.Redirect,
	)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	return r
}
