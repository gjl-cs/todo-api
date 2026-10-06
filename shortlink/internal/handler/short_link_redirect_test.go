package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// redirectTestService 是专门用于测试跳转功能的模拟服务。
type redirectTestService struct {
	longURL        string
	recordVisitErr error
}

func (s *redirectTestService) CreateShortLink(url string) (string, error) {
	return "abc123", nil
}

func (s *redirectTestService) GetLongURL(code string) (string, error) {
	return s.longURL, nil
}

func (s *redirectTestService) RecordVisit(code string) error {
	return s.recordVisitErr
}

func (s *redirectTestService) GetPV(code string) (int64, error) {
	return 0, nil
}

// 测试访问量记录失败时，短链接仍然可以正常跳转。
func TestRedirectSucceedsWhenPVRecordingFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &redirectTestService{
		longURL:        "https://example.com",
		recordVisitErr: errors.New("redis connection failed"),
	}

	h := NewShortLinkHandler(mockService)

	router := gin.New()
	router.GET("/:code", h.Redirect)

	req := httptest.NewRequest(
		http.MethodGet,
		"/abc123",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	// 即使访问量记录失败，也应该正常返回 302。
	if recorder.Code != http.StatusFound {
		t.Fatalf(
			"状态码错误：期望 %d，实际 %d",
			http.StatusFound,
			recorder.Code,
		)
	}

	// 检查跳转目标是否正确。
	gotURL := recorder.Header().Get("Location")
	wantURL := "https://example.com"

	if gotURL != wantURL {
		t.Fatalf(
			"跳转地址错误：期望 %q，实际 %q",
			wantURL,
			gotURL,
		)
	}
}
