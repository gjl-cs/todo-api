package middleware

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"time"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		cost := time.Since(start)
		fmt.Printf(
			"[%s] %s %v\n",
			c.Request.Method,
			c.Request.URL.Path,
			cost,
		)
	}
}