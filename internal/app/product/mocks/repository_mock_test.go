package mocks_test

import (
	"context"
	"errors"
	gomock "go.uber.org/mock/gomock"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"pvz_service/internal/app/product"
	"pvz_service/internal/app/product/mocks"
)

func TestMockRepoInterface_AddProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepoInterface(ctrl)
	productToAdd := &product.Product{ID: uuid.New(), Type: product.TypeElectronics, ReceptionID: uuid.New()}

	mockRepo.EXPECT().AddProduct(context.Background(), productToAdd).Return(productToAdd, nil)

	addedProduct, err := mockRepo.AddProduct(context.Background(), productToAdd)

	assert.NoError(t, err)
	assert.Equal(t, productToAdd, addedProduct)
}

func TestMockRepoInterface_AddProduct_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepoInterface(ctrl)
	productToAdd := &product.Product{ID: uuid.New(), Type: product.TypeElectronics, ReceptionID: uuid.New()}

	mockRepo.EXPECT().AddProduct(context.Background(), productToAdd).Return(nil, assert.AnError)

	addedProduct, err := mockRepo.AddProduct(context.Background(), productToAdd)

	assert.Error(t, err)
	assert.Nil(t, addedProduct)
}

func TestMockRepoInterface_DeleteProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepoInterface(ctrl)
	receptionID := uuid.New()
	productToDelete := &product.Product{ID: uuid.New(), Type: product.TypeElectronics, ReceptionID: uuid.New()}

	mockRepo.EXPECT().DeleteProduct(context.Background(), receptionID).Return(productToDelete, nil)

	deletedProduct, err := mockRepo.DeleteProduct(context.Background(), receptionID)

	assert.NoError(t, err)
	assert.Equal(t, productToDelete, deletedProduct)
}

func TestMockRepoInterface_DeleteProduct_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepoInterface(ctrl)
	receptionID := uuid.New()

	mockRepo.EXPECT().DeleteProduct(context.Background(), receptionID).Return(nil, errors.New("нет такого продукта"))

	deletedProduct, err := mockRepo.DeleteProduct(context.Background(), receptionID)

	assert.Error(t, err)
	assert.Nil(t, deletedProduct)
}
