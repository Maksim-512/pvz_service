package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"pvz_service/internal/lib/logger/sl"
	"pvz_service/pkg/response"
)

type Handler struct {
	myLogger *slog.Logger
	service  ServiceInterface
}

func NewAuthHandler(myLogger *slog.Logger, service ServiceInterface) *Handler {
	return &Handler{
		myLogger: myLogger,
		service:  service,
	}
}

func (h *Handler) HandlerRegister(w http.ResponseWriter, r *http.Request) {
	const op = "app.auth.HandlerRegister"

	myLogger := h.myLogger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var registerRequest RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&registerRequest)
	if err != nil {
		myLogger.Error("Ошибка декодирования", sl.Err(err))
		response.MyResponseError(w, http.StatusBadRequest, "Неверный запрос")
		return
	}

	user, err := h.service.ServiceCreateUser(r.Context(), &registerRequest)
	if err != nil {
		myLogger.Error("ошибка создания пользователя", sl.Err(err))
		response.MyResponseError(w, http.StatusBadRequest, "Неверный запрос")
		return
	}

	myLogger.Info("Пользователь создан", "email", user.Email)

	response.MyResponseJSON(w, http.StatusCreated, &RegisterResponse{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	})
}

func (h *Handler) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	const op = "app.auth.HandlerLogin"

	myLogger := h.myLogger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var loginRequest LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		myLogger.Error("Ошибка декодирования", sl.Err(err))
		response.MyResponseError(w, http.StatusUnauthorized, "Неверные учетные данные")
		return
	}

	token, err := h.service.ServiceLoginUser(r.Context(), &loginRequest)
	if err != nil {
		myLogger.Error("Ошибка авторизация", sl.Err(err))
		response.MyResponseError(w, http.StatusUnauthorized, "Неверные учетные данные")
		return
	}

	myLogger.Info("Успешная авторизация")

	response.MyResponseJSON(w, http.StatusOK, &LoginResponse{
		Token: token.Token,
	})
}

func (h *Handler) HandlerDummyLogin(w http.ResponseWriter, r *http.Request) {
	const op = "app.auth.HandlerDummyLogin"

	myLogger := h.myLogger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var dummyLoginRequest DummyLoginRequest

	err := json.NewDecoder(r.Body).Decode(&dummyLoginRequest)
	if err != nil {
		myLogger.Error("Ошибка декодирования", sl.Err(err))
		response.MyResponseError(w, http.StatusBadRequest, "Неверный запрос")
		return
	}

	token, err := h.service.ServiceDummyLogin(r.Context(), &dummyLoginRequest)
	if err != nil {
		myLogger.Error("Неверный запрос", sl.Err(err))
		response.MyResponseError(w, http.StatusBadRequest, "Неверный запрос")
		return
	}

	myLogger.Info("Успешная авторизация через dummyLogin")

	response.MyResponseJSON(w, http.StatusOK, &LoginResponse{
		Token: token.Token,
	})

}
