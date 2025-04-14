package pvz

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"pvz_service/internal/lib/logger/sl"
)

// mockgen -source=internal/app/gen/service.go -destination=internal/app/gen/mocks/service_mock.go -package=mocks

type ServiceInterface interface {
	ServiceAddPVZ(ctx context.Context, pvzRequest *CreatePVZRequest) (*PVZ, error)
	ServiceGetPVZs(ctx context.Context, startDateStr, endDateStr, page, limit string) ([]*PVZWithReceptions, error)
}

type Service struct {
	pvzRepo  RepoInterface
	myLogger *slog.Logger
}

func NewPVZService(myLogger *slog.Logger, pvzRepo RepoInterface) *Service {
	return &Service{
		pvzRepo:  pvzRepo,
		myLogger: myLogger,
	}
}

func (p *Service) ServiceAddPVZ(ctx context.Context, pvzRequest *CreatePVZRequest) (*PVZ, error) {
	const op = "app.pvz.ServiceAddPVZ"

	myLogger := p.myLogger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(ctx)),
	)

	err := validatePVZRequest(pvzRequest)
	if err != nil {
		return nil, err
	}

	pvzID := uuid.New()

	pvz := &PVZ{
		ID:   pvzID,
		City: pvzRequest.City,
	}

	createdPVZ, err := p.pvzRepo.CreatePVZ(ctx, pvz)
	if err != nil {
		myLogger.Error("Ошибка создания ПВЗ", sl.Err(err))
		return nil, err
	}

	return createdPVZ, nil
}

func (p *Service) ServiceGetPVZs(
	ctx context.Context,
	startDateStr,
	endDateStr,
	page,
	limit string,
) ([]*PVZWithReceptions, error) {
	const op = "app.pvz.ServiceGetPVZs"

	myLogger := p.myLogger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(ctx)),
	)

	validatePage, err := ValidateParamsPage(page)
	if err != nil {
		myLogger.Error("Неверно указан параметр page: ", page, sl.Err(err))
		return nil, err
	}

	validateLimit, err := ValidateParamsLimit(limit)
	if err != nil {
		myLogger.Error("Неверно указан параметр limit: ", limit, sl.Err(err))
		return nil, err
	}

	startDate, err := ValidateDate(startDateStr)
	if err != nil {
		myLogger.Error("неверный формат startDate", sl.Err(err))
		return nil, fmt.Errorf("неверный формат startDate, должен быть RFC3339")
	}

	endDate, err := ValidateDate(endDateStr)
	if err != nil {
		myLogger.Error("неверный формат endDate", sl.Err(err))
		return nil, fmt.Errorf("неверный формат endDate, должен быть RFC3339")
	}

	res, err := p.pvzRepo.GetPVZWithReceptions(ctx, startDate, endDate, validatePage, validateLimit)
	if err != nil {
		myLogger.Error("Ошибка получения данных о ПВЗ", sl.Err(err))
		return nil, err
	}

	return res, nil
}

func validatePVZRequest(pvzRequest *CreatePVZRequest) error {
	validCity := []string{"Москва", "Санкт-Петербург", "Казань"}
	for _, city := range validCity {
		if pvzRequest.City == city {
			return nil
		}
	}

	return fmt.Errorf("недопустимый город: %s", pvzRequest.City)
}

func ValidateParamsPage(page string) (int, error) {
	if page == "" {
		return 1, nil
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return 0, err
	}

	if pageInt < 1 {
		return 0, errors.New("page должен быть ≥ 1")
	}

	return pageInt, nil
}

func ValidateParamsLimit(limit string) (int, error) {
	if limit == "" {
		return 10, nil
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return 0, fmt.Errorf("limit должен быть числом: %w", err)
	}

	if limitInt < 1 || limitInt > 30 {
		return 0, errors.New("limit должен быть от 1 до 30 включительно")
	}

	return limitInt, nil
}

func ValidateDate(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}

	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return nil, fmt.Errorf("неверный формат даты, должен быть RFC3339: %w", err)
	}

	return &t, nil
}
