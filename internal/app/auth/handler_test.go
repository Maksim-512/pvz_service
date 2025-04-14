package auth_test

import (
	"bytes"

	"context"
	"encoding/json"
	"errors"
	"go.uber.org/mock/gomock"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pvz_service/internal/app/auth"
	authmock "pvz_service/internal/app/auth/mocks"
	"pvz_service/pkg/jwt"
)

func TestHandler_HandlerRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := authmock.NewMockServiceInterface(ctrl)
	logger := slog.Default()

	handler := auth.NewAuthHandler(logger, mockService)

	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:        "Успешная регистрация",
			requestBody: `{"email":"employee1@yandex.ru", "password":"12344321", "role":"employee"}`,
			mockSetup: func() {
				mockService.EXPECT().ServiceCreateUser(gomock.Any(), gomock.Any()).Return(&auth.User{
					ID:    uuid.MustParse("5cf259be-101b-4cbe-9846-77a3d437ec68"),
					Email: "employee1@yandex.ru",
					Role:  "employee",
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"id":    "5cf259be-101b-4cbe-9846-77a3d437ec68",
				"email": "employee1@yandex.ru",
				"role":  "employee",
			},
		},
		{
			name:           "Некорректный формат JSON",
			requestBody:    `{"email": "employee1@yandex.ru", "password": "12344321", "rolesss": }`,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"message": "Неверный запрос"},
		},
		{
			name:           "Пустое тело запроса",
			requestBody:    ``,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"message": "Неверный запрос"},
		},
		{
			name:        "Отсутствуют обязательные поля (только email)",
			requestBody: `{"email": "tests@yandex.ru"}`,
			mockSetup: func() {
				mockService.EXPECT().ServiceCreateUser(gomock.Any(), gomock.Any()).Return(
					nil,
					errors.New("пользователь с таким email уже существует"),
				)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"message": "Неверный запрос"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockSetup()

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(test.requestBody))
			req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "tests-request-id"))

			w := httptest.NewRecorder()

			handler.HandlerRegister(w, req)

			assert.Equal(t, test.expectedStatus, w.Result().StatusCode)

			res := w.Result()
			defer res.Body.Close()

			if test.expectedBody != nil {
				var responseBody map[string]interface{}
				err := json.NewDecoder(res.Body).Decode(&responseBody)
				require.NoError(t, err)
				assert.Equal(t, test.expectedBody, responseBody)
			}
		})
	}
}

func TestHandler_HandlerLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := authmock.NewMockServiceInterface(ctrl)
	logger := slog.Default()
	handler := auth.NewAuthHandler(logger, mockService)

	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:        "Успешный логин",
			requestBody: `{"email":"test@yandex.ru","password":"12345678"}`,
			mockSetup: func() {
				mockService.EXPECT().ServiceLoginUser(gomock.Any(), gomock.Any()).Return(&jwt.TokenResponse{
					Token: "jwt-token-string",
				},
					nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{"token": "jwt-token-string"},
		},
		{
			name:           "Некорректный JSON",
			requestBody:    `{"email": "test@yandex.ru", "password": }`,
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   map[string]interface{}{"message": "Неверные учетные данные"},
		},
		{
			name:        "Ошибка авторизации",
			requestBody: `{"email":"test@yandex.ru","password":"invalidpass"}`,
			mockSetup: func() {
				mockService.EXPECT().ServiceLoginUser(gomock.Any(), gomock.Any()).Return(nil, errors.New("ошибка авторизации"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   map[string]interface{}{"message": "Неверные учетные данные"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(tt.requestBody))
			req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "req-id"))
			w := httptest.NewRecorder()

			handler.HandlerLogin(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var body map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&body)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, body)
		})
	}
}

func TestHandler_HandlerDummyLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := authmock.NewMockServiceInterface(ctrl)
	logger := slog.Default()
	handler := auth.NewAuthHandler(logger, mockService)

	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:        "Успешный dummy логин",
			requestBody: `{"role": "employee"}`,
			mockSetup: func() {
				mockService.EXPECT().ServiceDummyLogin(gomock.Any(), &auth.DummyLoginRequest{
					Role: "employee",
				}).Return(&jwt.TokenResponse{
					Token: "dummy-jwt-token",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{"token": "dummy-jwt-token"},
		},
		{
			name:           "Некорректный JSON",
			requestBody:    `{"role": }`,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"message": "Неверный запрос"},
		},
		{
			name:        "Ошибка авторизации dummy",
			requestBody: `{"role":"invalid"}`,
			mockSetup: func() {
				mockService.EXPECT().ServiceDummyLogin(gomock.Any(), gomock.Any()).Return(
					nil,
					errors.New("ошибка авторизации"),
				)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"message": "Неверный запрос"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodPost, "/dummy-login", bytes.NewBufferString(tt.requestBody))
			req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "req-id"))
			w := httptest.NewRecorder()

			handler.HandlerDummyLogin(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var body map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&body)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, body)
		})
	}
}
