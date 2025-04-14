package mocks_test

import (
	"context"
	gomock "go.uber.org/mock/gomock"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"pvz_service/internal/app/reception"
	"pvz_service/internal/app/reception/mocks"
)

func TestCreateReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepoInterface(ctrl)

	pvzID := uuid.New()
	receptionID := uuid.New()
	receptionTime := time.Now()

	req := &reception.Reception{
		ID:       receptionID,
		DateTime: receptionTime,
		PvzID:    pvzID,
		Status:   reception.StatusInProgress,
	}

	mockRepo.EXPECT().
		CreateReception(gomock.Any(), req).
		Return(req, nil)

	ctx := context.Background()
	resp, err := mockRepo.CreateReception(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, req, resp)
}

func TestCheckOpenReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepoInterface(ctrl)

	pvzID := uuid.New()
	openReceptionID := uuid.New()

	mockRepo.EXPECT().
		CheckOpenReception(gomock.Any(), pvzID).
		Return(true, openReceptionID, nil)

	ctx := context.Background()
	open, id, err := mockRepo.CheckOpenReception(ctx, pvzID)

	assert.NoError(t, err)
	assert.True(t, open)
	assert.Equal(t, openReceptionID, id)
}

func TestCheckValidPVZID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepoInterface(ctrl)

	pvzID := uuid.New()

	mockRepo.EXPECT().
		CheckValidPVZID(gomock.Any(), pvzID).
		Return(true, nil)

	ctx := context.Background()
	valid, err := mockRepo.CheckValidPVZID(ctx, pvzID)

	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestCloseReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepoInterface(ctrl)

	receptionID := uuid.New()
	expected := &reception.Reception{
		ID:       receptionID,
		PvzID:    uuid.New(),
		Status:   reception.StatusClosed,
		DateTime: time.Now(),
	}

	mockRepo.EXPECT().
		CloseReception(gomock.Any(), receptionID).
		Return(expected, nil)

	ctx := context.Background()
	resp, err := mockRepo.CloseReception(ctx, receptionID)

	assert.NoError(t, err)
	assert.Equal(t, expected, resp)
}
