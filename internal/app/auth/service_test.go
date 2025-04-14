package auth_test

import (
	"context"
	"errors"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"pvz_service/internal/app/auth"
	"pvz_service/internal/app/auth/mocks"
	"pvz_service/pkg/jwt"
	mockJWT "pvz_service/pkg/jwt/mocks"
)

func newServiceWithMocks(t *testing.T) (*auth.Service, *mocks.MockUserRepository, *mockJWT.MockTokenService) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)
	logger := slog.Default()
	mockToken := mockJWT.NewMockTokenService(ctrl)
	service := auth.NewAuthService(logger, mockRepo, mockToken)
	return service, mockRepo, mockToken
}

func TestServiceCreateUser_Success(t *testing.T) {
	service, mockRepo, _ := newServiceWithMocks(t)
	ctx := context.Background()

	req := &auth.RegisterRequest{
		Email:    "test@yandex.ru",
		Password: "password",
		Role:     auth.RoleEmployee,
	}

	mockRepo.EXPECT().CheckUser(ctx, req).Return(nil)
	mockRepo.EXPECT().CreateUser(ctx, gomock.Any()).Return(nil)

	user, err := service.ServiceCreateUser(ctx, req)

	require.NoError(t, err)
	require.Equal(t, req.Email, user.Email)
	require.Equal(t, req.Role, user.Role)
}

func TestServiceCreateUser_InvalidEmail(t *testing.T) {
	service, _, _ := newServiceWithMocks(t)
	ctx := context.Background()

	req := &auth.RegisterRequest{
		Email:    "invalid@yandexru",
		Password: "password",
		Role:     auth.RoleEmployee,
	}

	user, err := service.ServiceCreateUser(ctx, req)

	require.Error(t, err)
	require.Nil(t, user)
	require.Equal(t, "невалидный email", err.Error())
}

func TestServiceCreateUser_UserAlreadyExists(t *testing.T) {
	service, mockRepo, _ := newServiceWithMocks(t)
	ctx := context.Background()

	req := &auth.RegisterRequest{
		Email:    "test@yandex.ru",
		Password: "password",
		Role:     auth.RoleEmployee,
	}

	mockRepo.EXPECT().CheckUser(ctx, req).Return(errors.New("пользователь уже существует"))

	user, err := service.ServiceCreateUser(ctx, req)

	require.Error(t, err)
	require.Nil(t, user)
}

func TestServiceCreateUser_HashError(t *testing.T) {
	service, mockRepo, _ := newServiceWithMocks(t)
	ctx := context.Background()

	req := &auth.RegisterRequest{
		Email:    "test@yandex.ru",
		Password: "password",
		Role:     auth.RoleEmployee,
	}

	mockRepo.EXPECT().CheckUser(ctx, req).Return(nil)

	service.HashPassword = func(_ string) (string, error) {
		return "", errors.New("ошибка хеширования")
	}

	user, err := service.ServiceCreateUser(ctx, req)

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "ошибка хеширования")

}

func TestServiceCreateUser_CreateUserError(t *testing.T) {
	service, mockRepo, _ := newServiceWithMocks(t)
	ctx := context.Background()

	req := &auth.RegisterRequest{
		Email:    "test@yandex.ru",
		Password: "password",
		Role:     auth.RoleEmployee,
	}

	mockRepo.EXPECT().CheckUser(ctx, req).Return(nil)
	mockRepo.EXPECT().CreateUser(ctx, gomock.Any()).Return(errors.New("ошибка создания"))

	user, err := service.ServiceCreateUser(ctx, req)

	require.Error(t, err)
	require.Nil(t, user)
}

func TestServiceCreateUser_InvalidRole(t *testing.T) {
	service, _, _ := newServiceWithMocks(t)
	ctx := context.Background()

	req := &auth.RegisterRequest{
		Email:    "test@yandex.ru",
		Password: "password",
		Role:     "admin",
	}

	user, err := service.ServiceCreateUser(ctx, req)

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "некорректная роль")
}

func TestServiceCreateUser_MissingRequiredFields(t *testing.T) {
	service, _, _ := newServiceWithMocks(t)
	ctx := context.Background()

	req := &auth.RegisterRequest{
		Email:    "",
		Password: "",
		Role:     "",
	}

	user, err := service.ServiceCreateUser(ctx, req)

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "не все обязательные поля заполнены")
}

func TestServiceLoginUser_Success(t *testing.T) {
	service, mockRepo, mockToken := newServiceWithMocks(t)
	ctx := context.Background()

	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	mockUser := &auth.User{
		Email:        "test@yandex.ru",
		PasswordHash: string(hash),
		Role:         auth.RoleEmployee,
	}

	mockRepo.EXPECT().GetUserByEmail(ctx, "test@yandex.ru").Return(mockUser, nil)

	mockToken.EXPECT().GenerateToken(auth.RoleEmployee).Return(&jwt.TokenResponse{
		Token: "valid_token",
	}, nil)

	req := &auth.LoginRequest{
		Email:    "test@yandex.ru",
		Password: "password",
	}

	resp, err := service.ServiceLoginUser(ctx, req)

	require.NoError(t, err)
	require.NotEmpty(t, resp.Token)
}

func TestServiceLoginUser_InvalidEmail(t *testing.T) {
	service, mockRepo, _ := newServiceWithMocks(t)
	ctx := context.Background()

	mockRepo.EXPECT().GetUserByEmail(ctx, "invalid@yandex.ru").Return(
		nil,
		errors.New("ошибка получения пользователя"),
	)

	req := &auth.LoginRequest{
		Email:    "invalid@yandex.ru",
		Password: "password",
	}

	resp, err := service.ServiceLoginUser(ctx, req)

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestServiceLoginUser_InvalidPassword(t *testing.T) {
	service, mockRepo, _ := newServiceWithMocks(t)
	ctx := context.Background()

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)

	mockUser := &auth.User{
		Email:        "test@yandex.ru",
		PasswordHash: string(hash),
		Role:         auth.RoleEmployee,
	}

	mockRepo.EXPECT().GetUserByEmail(ctx, "test@yandex.ru").Return(mockUser, nil)

	req := &auth.LoginRequest{
		Email:    "test@yandex.ru",
		Password: "invalidpassword",
	}

	resp, err := service.ServiceLoginUser(ctx, req)

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestServiceLoginUser_TokenGenerationError(t *testing.T) {
	service, mockRepo, mockToken := newServiceWithMocks(t)

	req := &auth.LoginRequest{
		Email:    "tests@yandex.ru",
		Password: "password",
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	require.NoError(t, err)

	mockRepo.EXPECT().GetUserByEmail(gomock.Any(), req.Email).
		Return(&auth.User{
			Email:        req.Email,
			PasswordHash: string(hashedPassword),
			Role:         "employee",
		}, nil)

	mockToken.EXPECT().GenerateToken(gomock.Any()).
		Return(nil, errors.New("ошибка генерации токена"))

	token, err := service.ServiceLoginUser(context.Background(), req)

	require.Error(t, err)
	require.Nil(t, token)
	require.Equal(t, "ошибка генерации токена", err.Error())
}

func TestServiceDummyLogin_SuccessEmployee(t *testing.T) {
	service, _, mockToken := newServiceWithMocks(t)
	ctx := context.Background()

	req := &auth.DummyLoginRequest{
		Role: auth.RoleEmployee,
	}

	mockToken.EXPECT().GenerateToken(auth.RoleEmployee).Return(
		&jwt.TokenResponse{
			Token: "valid_token",
		},
		nil,
	)

	resp, err := service.ServiceDummyLogin(ctx, req)

	require.NoError(t, err)
	require.NotEmpty(t, resp.Token)
}

func TestServiceDummyLogin_SuccessModerator(t *testing.T) {
	service, _, mockToken := newServiceWithMocks(t)
	ctx := context.Background()

	req := &auth.DummyLoginRequest{
		Role: auth.RoleModerator,
	}

	mockToken.EXPECT().GenerateToken(auth.RoleModerator).Return(
		&jwt.TokenResponse{
			Token: "valid_token",
		},
		nil,
	)

	resp, err := service.ServiceDummyLogin(ctx, req)

	require.NoError(t, err)
	require.NotEmpty(t, resp.Token)
}

func TestServiceDummyLogin_InvalidRole(t *testing.T) {
	service, _, _ := newServiceWithMocks(t)
	ctx := context.Background()

	req := &auth.DummyLoginRequest{
		Role: "admin",
	}

	resp, err := service.ServiceDummyLogin(ctx, req)

	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, "неверная роль", err.Error())
}

func TestServiceDummyLogin_TokenGenerationError(t *testing.T) {
	service, _, mockToken := newServiceWithMocks(t)
	ctx := context.Background()

	dummyRequest := &auth.DummyLoginRequest{
		Role: auth.RoleModerator,
	}

	mockToken.EXPECT().GenerateToken(dummyRequest.Role).
		Return(nil, errors.New("ошибка генерации токена"))

	tokenResp, err := service.ServiceDummyLogin(ctx, dummyRequest)

	require.Error(t, err)
	require.Nil(t, tokenResp)
	require.Equal(t, "ошибка генерации токена", err.Error())
}
