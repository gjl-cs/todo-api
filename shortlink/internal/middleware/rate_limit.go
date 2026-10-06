package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimit 使用 Redis 实现固定时间窗口限流。
func RateLimit(
	redisClient *redis.Client,
	limit int64,
	window time.Duration,
) gin.HandlerFunc {

	// Lua 脚本保证递增计数和设置过期时间的操作具有原子性。
	script := redis.NewScript(`
		local current = redis.call("INCR", KEYS[1])

		if current == 1 then
			redis.call("EXPIRE", KEYS[1], ARGV[1])
		end

		return current
	`)

	return func(c *gin.Context) {
		// 每个客户端 IP 使用独立的 Redis 计数器。
		key := "rate_limit:shortlink:create:" + c.ClientIP()

		// 执行 Redis 脚本。
		count, err := script.Run(
			c.Request.Context(),
			redisClient,
			[]string{key},
			int64(window.Seconds()),
		).Int64()

		if err != nil {
			log.Printf(
				"rate limit redis error: request_id=%v error=%v",
				c.GetString("request_id"),
				err,
			)

			// Redis 异常时拒绝请求，避免限流失效后请求继续进入数据库。
			c.Header("Retry-After", "5")
			c.AbortWithStatusJSON(
				http.StatusServiceUnavailable,
				gin.H{
					"code":    http.StatusServiceUnavailable,
					"message": "rate limit service unavailable",
					"data":    nil,
				},
			)
			return
		}

		// 超出请求次数限制。
		if count > limit {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(
				http.StatusTooManyRequests,
				gin.H{
					"code":    http.StatusTooManyRequests,
					"message": "too many requests",
					"data":    nil,
				},
			)
			return
		}

		// 没有超限，继续执行后续 Handler。
		c.Next()
	}
}
