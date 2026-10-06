package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"shortlink/internal/utils"
)

type mockHandlerService struct {
	code       string
	createErr  error
	createCall int
}

func (m *mockHandlerService) CreateShortLink(url string) (string, error) {
	m.createCall++

	if m.createErr != nil {
		return "", m.createErr
	}

	return m.code, nil
}

func (m *mockHandlerService) GetLongURL(code string) (string, error) {
	return "", errors.New("not implemented")
}

func (m *mockHandlerService) RecordVisit(code string) error {
	return nil
}

func (m *mockHandlerService) GetPV(code string) (int64, error) {
	return 0, nil
}

func setupCreateHandler(service ShortLinkService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := NewShortLinkHandler(service)
	router.POST("/api/v1/short-links", handler.Create)

	return router
}

func TestCreateShortLinkSuccess(t *testing.T) {
	mockService := &mockHandlerService{
		code: "abc123",
	}

	router := setupCreateHandler(mockService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/short-links",
		strings.NewReader(`{"url":"https://example.com"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"got status %d, want %d",
			recorder.Code,
			http.StatusOK,
		)
	}

	if mockService.createCall != 1 {
		t.Fatalf(
			"CreateShortLink called %d times, want 1",
			mockService.createCall,
		)
	}

	var body struct {
		Code    int            `json:"code"`
		Message string         `json:"message"`
		Data    map[string]any `json:"data"`
	}

	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if body.Code != 0 {
		t.Errorf("got response code %d, want 0", body.Code)
	}

	if body.Data["code"] != "abc123" {
		t.Errorf("got short code %v, want abc123", body.Data["code"])
	}
}

func TestCreateShortLinkInvalidURL(t *testing.T) {
	mockService := &mockHandlerService{
		createErr: utils.ErrInvalidLongURL,
	}

	router := setupCreateHandler(mockService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/short-links",
		strings.NewReader(`{"url":"javascript:alert(1)"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"got status %d, want %d",
			recorder.Code,
			http.StatusBadRequest,
		)
	}
}

func TestCreateShortLinkServiceError(t *testing.T) {
	mockService := &mockHandlerService{
		createErr: errors.New("database unavailable"),
	}

	router := setupCreateHandler(mockService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/short-links",
		strings.NewReader(`{"url":"https://example.com"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"got status %d, want %d",
			recorder.Code,
			http.StatusInternalServerError,
		)
	}
}
