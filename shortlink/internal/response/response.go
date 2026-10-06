package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 是统一的 HTTP JSON 响应结构。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Success 返回成功响应。
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error 返回统一格式的错误响应。
func Error(c *gin.Context, httpStatus int, message string) {
	c.JSON(httpStatus, Response{
		Code:    httpStatus,
		Message: message,
		Data:    nil,
	})
}
