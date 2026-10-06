package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"shortlink/internal/metrics"
	"shortlink/internal/response"
	shortlinkservice "shortlink/internal/service"
	"shortlink/internal/utils"
)

// ShortLinkService 定义 Handler 所需的业务操作。
type ShortLinkService interface {
	CreateShortLink(url string) (string, error)
	GetLongURL(code string) (string, error)
	RecordVisit(code string) error
	GetPV(code string) (int64, error)
}

// ShortLinkHandler 处理短链接 HTTP 请求。
type ShortLinkHandler struct {
	service ShortLinkService
}

// NewShortLinkHandler 创建 Handler。
func NewShortLinkHandler(
	service ShortLinkService,
) *ShortLinkHandler {
	return &ShortLinkHandler{
		service: service,
	}
}

// CreateShortLinkRequest 创建短链接的请求参数。
type CreateShortLinkRequest struct {
	URL string `json:"url" binding:"required"`
}

// Create 创建短链接。
func (h *ShortLinkHandler) Create(c *gin.Context) {
	var req CreateShortLinkRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid request",
		)
		return
	}

	code, err := h.service.CreateShortLink(req.URL)
	if err != nil {
		// URL 格式错误属于客户端输入错误。
		if errors.Is(err, utils.ErrInvalidLongURL) {
			response.Error(
				c,
				http.StatusBadRequest,
				"invalid URL: only valid http and https URLs are allowed",
			)
			return
		}

		// 其他错误暂时统一返回服务器内部错误。
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to create short link",
		)
		return
	}

	response.Success(c, gin.H{
		"code": code,
	})
}

// Redirect 根据短码重定向到原始 URL。
func (h *ShortLinkHandler) Redirect(c *gin.Context) {
	code := c.Param("code")

	// 1. 查询原始 URL。
	longURL, err := h.service.GetLongURL(code)
	if err != nil {
		// 2. 明确不存在：返回 404。
		if errors.Is(err, shortlinkservice.ErrShortLinkNotFound) {
			response.Error(
				c,
				http.StatusNotFound,
				"short link not found",
			)
			return
		}

		// 3. 其他查询错误：返回 500。
		log.Printf("get long URL failed, code=%s, error=%v", code, err)

		response.Error(
			c,
			http.StatusInternalServerError,
			"internal server error",
		)
		return
	}

	// 4. 记录访问量。
	// 统计失败只记录日志，不阻止跳转。
	if err := h.service.RecordVisit(code); err != nil {
		log.Printf(
			"record visit failed, code=%s, error=%v",
			code,
			err,
		)
	}

	// 5. 跳转到原始 URL。
	metrics.RedirectsTotal.Inc()
	c.Redirect(http.StatusFound, longURL)
}

// Stats 查询短链接统计信息。
func (h *ShortLinkHandler) Stats(c *gin.Context) {
	code := c.Param("code")

	_, err := h.service.GetLongURL(code)
	if err != nil {
		response.Error(
			c,
			http.StatusNotFound,
			"short link not found",
		)
		return
	}

	pv, err := h.service.GetPV(code)
	if err != nil {
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to get statistics",
		)
		return
	}

	response.Success(c, gin.H{
		"code": code,
		"pv":   pv,
	})
}
