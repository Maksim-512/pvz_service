package auth

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"pvz_service/internal/lib/logger/sl"
	"pvz_service/pkg/jwt"
)

// mockgen -source=internal/app/auth/service.go -destination=internal/app/auth/mocks/service_mock.go -package=mocks

type ServiceInterface interface {
	ServiceCreateUser(ctx context.Context, registerRequest *RegisterRequest) (*User, error)
	ServiceLoginUser(ctx context.Context, loginRequest *LoginRequest) (*jwt.TokenResponse, error)
	ServiceDummyLogin(ctx context.Context, dummyRequest *DummyLoginRequest) (*jwt.TokenResponse, error)
}

type Service struct {
	userRepo      UserRepository
	myLogger      *slog.Logger
	jwtService    jwt.TokenService
	HashPassword  func(password string) (string, error)
	CheckPassword func(hashed, plain string) error
}

func NewAuthService(myLogger *slog.Logger, userRepo UserRepository, jwtService jwt.TokenService) *Service {
	return &Service{
		userRepo:   userRepo,
		myLogger:   myLogger,
		jwtService: jwtService,
		HashPassword: func(password string) (string, error) {
			bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
			return string(bytes), err
		},
		CheckPassword: func(hashed, plain string) error {
			return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
		},
	}
}

func (a *Service) ServiceCreateUser(ctx context.Context, registerRequest *RegisterRequest) (*User, error) {
	const op = "app.auth.ServiceCreateUser"

	myLogger := a.myLogger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(ctx)),
	)

	err := validateRegisterRequest(registerRequest)
	if err != nil {
		myLogger.Error(
			"Неверный запрос",
			sl.Err(err),
		)
		return nil, err
	}

	err = a.userRepo.CheckUser(ctx, registerRequest)
	if err != nil {
		myLogger.Error("Пользователь не может быть добавлен", sl.Err(err))
		return nil, err
	}

	passHash, err := a.HashPassword(registerRequest.Password)
	if err != nil {
		myLogger.Error("ошибка хеширования пароля", sl.Err(err))
		return nil, err
	}

	userID := uuid.New()

	user := &User{
		ID:           userID,
		Email:        registerRequest.Email,
		PasswordHash: string(passHash),
		Role:         registerRequest.Role,
	}

	err = a.userRepo.CreateUser(ctx, user)
	if err != nil {
		myLogger.Error("Ошибка создания пользователя", sl.Err(err))
		return nil, err
	}

	return user, nil
}

func (a *Service) ServiceLoginUser(ctx context.Context, loginRequest *LoginRequest) (*jwt.TokenResponse, error) {
	const op = "app.auth.ServiceLoginUser"

	myLogger := a.myLogger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(ctx)),
	)

	user, err := a.userRepo.GetUserByEmail(ctx, loginRequest.Email)
	if err != nil {
		myLogger.Error("Ошибка получения пользователя, неверный email", sl.Err(err))
		return nil, err
	}

	err = a.CheckPassword(user.PasswordHash, loginRequest.Password)
	if err != nil {
		myLogger.Error("Ошибка получения пользователя, неверный пароль", sl.Err(err))
		return nil, err
	}

	token, err := a.jwtService.GenerateToken(user.Role)
	if err != nil {
		myLogger.Error("Ошибка получения токена авторизации", sl.Err(err))
		return nil, err
	}

	return token, nil
}

func (a *Service) ServiceDummyLogin(ctx context.Context, dummyRequest *DummyLoginRequest) (*jwt.TokenResponse, error) {
	op := "app.auth.ServiceDummyLogin"

	myLogger := a.myLogger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(ctx)),
	)

	if dummyRequest.Role != RoleEmployee && dummyRequest.Role != RoleModerator {
		myLogger.Error("Неверная роль")
		return nil, errors.New("неверная роль")
	}

	token, err := a.jwtService.GenerateToken(dummyRequest.Role)
	if err != nil {
		myLogger.Error("Ошибка получения токена авторизации", sl.Err(err))
		return nil, err
	}

	return token, nil
}

func validateRegisterRequest(r *RegisterRequest) error {
	if r.Email == "" || r.Password == "" || r.Role == "" {
		return errors.New("не все обязательные поля заполнены")
	}
	if !isValidEmail(r.Email) {
		return errors.New("невалидный email")
	}
	if r.Role != RoleEmployee && r.Role != RoleModerator {
		return errors.New("некорректная роль")
	}
	return nil
}

func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
