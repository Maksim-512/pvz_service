package reception_test

import (
	"context"
	"errors"
	"go.uber.org/mock/gomock"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"pvz_service/internal/app/reception"
	"pvz_service/internal/app/reception/mocks"
)

func newServiceWithMocks(t *testing.T) (*reception.Service, *mocks.MockRepoInterface) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	receptionRepoMock := mocks.NewMockRepoInterface(ctrl)
	logger := slog.Default()

	service := reception.NewReceptionService(logger, receptionRepoMock)

	return service, receptionRepoMock
}

func TestServiceOpenReception_Success(t *testing.T) {
	service, receptionRepoMock := newServiceWithMocks(t)

	pvzID := uuid.New()
	receptionRequest := &reception.OpenReceptionRequest{PvzID: pvzID}

	receptionRepoMock.EXPECT().
		CheckValidPVZID(gomock.Any(), pvzID).
		Return(true, nil)

	receptionRepoMock.EXPECT().
		CheckOpenReception(gomock.Any(), pvzID).
		Return(false, uuid.Nil, nil)

	receptionRepoMock.EXPECT().
		CreateReception(gomock.Any(), gomock.Any()).
		Return(&reception.Reception{ID: uuid.New(), PvzID: pvzID, Status: reception.StatusInProgress}, nil)

	myReception, err := service.ServiceOpenReception(context.Background(), receptionRequest)

	assert.NoError(t, err)
	assert.NotNil(t, myReception)
	assert.Equal(t, reception.StatusInProgress, myReception.Status)
}

func TestServiceOpenReception_InvalidPVZ(t *testing.T) {
	service, receptionRepoMock := newServiceWithMocks(t)

	pvzID := uuid.New()
	receptionRequest := &reception.OpenReceptionRequest{PvzID: pvzID}

	receptionRepoMock.EXPECT().
		CheckValidPVZID(gomock.Any(), pvzID).
		Return(false, nil).
		Times(1)

	myReception, err := service.ServiceOpenReception(context.Background(), receptionRequest)

	assert.Error(t, err)
	assert.Nil(t, myReception)
}

func TestServiceOpenReception_ExistingOpenReception(t *testing.T) {
	service, receptionRepoMock := newServiceWithMocks(t)

	pvzID := uuid.New()
	receptionRequest := &reception.OpenReceptionRequest{PvzID: pvzID}

	receptionRepoMock.EXPECT().
		CheckValidPVZID(gomock.Any(), pvzID).
		Return(true, nil)

	receptionRepoMock.EXPECT().
		CheckOpenReception(gomock.Any(), pvzID).
		Return(true, uuid.New(), nil)

	myReception, err := service.ServiceOpenReception(context.Background(), receptionRequest)

	assert.Error(t, err)
	assert.Nil(t, myReception)
}

func TestServiceCloseReception_Success(t *testing.T) {
	service, receptionRepoMock := newServiceWithMocks(t)

	pvzID := uuid.New()
	receptionRequest := &reception.CloseReceptionRequest{PvzID: pvzID}
	receptionID := uuid.New()

	receptionRepoMock.EXPECT().
		CheckValidPVZID(gomock.Any(), pvzID).
		Return(true, nil)

	receptionRepoMock.EXPECT().
		CheckOpenReception(gomock.Any(), pvzID).
		Return(true, receptionID, nil)

	receptionRepoMock.EXPECT().
		CloseReception(gomock.Any(), receptionID).
		Return(&reception.Reception{ID: receptionID, PvzID: pvzID, Status: reception.StatusClosed}, nil)

	myReception, err := service.ServiceCloseReception(context.Background(), receptionRequest)

	assert.NoError(t, err)
	assert.NotNil(t, myReception)
	assert.Equal(t, reception.StatusClosed, myReception.Status)
}

func TestServiceCloseReception_InvalidPVZ(t *testing.T) {
	service, receptionRepoMock := newServiceWithMocks(t)

	pvzID := uuid.New()
	receptionRequest := &reception.CloseReceptionRequest{PvzID: pvzID}

	receptionRepoMock.EXPECT().
		CheckValidPVZID(gomock.Any(), pvzID).
		Return(false, nil)

	myReception, err := service.ServiceCloseReception(context.Background(), receptionRequest)

	assert.Error(t, err)
	assert.Nil(t, myReception)
}

func TestServiceCloseReception_NoOpenReception(t *testing.T) {
	service, receptionRepoMock := newServiceWithMocks(t)

	pvzID := uuid.New()
	receptionRequest := &reception.CloseReceptionRequest{PvzID: pvzID}

	receptionRepoMock.EXPECT().
		CheckValidPVZID(gomock.Any(), pvzID).
		Return(true, nil)

	receptionRepoMock.EXPECT().
		CheckOpenReception(gomock.Any(), pvzID).
		Return(false, uuid.Nil, nil)

	myReception, err := service.ServiceCloseReception(context.Background(), receptionRequest)

	assert.Error(t, err)
	assert.Nil(t, myReception)
}

func TestServiceOpenReception_Error_CheckValidPVZID(t *testing.T) {
	service, receptionRepoMock := newServiceWithMocks(t)

	pvzID := uuid.New()
	receptionRequest := &reception.OpenReceptionRequest{PvzID: pvzID}

	expectedErr := errors.New("ошибка БД")

	receptionRepoMock.EXPECT().
		CheckValidPVZID(gomock.Any(), pvzID).
		Return(false, expectedErr).
		Times(1)

	myReception, err := service.ServiceOpenReception(context.Background(), receptionRequest)

	assert.Error(t, err)
	assert.Nil(t, myReception)
	assert.Contains(t, err.Error(), "ошибка при проверке на наличие пункта PVZ")
}

func TestServiceOpenReception_Error_CheckOpenReception(t *testing.T) {
	service, receptionRepoMock := newServiceWithMocks(t)

	pvzID := uuid.New()
	receptionRequest := &reception.OpenReceptionRequest{PvzID: pvzID}

	receptionRepoMock.EXPECT().
		CheckValidPVZID(gomock.Any(), pvzID).
		Return(true, nil).
		Times(1)

	expectedErr := errors.New("ошибка БД")

	receptionRepoMock.EXPECT().
		CheckOpenReception(gomock.Any(), pvzID).
		Return(false, uuid.Nil, expectedErr).
		Times(1)

	myReception, err := service.ServiceOpenReception(context.Background(), receptionRequest)

	assert.Error(t, err)
	assert.Nil(t, myReception)
	assert.Contains(t, err.Error(), "ошибка при проверке на наличие открытой приемки")
}

func TestServiceOpenReception_Error_CreateReception(t *testing.T) {
	service, receptionRepoMock := newServiceWithMocks(t)

	pvzID := uuid.New()
	receptionRequest := &reception.OpenReceptionRequest{PvzID: pvzID}

	receptionRepoMock.EXPECT().
		CheckValidPVZID(gomock.Any(), pvzID).
		Return(true, nil).
		Times(1)

	receptionRepoMock.EXPECT().
		CheckOpenReception(gomock.Any(), pvzID).
		Return(false, uuid.Nil, nil).
		Times(1)

	expectedErr := errors.New("ошибка БД")

	receptionRepoMock.EXPECT().
		CreateReception(gomock.Any(), gomock.Any()).
		Return(nil, expectedErr).
		Times(1)

	myReception, err := service.ServiceOpenReception(context.Background(), receptionRequest)

	assert.Error(t, err)
	assert.Nil(t, myReception)
	assert.Contains(t, err.Error(), "oшибка создания приемки")
}

func TestServiceCloseReception_Error_CheckValidPVZID(t *testing.T) {
	service, receptionRepoMock := newServiceWithMocks(t)

	pvzID := uuid.New()
	receptionRequest := &reception.CloseReceptionRequest{PvzID: pvzID}

	expectedErr := errors.New("ошибка БД")

	receptionRepoMock.EXPECT().
		CheckValidPVZID(gomock.Any(), pvzID).
		Return(false, expectedErr).
		Times(1)

	myReception, err := service.ServiceCloseReception(context.Background(), receptionRequest)

	assert.Error(t, err)
	assert.Nil(t, myReception)
	assert.Contains(t, err.Error(), "oшибка при проверке на наличие пункта PVZ")
}

func TestServiceCloseReception_Error_CheckOpenReception(t *testing.T) {
	service, receptionRepoMock := newServiceWithMocks(t)

	pvzID := uuid.New()
	receptionRequest := &reception.CloseReceptionRequest{PvzID: pvzID}

	receptionRepoMock.EXPECT().
		CheckValidPVZID(gomock.Any(), pvzID).
		Return(true, nil).
		Times(1)

	expectedErr := errors.New("ошибка БД")

	receptionRepoMock.EXPECT().
		CheckOpenReception(gomock.Any(), pvzID).
		Return(false, uuid.Nil, expectedErr).
		Times(1)

	myReception, err := service.ServiceCloseReception(context.Background(), receptionRequest)

	assert.Error(t, err)
	assert.Nil(t, myReception)
	assert.Contains(t, err.Error(), "ошибка при проверке на наличие открытой приемки")
}

func TestServiceCloseReception_Error_CloseReception(t *testing.T) {
	service, receptionRepoMock := newServiceWithMocks(t)

	pvzID := uuid.New()
	receptionRequest := &reception.CloseReceptionRequest{PvzID: pvzID}
	receptionID := uuid.New()

	receptionRepoMock.EXPECT().
		CheckValidPVZID(gomock.Any(), pvzID).
		Return(true, nil).
		Times(1)

	receptionRepoMock.EXPECT().
		CheckOpenReception(gomock.Any(), pvzID).
		Return(true, receptionID, nil).
		Times(1)

	expectedErr := errors.New("ошибка БД")

	receptionRepoMock.EXPECT().
		CloseReception(gomock.Any(), receptionID).
		Return(nil, expectedErr).
		Times(1)

	myReception, err := service.ServiceCloseReception(context.Background(), receptionRequest)

	assert.Error(t, err)
	assert.Nil(t, myReception)
	assert.Contains(t, err.Error(), "oшибка закрытия приемки")
}

func TestServiceCloseReception_Error_InvalidPvzID(t *testing.T) {
	service, receptionRepoMock := newServiceWithMocks(t)

	pvzID := uuid.New()
	receptionRequest := &reception.CloseReceptionRequest{PvzID: pvzID}

	receptionRepoMock.EXPECT().
		CheckValidPVZID(gomock.Any(), pvzID).
		Return(false, nil).
		Times(1)

	result, err := service.ServiceCloseReception(context.Background(), receptionRequest)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "неверный запрос или приемка уже закрыта", err.Error())
}

func TestMockSetup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepoInterface(ctrl)
	receptionID := uuid.New()

	mockRepo.EXPECT().
		CloseReception(gomock.Any(), receptionID).
		Return(&reception.Reception{ID: receptionID}, nil)

	result, err := mockRepo.CloseReception(context.Background(), receptionID)
	assert.NoError(t, err)
	assert.Equal(t, receptionID, result.ID)
}
