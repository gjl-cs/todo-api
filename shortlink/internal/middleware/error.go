package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"shortlink/internal/response"
)

// ErrorHandler 捕获 panic，并返回统一格式的错误响应。
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				// 记录错误及对应的请求 ID。
				requestID, _ := c.Get("request_id")

				log.Printf(
					"panic recovered: request_id=%v error=%v",
					requestID,
					recovered,
				)

				// 如果响应头已经提交，就不能再安全地写入新的 JSON 响应。
				if !c.Writer.Written() {
					response.Error(
						c,
						http.StatusInternalServerError,
						"internal server error",
					)
				}

				c.Abort()
			}
		}()

		c.Next()
	}
}
