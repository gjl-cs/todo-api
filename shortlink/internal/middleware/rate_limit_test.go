package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// 创建测试用 Redis 和 Gin 路由。
func setupRateLimitTest(
	t *testing.T,
	limit int64,
) (*miniredis.Miniredis, *redis.Client, *gin.Engine) {
	t.Helper()

	mockRedis := miniredis.RunT(t)

	redisClient := redis.NewClient(&redis.Options{
		Addr: mockRedis.Addr(),
	})

	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RateLimit(redisClient, limit, time.Minute))

	router.POST("/api/v1/short-links", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{
			"message": "created",
		})
	})

	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	return mockRedis, redisClient, router
}

// 验证未超限时，请求能够正常执行。
func TestRateLimitAllowsRequestsWithinLimit(t *testing.T) {
	_, _, router := setupRateLimitTest(t, 2)

	for i := 1; i <= 2; i++ {
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/short-links",
			nil,
		)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusCreated {
			t.Fatalf(
				"request %d: got status %d, want %d",
				i,
				recorder.Code,
				http.StatusCreated,
			)
		}
	}
}

// 验证超过限制后，服务器返回 HTTP 429。
func TestRateLimitRejectsExcessRequests(t *testing.T) {
	_, _, router := setupRateLimitTest(t, 2)

	// 前两次请求应该通过。
	for i := 1; i <= 2; i++ {
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/short-links",
			nil,
		)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusCreated {
			t.Fatalf(
				"request %d: got status %d, want %d",
				i,
				recorder.Code,
				http.StatusCreated,
			)
		}
	}

	// 第三次请求应该被拒绝。
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/short-links",
		nil,
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"got status %d, want %d",
			recorder.Code,
			http.StatusTooManyRequests,
		)
	}

	if recorder.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
}

// 验证 Redis 故障时，请求不会继续执行 Handler。
func TestRateLimitReturnsServiceUnavailableOnRedisError(t *testing.T) {
	_, redisClient, router := setupRateLimitTest(t, 2)

	// 关闭模拟 Redis，使后续请求无法执行 Redis 命令。
	if err := redisClient.Close(); err != nil {
		t.Fatalf("failed to close Redis client: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/short-links",
		nil,
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"got status %d, want %d",
			recorder.Code,
			http.StatusServiceUnavailable,
		)
	}
}
