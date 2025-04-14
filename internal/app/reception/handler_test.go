package reception_test

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

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"pvz_service/internal/app/reception"
	"pvz_service/internal/app/reception/mocks"
)

func setup(t *testing.T) (*gomock.Controller, *mocks.MockServiceInterface, *reception.Handler) {
	ctrl := gomock.NewController(t)
	mockService := mocks.NewMockServiceInterface(ctrl)
	logger := slog.Default()
	handler := reception.NewReceptionHandler(logger, mockService)
	return ctrl, mockService, handler
}

func TestHandlerOpenReception_Success(t *testing.T) {
	ctrl, mockService, handler := setup(t)
	defer ctrl.Finish()

	pvzID := uuid.New()
	receptionID := uuid.New()
	now := time.Now()

	body := map[string]string{
		"pvzId": pvzID.String(),
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/receptions", bytes.NewReader(jsonBody))
	w := httptest.NewRecorder()

	mockService.EXPECT().ServiceOpenReception(gomock.Any(), gomock.Any()).Return(&reception.Reception{
		ID:       receptionID,
		DateTime: now,
		PvzID:    pvzID,
		Status:   "in_progress",
	}, nil)

	handler.HandlerOpenReception(w, req)
	resp := w.Result()

	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestHandlerOpenReception_InvalidJSON(t *testing.T) {
	_, _, handler := setup(t)

	req := httptest.NewRequest(http.MethodPost, "/receptions", bytes.NewReader([]byte("{invalid_json")))
	w := httptest.NewRecorder()

	handler.HandlerOpenReception(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlerOpenReception_ServiceError(t *testing.T) {
	ctrl, mockService, handler := setup(t)
	defer ctrl.Finish()

	pvzID := uuid.New()
	body := map[string]string{
		"pvzId": pvzID.String(),
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/receptions", bytes.NewReader(jsonBody))
	w := httptest.NewRecorder()

	mockService.EXPECT().ServiceOpenReception(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("ошибка"))

	handler.HandlerOpenReception(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlerCloseReception_Success(t *testing.T) {
	ctrl, mockService, handler := setup(t)
	defer ctrl.Finish()

	pvzID := uuid.New()
	receptionID := uuid.New()
	now := time.Now()

	req := httptest.NewRequest(http.MethodPost, "/pvz/"+pvzID.String()+"/close_last_reception", nil)
	w := httptest.NewRecorder()

	rCtx := chi.NewRouteContext()
	rCtx.URLParams.Add("pvzID", pvzID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rCtx))

	mockService.EXPECT().ServiceCloseReception(gomock.Any(), gomock.Any()).
		Return(&reception.Reception{
			ID:       receptionID,
			DateTime: now,
			PvzID:    pvzID,
			Status:   "close",
		}, nil)

	handler.HandlerCloseReception(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestHandlerCloseReception_InvalidUUID(t *testing.T) {
	_, _, handler := setup(t)

	req := httptest.NewRequest(http.MethodPost, "/pvz/invalid-uuid/close_last_reception", nil)
	w := httptest.NewRecorder()

	rCtx := chi.NewRouteContext()
	rCtx.URLParams.Add("pvzID", "invalid-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rCtx))

	handler.HandlerCloseReception(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlerCloseReception_ServiceError(t *testing.T) {
	ctrl, mockService, handler := setup(t)
	defer ctrl.Finish()

	pvzID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/pvz/"+pvzID.String()+"/close_last_reception", nil)
	w := httptest.NewRecorder()

	rCtx := chi.NewRouteContext()
	rCtx.URLParams.Add("pvzID", pvzID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rCtx))

	mockService.EXPECT().ServiceCloseReception(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("приемка уже закрыта"))

	handler.HandlerCloseReception(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}
