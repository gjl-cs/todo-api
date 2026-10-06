package handler

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"

	"shortlink/internal/service"
)

func setupTestHandler() *ShortLinkHandler {
	gin.SetMode(gin.TestMode)

	repo := &mockHandlerRepository{}
	svc := service.NewShortLinkService(repo, nil)

	return NewShortLinkHandler(svc)
}

type mockHandlerRepository struct{}

func (m *mockHandlerRepository) Create(code string, longURL string) error {
	return nil
}

func (m *mockHandlerRepository) GetLongURLByCode(code string) (string, error) {
	return "https://www.google.com", nil
}

func TestCreateMissingURL(t *testing.T) {
	handler := setupTestHandler()

	router := gin.New()
	router.POST("/api/v1/short-links", handler.Create)

	body := []byte(`{}`)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/short-links",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"状态码错误: got %d, want %d",
			recorder.Code,
			http.StatusBadRequest,
		)
	}
}

func TestCreateInvalidURL(t *testing.T) {
	handler := setupTestHandler()

	router := gin.New()
	router.POST("/api/v1/short-links", handler.Create)

	body := []byte(`{"url":"not-a-valid-url"}`)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/short-links",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"状态码错误: got %d, want %d",
			recorder.Code,
			http.StatusBadRequest,
		)
	}
}

func TestCreateValidURL(t *testing.T) {
	handler := setupTestHandler()

	router := gin.New()
	router.POST("/api/v1/short-links", handler.Create)

	body := []byte(`{"url":"https://www.google.com"}`)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/short-links",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"状态码错误: got %d, want %d, response=%s",
			recorder.Code,
			http.StatusOK,
			recorder.Body.String(),
		)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatalf("响应不是有效 JSON: %v", err)
	}

	if result["message"] != "success" {
		t.Fatalf("响应 message 错误: got %v", result["message"])
	}
}
