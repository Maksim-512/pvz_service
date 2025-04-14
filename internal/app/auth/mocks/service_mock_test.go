package mocks_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"

	"pvz_service/internal/app/auth"
	"pvz_service/internal/app/auth/mocks"
	"pvz_service/pkg/jwt"
)

func TestMockServiceInterface_ServiceCreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceInterface(ctrl)
	registerRequest := &auth.RegisterRequest{
		Email:    "tests@yandex.ru",
		Password: "password",
	}

	mockService.EXPECT().ServiceCreateUser(context.Background(), registerRequest).Return(
		&auth.User{
			Email: "tests@yandex.ru",
		},
		nil,
	)

	user, err := mockService.ServiceCreateUser(context.Background(), registerRequest)
	assert.NoError(t, err)
	assert.Equal(t, "tests@yandex.ru", user.Email)
}

func TestMockServiceInterface_ServiceDummyLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceInterface(ctrl)
	dummyRequest := &auth.DummyLoginRequest{Role: "employee"}

	mockService.EXPECT().ServiceDummyLogin(context.Background(), dummyRequest).Return(
		&jwt.TokenResponse{
			Token: "dummy_token",
		},
		nil,
	)

	tokenResponse, err := mockService.ServiceDummyLogin(context.Background(), dummyRequest)
	assert.NoError(t, err)
	assert.Equal(t, "dummy_token", tokenResponse.Token)
}

func TestMockServiceInterface_ServiceLoginUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceInterface(ctrl)
	loginRequest := &auth.LoginRequest{
		Email:    "tests@yandex.ru",
		Password: "password",
	}

	mockService.EXPECT().ServiceLoginUser(context.Background(), loginRequest).Return(
		&jwt.TokenResponse{
			Token: "login_token",
		},
		nil,
	)

	tokenResponse, err := mockService.ServiceLoginUser(context.Background(), loginRequest)
	assert.NoError(t, err)
	assert.Equal(t, "login_token", tokenResponse.Token)
}

func TestMockServiceInterface_ServiceLoginUser_InvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceInterface(ctrl)
	loginRequest := &auth.LoginRequest{Email: "invalid@yandex.ru", Password: "invalidpassword"}

	mockService.EXPECT().ServiceLoginUser(context.Background(), loginRequest).Return(
		nil,
		errors.New("ошибка прав"),
	)

	tokenResponse, err := mockService.ServiceLoginUser(context.Background(), loginRequest)
	assert.Error(t, err)
	assert.Nil(t, tokenResponse)
}
