package mocks_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"

	"pvz_service/internal/app/product"
	"pvz_service/internal/app/pvz"
	"pvz_service/internal/app/pvz/mocks"
	"pvz_service/internal/app/reception"
)

func TestMockRepository_CreatePVZ_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	ctx := context.Background()

	pvzToCreate := &pvz.PVZ{
		City: "Москва",
	}
	createdPVZ := &pvz.PVZ{
		ID:               uuid.New(),
		City:             "Москва",
		RegistrationDate: time.Now(),
	}

	mockRepo.EXPECT().
		CreatePVZ(ctx, pvzToCreate).
		Return(createdPVZ, nil)

	result, err := mockRepo.CreatePVZ(ctx, pvzToCreate)

	assert.NoError(t, err)
	assert.Equal(t, createdPVZ.City, result.City)
	assert.Equal(t, createdPVZ.ID, result.ID)
}

func TestMockRepository_CreatePVZ_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	ctx := context.Background()

	pvzToCreate := &pvz.PVZ{City: "Казань"}

	mockRepo.EXPECT().
		CreatePVZ(ctx, pvzToCreate).
		Return(nil, assert.AnError)

	result, err := mockRepo.CreatePVZ(ctx, pvzToCreate)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestMockRepository_GetPVZWithReceptions_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	ctx := context.Background()

	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()
	page := 1
	limit := 10

	pvzID := uuid.New()
	expectedResult := []*pvz.PVZWithReceptions{
		{
			PVZ: &pvz.PVZ{
				ID:               pvzID,
				City:             "Санкт-Петербург",
				RegistrationDate: time.Now(),
			},
			Receptions: []*pvz.ReceptionWithProducts{
				{
					Reception: &reception.Reception{
						ID:       uuid.New(),
						PvzID:    pvzID,
						DateTime: time.Now(),
					},
					Products: []*product.Product{
						{
							ID:          uuid.New(),
							Type:        product.TypeElectronics,
							ReceptionID: pvzID,
							DateTime:    time.Now(),
						},
					},
				},
			},
		},
	}

	mockRepo.EXPECT().
		GetPVZWithReceptions(ctx, &startDate, &endDate, page, limit).
		Return(expectedResult, nil)

	result, err := mockRepo.GetPVZWithReceptions(ctx, &startDate, &endDate, 1, 10)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}
