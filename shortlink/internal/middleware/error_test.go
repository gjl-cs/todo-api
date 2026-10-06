package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestErrorHandlerRecoversPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandler())

	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			recorder.Code,
		)
	}

	var body struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}

	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	if body.Code != http.StatusInternalServerError {
		t.Errorf(
			"expected JSON code %d, got %d",
			http.StatusInternalServerError,
			body.Code,
		)
	}

	if body.Message != "internal server error" {
		t.Errorf(
			"expected message %q, got %q",
			"internal server error",
			body.Message,
		)
	}

	if body.Data != nil {
		t.Errorf("expected data to be null, got %v", body.Data)
	}
}
