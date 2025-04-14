package mocks_test

import (
	"context"
	gomock "go.uber.org/mock/gomock"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"pvz_service/internal/app/product"
	"pvz_service/internal/app/product/mocks"
)

func TestServiceAddProduct_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceInterface(ctrl)
	ctx := context.Background()

	pvzID := uuid.New()
	productRequest := &product.CreateProductRequest{
		Type:  product.TypeElectronics,
		PvzID: pvzID,
	}

	expectedProduct := &product.Product{
		ID:          uuid.New(),
		Type:        product.TypeElectronics,
		ReceptionID: pvzID,
		DateTime:    time.Now(),
	}

	mockService.EXPECT().
		ServiceAddProduct(ctx, productRequest).
		Return(expectedProduct, nil)

	actualProduct, err := mockService.ServiceAddProduct(ctx, productRequest)

	assert.NoError(t, err)
	assert.NotNil(t, actualProduct)
	assert.Equal(t, expectedProduct.Type, actualProduct.Type)
	assert.Equal(t, expectedProduct.ReceptionID, actualProduct.ReceptionID)
}

func TestServiceAddProduct_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceInterface(ctrl)
	ctx := context.Background()

	pvzID := uuid.New()
	productRequest := &product.CreateProductRequest{
		Type:  product.TypeClothing,
		PvzID: pvzID,
	}

	mockService.EXPECT().
		ServiceAddProduct(ctx, productRequest).
		Return(nil, assert.AnError)

	productResult, err := mockService.ServiceAddProduct(ctx, productRequest)

	assert.Error(t, err)
	assert.Nil(t, productResult)
}

func TestServiceDeleteProduct_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceInterface(ctrl)
	ctx := context.Background()

	pvzID := uuid.New()
	expectedProduct := &product.Product{
		ID:          uuid.New(),
		Type:        product.TypeShoes,
		ReceptionID: pvzID,
		DateTime:    time.Now(),
	}

	mockService.EXPECT().
		ServiceDeleteProduct(ctx, pvzID).
		Return(expectedProduct, nil)

	actualProduct, err := mockService.ServiceDeleteProduct(ctx, pvzID)

	assert.NoError(t, err)
	assert.NotNil(t, actualProduct)
	assert.Equal(t, expectedProduct.Type, actualProduct.Type)
	assert.Equal(t, expectedProduct.ReceptionID, actualProduct.ReceptionID)
}

func TestServiceDeleteProduct_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceInterface(ctrl)
	ctx := context.Background()

	pvzID := uuid.New()

	mockService.EXPECT().
		ServiceDeleteProduct(ctx, pvzID).
		Return(nil, assert.AnError)

	productResult, err := mockService.ServiceDeleteProduct(ctx, pvzID)

	assert.Error(t, err)
	assert.Nil(t, productResult)
}
