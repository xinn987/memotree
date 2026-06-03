package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"memotree/server/api/internal/config"
)

func TestHealthz(t *testing.T) {
	router := NewRouter(config.Config{AppEnv: "test"})
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}
