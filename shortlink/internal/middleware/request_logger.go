package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// generateRequestID 生成随机请求 ID。
func generateRequestID() (string, error) {
	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

// isValidRequestID 校验客户端传入的请求 ID。
func isValidRequestID(id string) bool {
	if len(id) == 0 || len(id) > 64 {
		return false
	}

	for _, ch := range id {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' ||
			ch == '_') {
			return false
		}
	}

	return true
}

// RequestLogger 记录每个请求的基本信息。
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))

		// 客户端没有传入，或者传入的格式不合法，就重新生成。
		if !isValidRequestID(requestID) {
			var err error

			requestID, err = generateRequestID()
			if err != nil {
				log.Printf("failed to generate request ID: %v", err)

				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					gin.H{
						"code":    http.StatusInternalServerError,
						"message": "internal server error",
						"data":    nil,
					},
				)
				return
			}
		}

		// 将请求 ID 返回给客户端。
		c.Header("X-Request-ID", requestID)

		// 将请求 ID 保存到 Gin 上下文，方便其他中间件和 Handler 使用。
		c.Set("request_id", requestID)

		c.Next()

		duration := time.Since(start)

		log.Printf(
			"request_id=%s method=%s path=%s status=%d duration=%s client_ip=%s",
			requestID,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
			c.ClientIP(),
		)
	}
}
