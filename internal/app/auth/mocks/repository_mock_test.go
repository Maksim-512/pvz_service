package mocks_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"

	"pvz_service/internal/app/auth"
	"pvz_service/internal/app/auth/mocks"
)

func TestMockUserRepository_CheckUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	registerRequest := &auth.RegisterRequest{
		Email:    "tests@yandex.ru",
		Password: "password",
	}

	mockRepo.EXPECT().CheckUser(context.Background(), registerRequest).Return(nil)

	err := mockRepo.CheckUser(context.Background(), registerRequest)
	assert.NoError(t, err)
}

func TestMockUserRepository_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	user := &auth.User{
		Email:        "tests@yandex.ru",
		PasswordHash: "password",
	}

	mockRepo.EXPECT().CreateUser(context.Background(), user).Return(nil)

	err := mockRepo.CreateUser(context.Background(), user)
	assert.NoError(t, err)
}

func TestMockUserRepository_GetUserByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	email := "tests@yandex.ru"
	expectedUser := &auth.User{
		Email:        email,
		PasswordHash: "password",
	}

	mockRepo.EXPECT().GetUserByEmail(context.Background(), email).Return(expectedUser, nil)

	user, err := mockRepo.GetUserByEmail(context.Background(), email)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestMockUserRepository_GetUserByEmail_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	email := "invalid@yandex.ru"

	mockRepo.EXPECT().GetUserByEmail(context.Background(), email).Return(
		nil,
		errors.New("нет такого пользователя"),
	)

	user, err := mockRepo.GetUserByEmail(context.Background(), email)
	assert.Error(t, err)
	assert.Nil(t, user)
}
