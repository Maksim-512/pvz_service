package product_test

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
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"pvz_service/internal/app/product"
	"pvz_service/internal/app/product/mocks"
	"pvz_service/pkg/response"
)

func TestHandlerAddProduct_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductService := mocks.NewMockServiceInterface(ctrl)
	mockProductService.EXPECT().ServiceAddProduct(gomock.Any(), gomock.Any()).Return(&product.Product{
		ID:          uuid.New(),
		DateTime:    time.Now(),
		Type:        product.TypeShoes,
		ReceptionID: uuid.MustParse("d730a5b3-7ea9-494a-ac5e-2f5dd1aaa83b"),
	}, nil)

	handler := product.NewProductHandler(slog.Default(), mockProductService)

	createProductRequest := product.CreateProductRequest{
		Type:  product.TypeShoes,
		PvzID: uuid.MustParse("d730a5b3-7ea9-494a-ac5e-2f5dd1aaa83b"),
	}

	body, _ := json.Marshal(createProductRequest)
	req := httptest.NewRequest("POST", "/products", bytes.NewReader(body))

	rr := httptest.NewRecorder()

	handler.HandlerAddProduct(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var responseBody map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&responseBody)
	assert.NoError(t, err)

	expectedReceptionID := "d730a5b3-7ea9-494a-ac5e-2f5dd1aaa83b"
	actualReceptionID := responseBody["receptionID"].(string)

	assert.Equal(t, expectedReceptionID, actualReceptionID)
	assert.Equal(t, product.TypeShoes, responseBody["type"])
}

func TestHandlerAddProduct_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductService := mocks.NewMockServiceInterface(ctrl)
	handler := product.NewProductHandler(slog.Default(), mockProductService)

	req := httptest.NewRequest("POST", "/products", bytes.NewReader([]byte(`{invalid-json}`)))

	rr := httptest.NewRecorder()

	handler.HandlerAddProduct(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandlerAddProduct_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductService := mocks.NewMockServiceInterface(ctrl)

	handler := product.NewProductHandler(slog.Default(), mockProductService)

	requestData := &product.CreateProductRequest{
		Type:  "обувь",
		PvzID: uuid.New(),
	}

	mockProductService.EXPECT().ServiceAddProduct(gomock.Any(), requestData).Return(
		nil,
		errors.New("ошибка добавления товара"),
	)

	reqBody, _ := json.Marshal(requestData)
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	req = req.WithContext(context.WithValue(req.Context(), "role", "employee"))

	rr := httptest.NewRecorder()

	handler.HandlerAddProduct(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp response.ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.Nil(t, err)
	assert.Equal(t, "Неверный запрос или нет активной приемки", resp.Errors)
}

func TestHandlerDeleteProduct_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductService := mocks.NewMockServiceInterface(ctrl)
	handler := product.NewProductHandler(slog.Default(), mockProductService)

	testProduct := &product.Product{
		ID:          uuid.New(),
		Type:        "обувь",
		ReceptionID: uuid.New(),
		DateTime:    time.Now(),
	}

	pvzID := uuid.New()
	mockProductService.EXPECT().ServiceDeleteProduct(gomock.Any(), gomock.Any()).Return(testProduct, nil)

	req := httptest.NewRequest(http.MethodPost, "/pvz/"+pvzID.String()+"/delete_last_product", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("pvzID", pvzID.String())

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(context.WithValue(req.Context(), "role", "employee"))

	rr := httptest.NewRecorder()

	handler.HandlerDeleteProduct(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp product.Product
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.Nil(t, err)
	assert.Equal(t, testProduct.Type, resp.Type)
}

func TestHandlerDeleteProduct_InvalidPVZID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	handler := product.NewProductHandler(slog.Default(), nil)

	invalidPvzID := "invalid"

	req := httptest.NewRequest(http.MethodPost, "/pvz/"+invalidPvzID+"/delete_last_product", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("pvzID", invalidPvzID)

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(context.WithValue(req.Context(), "role", "employee"))

	rr := httptest.NewRecorder()

	handler.HandlerDeleteProduct(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp response.ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.Nil(t, err)
	assert.Equal(t, "Неверный запрос, нет активной приемки или нет товаров для удаления", resp.Errors)
}

func TestHandlerDeleteProduct_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductService := mocks.NewMockServiceInterface(ctrl)
	handler := product.NewProductHandler(slog.Default(), mockProductService)

	pvzID := uuid.New()

	mockProductService.EXPECT().ServiceDeleteProduct(gomock.Any(), pvzID).Return(
		nil,
		errors.New("ошибка удаления товара"),
	)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("pvzID", pvzID.String())

	req := httptest.NewRequest(http.MethodPost, "/pvz/"+pvzID.String()+"/delete_last_product", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(context.WithValue(req.Context(), "role", "employee"))
	req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "tests-req-id"))

	rr := httptest.NewRecorder()

	handler.HandlerDeleteProduct(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp response.ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.Nil(t, err)
	assert.Equal(t, "Неверный запрос, нет активной приемки или нет товаров для удаления", resp.Errors)
}
