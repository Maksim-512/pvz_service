package product

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"pvz_service/internal/app/reception"
	"pvz_service/internal/lib/logger/sl"
)

// mockgen -source=internal/app/product/service.go -destination=internal/app/product/mocks/service_mock.go -package=mocks

type ServiceInterface interface {
	ServiceAddProduct(ctx context.Context, productRequest *CreateProductRequest) (*Product, error)
	ServiceDeleteProduct(ctx context.Context, pvzID uuid.UUID) (*Product, error)
}

type Service struct {
	productRepo   RepoInterface
	receptionRepo reception.RepoInterface
	myLogger      *slog.Logger
}

func NewProductService(myLogger *slog.Logger, productRepo RepoInterface, receptionRepo reception.RepoInterface) *Service {
	return &Service{
		productRepo:   productRepo,
		myLogger:      myLogger,
		receptionRepo: receptionRepo,
	}
}

func (p *Service) ServiceAddProduct(ctx context.Context, productRequest *CreateProductRequest) (*Product, error) {
	const op = "app.product.ServiceAddProduct"

	myLogger := p.myLogger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(ctx)),
	)

	err := validateProductRequest(productRequest)
	if err != nil {
		return nil, err
	}

	checkPVZ, err := p.receptionRepo.CheckValidPVZID(ctx, productRequest.PvzID)
	if err != nil {
		myLogger.Error("Ошибка при проверке на наличие пункта PVZ", sl.Err(err))
		return nil, err
	}

	if !checkPVZ {
		myLogger.Error("Невалидный PvzID")
		return nil, errors.New("неверный запрос или нет активной приемки")
	}

	checkOpenReception, receptionID, err := p.receptionRepo.CheckOpenReception(ctx, productRequest.PvzID)
	if err != nil {
		myLogger.Error("Ошибка при проверке на наличие открытой приемки", sl.Err(err))
		return nil, err
	}

	if !checkOpenReception {
		myLogger.Error("Нельзя добавить товар, так как нет открытых приемок")
		return nil, errors.New("ошибка при проверке на наличие открытой приемки")
	}

	product := &Product{
		ID:          uuid.New(),
		Type:        productRequest.Type,
		ReceptionID: receptionID,
	}

	addProduct, err := p.productRepo.AddProduct(ctx, product)
	if err != nil {
		myLogger.Error("Ошибка добавления товара", sl.Err(err))
		return nil, err
	}
	return addProduct, nil
}

func (p *Service) ServiceDeleteProduct(ctx context.Context, pvzID uuid.UUID) (*Product, error) {
	const op = "app.product.ServiceDeleteProduct"

	myLogger := p.myLogger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(ctx)),
	)

	checkPVZ, err := p.receptionRepo.CheckValidPVZID(ctx, pvzID)
	if err != nil {
		myLogger.Error("Ошибка при проверке на наличие пункта PVZ", sl.Err(err))
		return nil, err
	}

	if !checkPVZ {
		myLogger.Error("Невалидный PvzID")
		return nil, errors.New("неверный запрос или нет активной приемки")
	}

	checkOpenReception, receptionID, err := p.receptionRepo.CheckOpenReception(ctx, pvzID)
	if err != nil {
		myLogger.Error("Ошибка при проверке на наличие открытой приемки", sl.Err(err))
		return nil, err
	}

	if !checkOpenReception {
		myLogger.Error("Нельзя добавить товар, так как нет открытых приемок")
		return nil, errors.New("ошибка при проверке на наличие открытой приемки")
	}

	product, err := p.productRepo.DeleteProduct(ctx, receptionID)
	if err != nil {
		myLogger.Error("Ошибка удаления товара", sl.Err(err))
		return nil, err
	}

	return product, nil
}

func validateProductRequest(pvzRequest *CreateProductRequest) error {
	validProduct := []string{TypeElectronics, TypeClothing, TypeShoes}
	for _, product := range validProduct {
		if pvzRequest.Type == product {
			return nil
		}
	}

	return fmt.Errorf("недопустимый товар: %s", pvzRequest.Type)
}
