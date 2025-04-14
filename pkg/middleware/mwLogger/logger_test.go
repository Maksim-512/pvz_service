package mwLogger

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMwLogger(t *testing.T) {
	log := slog.Default()
	middleware := New(log)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("тест"))
	})

	middlewareHandler := middleware(handler)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("User-Agent", "Test-Agent")

	rr := httptest.NewRecorder()

	middlewareHandler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Equal(t, "тест", rr.Body.String())
}
