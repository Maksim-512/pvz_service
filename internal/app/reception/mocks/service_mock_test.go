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

func TestServiceOpenReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceInterface(ctrl)

	pvzID := uuid.New()
	req := &reception.OpenReceptionRequest{PvzID: pvzID}

	expected := &reception.Reception{
		ID:       uuid.New(),
		PvzID:    pvzID,
		Status:   reception.StatusInProgress,
		DateTime: time.Now(),
	}

	mockService.EXPECT().
		ServiceOpenReception(gomock.Any(), req).
		Return(expected, nil)

	ctx := context.Background()
	resp, err := mockService.ServiceOpenReception(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, expected, resp)
}

func TestServiceCloseReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceInterface(ctrl)

	pvzID := uuid.New()
	req := &reception.CloseReceptionRequest{PvzID: pvzID}

	expected := &reception.Reception{
		ID:       uuid.New(),
		PvzID:    pvzID,
		Status:   reception.StatusClosed,
		DateTime: time.Now(),
	}

	mockService.EXPECT().
		ServiceCloseReception(gomock.Any(), req).
		Return(expected, nil)

	ctx := context.Background()
	resp, err := mockService.ServiceCloseReception(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, expected, resp)
}
