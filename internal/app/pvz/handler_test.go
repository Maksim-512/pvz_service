package pvz_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go.uber.org/mock/gomock"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"pvz_service/internal/app/pvz"
	"pvz_service/internal/app/pvz/mocks"
	"pvz_service/pkg/response"
)

func TestHandlerAddPVZ_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceInterface(ctrl)

	handler := pvz.NewPVZHandler(slog.Default(), mockService)

	pvzRequest := &pvz.CreatePVZRequest{City: "Москва"}
	pvzResponse := &pvz.PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now(),
		City:             "Москва",
	}

	mockService.EXPECT().ServiceAddPVZ(gomock.Any(), pvzRequest).Return(pvzResponse, nil)

	reqBody, _ := json.Marshal(pvzRequest)
	req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewReader(reqBody))
	req = req.WithContext(context.WithValue(req.Context(), "role", "moderator"))

	rr := httptest.NewRecorder()

	handler.HandlerAddPVZ(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var responseBody pvz.PVZ
	err := json.NewDecoder(rr.Body).Decode(&responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "Москва", responseBody.City)
}

func TestHandlerAddPVZ_ErrorCreated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceInterface(ctrl)

	handler := pvz.NewPVZHandler(slog.Default(), mockService)

	pvzRequest := &pvz.CreatePVZRequest{City: "Москва"}

	mockService.EXPECT().ServiceAddPVZ(gomock.Any(), pvzRequest).Return(nil, errors.New("ошибка"))

	reqBody, _ := json.Marshal(pvzRequest)
	req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewReader(reqBody))
	req = req.WithContext(context.WithValue(req.Context(), "role", "moderator"))

	rr := httptest.NewRecorder()

	handler.HandlerAddPVZ(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var responseBody response.ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "Неверный запрос", responseBody.Errors)
}

func TestHandlerAddPVZ_InvalidJSON(t *testing.T) {
	handler := pvz.NewPVZHandler(slog.Default(), nil)

	req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewReader([]byte("невалидный json")))
	req = req.WithContext(context.WithValue(req.Context(), "role", "moderator"))

	rr := httptest.NewRecorder()

	handler.HandlerAddPVZ(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var responseBody response.ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "Неверный запрос", responseBody.Errors)
}

func TestHandlerAddPVZ_InvalidRole(t *testing.T) {
	handler := pvz.NewPVZHandler(slog.Default(), nil)

	pvzRequest := &pvz.CreatePVZRequest{City: "Москва"}

	reqBody, _ := json.Marshal(pvzRequest)

	req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewReader(reqBody))

	rr := httptest.NewRecorder()

	handler.HandlerAddPVZ(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var responseBody response.ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "Неверный запрос", responseBody.Errors)
}

func TestHandlerGetPVZ_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceInterface(ctrl)

	handler := pvz.NewPVZHandler(slog.Default(), mockService)

	expectedResult := []*pvz.PVZWithReceptions{
		{
			PVZ: &pvz.PVZ{
				ID:   uuid.New(),
				City: "Москва",
			},
			Receptions: nil,
		},
	}
	mockService.EXPECT().ServiceGetPVZs(
		gomock.Any(),
		"",
		"",
		"1",
		"10",
	).Return(expectedResult, nil)

	req := httptest.NewRequest(http.MethodGet, "/pvz?page=1&limit=10", nil)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.HandlerGetPVZ(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var responseBody []*pvz.PVZWithReceptions
	err := json.NewDecoder(rr.Body).Decode(&responseBody)
	assert.NoError(t, err)
	assert.Len(t, responseBody, 1)
	assert.Equal(t, "Москва", responseBody[0].PVZ.City)
}

func TestHandlerGetPVZ_Fail_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceInterface(ctrl)

	handler := pvz.NewPVZHandler(slog.Default(), mockService)

	mockService.EXPECT().ServiceGetPVZs(
		gomock.Any(),
		"",
		"",
		"1",
		"10",
	).Return(
		nil,
		errors.New("ошибка получения списка"),
	)

	req := httptest.NewRequest(http.MethodGet, "/pvz?page=1&limit=10", nil)

	rr := httptest.NewRecorder()

	handler.HandlerGetPVZ(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var responseBody response.ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&responseBody)
	assert.NoError(t, err)

	assert.Equal(t, "Неверный запрос", responseBody.Errors)
}
