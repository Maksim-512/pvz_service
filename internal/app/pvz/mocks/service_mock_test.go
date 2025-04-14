package mocks_test

import (
	"context"
	"errors"
	gomock "go.uber.org/mock/gomock"
	"testing"

	"github.com/stretchr/testify/assert"

	"pvz_service/internal/app/pvz"
	"pvz_service/internal/app/pvz/mocks"
)

func TestServiceGetPVZs_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceInterface(ctrl)

	ctx := context.Background()

	expected := []*pvz.PVZWithReceptions{
		{
			PVZ: &pvz.PVZ{
				City: "Москва",
			},
			Receptions: []*pvz.ReceptionWithProducts{},
		},
	}

	mockService.EXPECT().
		ServiceGetPVZs(ctx, "2024-01-01", "2024-01-31", "1", "10").
		Return(expected, nil)

	result, err := mockService.ServiceGetPVZs(
		ctx,
		"2024-01-01",
		"2024-01-31",
		"1",
		"10",
	)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestServiceGetPVZs_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceInterface(ctrl)

	ctx := context.Background()
	expectedErr := errors.New("ошибка получения данных")

	mockService.EXPECT().
		ServiceGetPVZs(ctx, "invalid", "invalid", "0", "0").
		Return(nil, expectedErr)

	result, err := mockService.ServiceGetPVZs(ctx, "invalid", "invalid", "0", "0")

	assert.Nil(t, result)
	assert.EqualError(t, err, "ошибка получения данных")
}

func TestServiceAddPVZ_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceInterface(ctrl)

	ctx := context.Background()
	req := &pvz.CreatePVZRequest{City: "Казань"}

	mockResponse := &pvz.PVZ{
		City: "Казань",
	}

	mockService.EXPECT().
		ServiceAddPVZ(ctx, req).
		Return(mockResponse, nil)

	result, err := mockService.ServiceAddPVZ(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, mockResponse, result)
}

func TestServiceAddPVZ_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceInterface(ctrl)

	ctx := context.Background()
	req := &pvz.CreatePVZRequest{City: "Мытищи"}
	mockErr := errors.New("ошибка создания ПВЗ")

	mockService.EXPECT().
		ServiceAddPVZ(ctx, req).
		Return(nil, mockErr)

	result, err := mockService.ServiceAddPVZ(ctx, req)

	assert.Nil(t, result)
	assert.EqualError(t, err, "ошибка создания ПВЗ")
}
