package auth_test

import (
	"context"
	"errors"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"pvz_service/pkg/jwt/mocks"
	"pvz_service/pkg/middleware/auth"
)

func makeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

func TestCheckAuthMiddleware_NoToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJwt := mocks.NewMockTokenService(ctrl)

	handler := auth.CheckAuthMiddleware(mockJwt)(makeHandler())

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "Доступ запрещен")
}

func TestCheckAuthMiddleware_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJwt := mocks.NewMockTokenService(ctrl)
	mockJwt.EXPECT().ValidateToken("bad.token").Return(nil, errors.New("invalid token"))

	handler := auth.CheckAuthMiddleware(mockJwt)(makeHandler())

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer bad.token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "Доступ запрещен")
}

func TestCheckAuthMiddleware_NoRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJwt := mocks.NewMockTokenService(ctrl)
	mockJwt.EXPECT().ValidateToken("valid.token").Return(map[string]any{}, nil)

	handler := auth.CheckAuthMiddleware(mockJwt)(makeHandler())

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer valid.token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "Доступ запрещен")
}

func TestCheckAuthMiddleware_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJwt := mocks.NewMockTokenService(ctrl)
	mockJwt.EXPECT().ValidateToken("valid.token").Return(map[string]any{"role": "admin"}, nil)

	handler := auth.CheckAuthMiddleware(mockJwt)(makeHandler())

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer valid.token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", strings.TrimSpace(rec.Body.String()))
}

func TestRoleMiddle_AllowedRole(t *testing.T) {
	middleware := auth.RoleMiddle("admin")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", "admin"))

	rr := httptest.NewRecorder()
	handlerToTest.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRoleMiddle_ForbiddenRole(t *testing.T) {
	middleware := auth.RoleMiddle("admin")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", "user"))

	rr := httptest.NewRecorder()
	handlerToTest.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.True(t, strings.Contains(rr.Body.String(), "Доступ запрещен"))
}

func TestRoleMiddle_MissingRole(t *testing.T) {
	middleware := auth.RoleMiddle("admin")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rr := httptest.NewRecorder()
	handlerToTest.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.True(t, strings.Contains(rr.Body.String(), "Доступ запрещен"))
}
