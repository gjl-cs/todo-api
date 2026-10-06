package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"shortlink/internal/cache"
	"shortlink/internal/repository"
	"shortlink/internal/router"
    _ "net/http/pprof"
	"net/http"
	"time"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("warning: .env file not found")
	}
	db, err := repository.NewDB()
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()
	redisClient := cache.NewRedis()
	if err := redisClient.Ping(cache.Ctx).Err(); err != nil {
		log.Fatal("failed to connect to Redis:", err)
	}
	go func() {
		log.Println("pprof server started at :6060")
		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.Println("pprof server stopped:", err)
		}
	}()
	log.Println("Gin mode:", gin.Mode())
	log.Println("redis connected")
	r := router.SetupRouter(db, redisClient)
	log.Println("server started at :8080")
	server := &http.Server{
    Addr:              ":8080",
    Handler:           r,
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       10 * time.Second,
    WriteTimeout:      10 * time.Second,
    IdleTimeout:       60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
