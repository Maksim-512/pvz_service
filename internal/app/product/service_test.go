package product_test

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/mock/gomock"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"pvz_service/internal/app/product"
	"pvz_service/internal/app/product/mocks"
	rece_mock "pvz_service/internal/app/reception/mocks"
	mock_jwt "pvz_service/pkg/jwt/mocks"
)

func newServiceWithMocks(t *testing.T) (
	*product.Service,
	*mocks.MockRepoInterface,
	*rece_mock.MockRepoInterface,
	*mock_jwt.MockTokenService,
) {
	ctrl := gomock.NewController(t)
	productRepoMock := mocks.NewMockRepoInterface(ctrl)
	receptionRepoMock := rece_mock.NewMockRepoInterface(ctrl)
	logger := slog.Default()
	mockToken := mock_jwt.NewMockTokenService(ctrl)

	service := product.NewProductService(logger, productRepoMock, receptionRepoMock)

	return service, productRepoMock, receptionRepoMock, mockToken
}

func TestServiceAddProduct_Success(t *testing.T) {
	service, productRepoMock, receptionRepoMock, _ := newServiceWithMocks(t)

	pvzID := uuid.New()
	productRequest := &product.CreateProductRequest{
		Type:  product.TypeElectronics,
		PvzID: pvzID,
	}

	receptionRepoMock.EXPECT().CheckValidPVZID(context.Background(), pvzID).Return(true, nil)
	receptionRepoMock.EXPECT().CheckOpenReception(context.Background(), pvzID).Return(true, uuid.New(), nil)
	productRepoMock.EXPECT().AddProduct(context.Background(), gomock.Any()).Return(&product.Product{
		ID:          uuid.New(),
		Type:        product.TypeElectronics,
		ReceptionID: uuid.New(),
	}, nil)

	myProduct, err := service.ServiceAddProduct(context.Background(), productRequest)

	assert.NoError(t, err)
	assert.NotNil(t, myProduct)
	assert.Equal(t, product.TypeElectronics, myProduct.Type)
}

func TestServiceAddProduct_InvalidPVZID(t *testing.T) {
	service, _, receptionRepoMock, _ := newServiceWithMocks(t)

	pvzID := uuid.New()
	productRequest := &product.CreateProductRequest{
		Type:  product.TypeElectronics,
		PvzID: pvzID,
	}

	receptionRepoMock.EXPECT().CheckValidPVZID(context.Background(), pvzID).Return(false, nil)

	myProduct, err := service.ServiceAddProduct(context.Background(), productRequest)

	assert.Error(t, err)
	assert.Nil(t, myProduct)
	assert.Equal(t, "неверный запрос или нет активной приемки", err.Error())
}

func TestServiceAddProduct_Error_CheckValidPVZID(t *testing.T) {
	service, _, receptionRepoMock, _ := newServiceWithMocks(t)

	pvzID := uuid.New()
	productRequest := &product.CreateProductRequest{
		Type:  product.TypeElectronics,
		PvzID: pvzID,
	}

	receptionRepoMock.EXPECT().CheckValidPVZID(context.Background(), pvzID).Return(
		false,
		errors.New("ошибка при проверке на наличие пункта PVZ"),
	)

	_, err := service.ServiceAddProduct(context.Background(), productRequest)

	assert.Error(t, err)
	assert.Equal(t, "ошибка при проверке на наличие пункта PVZ", err.Error())
}

func TestServiceAddProduct_Error_CheckOpenReception(t *testing.T) {
	service, _, receptionRepoMock, _ := newServiceWithMocks(t)

	pvzID := uuid.New()
	productRequest := &product.CreateProductRequest{
		Type:  product.TypeElectronics,
		PvzID: pvzID,
	}

	receptionRepoMock.EXPECT().CheckValidPVZID(context.Background(), pvzID).Return(true, nil)
	receptionRepoMock.EXPECT().CheckOpenReception(context.Background(), pvzID).Return(
		false,
		uuid.New(),
		errors.New("ошибка при проверке на наличие открытой приемки"),
	)

	_, err := service.ServiceAddProduct(context.Background(), productRequest)

	assert.Error(t, err)
	assert.Equal(t, "ошибка при проверке на наличие открытой приемки", err.Error())
}

func TestServiceAddProduct_Error_NoOpenReception(t *testing.T) {
	service, _, receptionRepoMock, _ := newServiceWithMocks(t)

	pvzID := uuid.New()
	productRequest := &product.CreateProductRequest{
		Type:  product.TypeElectronics,
		PvzID: pvzID,
	}

	receptionRepoMock.EXPECT().CheckValidPVZID(context.Background(), pvzID).Return(true, nil)
	receptionRepoMock.EXPECT().CheckOpenReception(context.Background(), pvzID).Return(false, uuid.New(), nil)

	_, err := service.ServiceAddProduct(context.Background(), productRequest)

	assert.Error(t, err)
	assert.Equal(t, "ошибка при проверке на наличие открытой приемки", err.Error())
}

func TestServiceAddProduct_Error_ValidateProductRequest(t *testing.T) {
	service, _, _, _ := newServiceWithMocks(t)

	productRequest := &product.CreateProductRequest{
		Type:  "",
		PvzID: uuid.New(),
	}

	myProduct, err := service.ServiceAddProduct(context.Background(), productRequest)

	assert.Error(t, err)
	assert.Nil(t, myProduct)

	resp := fmt.Sprintf("недопустимый товар: %s", productRequest.Type)
	assert.Equal(t, resp, err.Error())
}

func TestServiceAddProduct_Error_AddProduct(t *testing.T) {
	service, productRepoMock, receptionRepoMock, _ := newServiceWithMocks(t)

	pvzID := uuid.New()
	productRequest := &product.CreateProductRequest{
		Type:  product.TypeElectronics,
		PvzID: pvzID,
	}

	receptionRepoMock.EXPECT().CheckValidPVZID(context.Background(), pvzID).Return(true, nil)
	receptionRepoMock.EXPECT().CheckOpenReception(context.Background(), pvzID).Return(true, uuid.New(), nil)
	productRepoMock.EXPECT().AddProduct(context.Background(), gomock.Any()).Return(
		nil,
		errors.New("ошибка добавления товара"),
	)

	_, err := service.ServiceAddProduct(context.Background(), productRequest)

	assert.Error(t, err)
	assert.Equal(t, "ошибка добавления товара", err.Error())
}

func TestServiceDeleteProduct_Success(t *testing.T) {
	service, productRepoMock, receptionRepoMock, _ := newServiceWithMocks(t)

	pvzID := uuid.New()

	receptionRepoMock.EXPECT().CheckValidPVZID(context.Background(), pvzID).Return(true, nil)
	receptionRepoMock.EXPECT().CheckOpenReception(context.Background(), pvzID).Return(true, uuid.New(), nil)
	productRepoMock.EXPECT().DeleteProduct(context.Background(), gomock.Any()).Return(&product.Product{
		ID:          uuid.New(),
		Type:        product.TypeElectronics,
		ReceptionID: uuid.New(),
	}, nil)

	myProduct, err := service.ServiceDeleteProduct(context.Background(), pvzID)
	assert.NoError(t, err)
	assert.NotNil(t, myProduct)
}

func TestServiceDeleteProduct_InvalidPVZID(t *testing.T) {
	service, _, receptionRepoMock, _ := newServiceWithMocks(t)

	pvzID := uuid.New()

	receptionRepoMock.EXPECT().CheckValidPVZID(context.Background(), pvzID).Return(false, nil)

	myProduct, err := service.ServiceDeleteProduct(context.Background(), pvzID)

	assert.Error(t, err)
	assert.Nil(t, myProduct)
	assert.Equal(t, "неверный запрос или нет активной приемки", err.Error())
}

func TestServiceDeleteProduct_Error_CheckValidPVZID(t *testing.T) {
	service, _, receptionRepoMock, _ := newServiceWithMocks(t)

	pvzID := uuid.New()

	receptionRepoMock.EXPECT().CheckValidPVZID(context.Background(), pvzID).Return(
		false,
		errors.New("ошибка при проверке PVZ"),
	)

	myProduct, err := service.ServiceDeleteProduct(context.Background(), pvzID)

	assert.Error(t, err)
	assert.Nil(t, myProduct)
	assert.Equal(t, "ошибка при проверке PVZ", err.Error())
}

func TestServiceDeleteProduct_Error_NoOpenReception(t *testing.T) {
	service, _, receptionRepoMock, _ := newServiceWithMocks(t)

	pvzID := uuid.New()

	receptionRepoMock.EXPECT().CheckValidPVZID(context.Background(), pvzID).Return(true, nil)
	receptionRepoMock.EXPECT().CheckOpenReception(context.Background(), pvzID).Return(false, uuid.New(), nil)

	myProduct, err := service.ServiceDeleteProduct(context.Background(), pvzID)

	assert.Error(t, err)
	assert.Nil(t, myProduct)
	assert.Equal(t, "ошибка при проверке на наличие открытой приемки", err.Error())
}

func TestServiceDeleteProduct_Error_CheckOpenReception(t *testing.T) {
	service, _, receptionRepoMock, _ := newServiceWithMocks(t)

	pvzID := uuid.New()

	receptionRepoMock.EXPECT().CheckValidPVZID(context.Background(), pvzID).Return(true, nil)
	receptionRepoMock.EXPECT().CheckOpenReception(context.Background(), pvzID).Return(
		false,
		uuid.New(),
		errors.New("ошибка при получении открытой приемки"),
	)

	myProduct, err := service.ServiceDeleteProduct(context.Background(), pvzID)

	assert.Error(t, err)
	assert.Nil(t, myProduct)
	assert.Equal(t, "ошибка при получении открытой приемки", err.Error())
}

func TestServiceDeleteProduct_Error_DeleteProduct(t *testing.T) {
	service, productRepoMock, receptionRepoMock, _ := newServiceWithMocks(t)

	pvzID := uuid.New()

	receptionRepoMock.EXPECT().CheckValidPVZID(context.Background(), pvzID).Return(true, nil)
	receptionRepoMock.EXPECT().CheckOpenReception(context.Background(), pvzID).Return(true, uuid.New(), nil)
	productRepoMock.EXPECT().DeleteProduct(context.Background(), gomock.Any()).Return(
		nil,
		errors.New("ошибка при удалении товара"),
	)

	myProduct, err := service.ServiceDeleteProduct(context.Background(), pvzID)

	assert.Error(t, err)
	assert.Nil(t, myProduct)
	assert.Equal(t, "ошибка при удалении товара", err.Error())
}
