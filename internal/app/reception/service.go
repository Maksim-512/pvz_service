package reception

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"pvz_service/internal/lib/logger/sl"
)

// mockgen -source=internal/app/reception/service.go -destination=internal/app/reception/mocks/service_mock.go -package=mocks

type ServiceInterface interface {
	ServiceOpenReception(ctx context.Context, receptionRequest *OpenReceptionRequest) (*Reception, error)
	ServiceCloseReception(ctx context.Context, receptionRequest *CloseReceptionRequest) (*Reception, error)
}

type Service struct {
	receptionRepo RepoInterface
	myLogger      *slog.Logger
}

func NewReceptionService(myLogger *slog.Logger, receptionRepo RepoInterface) *Service {
	return &Service{
		receptionRepo: receptionRepo,
		myLogger:      myLogger,
	}
}

func (r *Service) ServiceOpenReception(
	ctx context.Context,
	receptionRequest *OpenReceptionRequest,
) (*Reception, error) {
	const op = "app.reception.ServiceOpenReception"

	myLogger := r.myLogger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(ctx)),
	)

	checkPVZ, err := r.receptionRepo.CheckValidPVZID(ctx, receptionRequest.PvzID)
	if err != nil {
		myLogger.Error("Ошибка при проверке на наличие пункта PVZ", sl.Err(err))
		return nil, fmt.Errorf("ошибка при проверке на наличие пункта PVZ: %w", err)
	}

	if !checkPVZ {
		myLogger.Error("Невалидный PvzID")
		return nil, errors.New("неверный запрос или есть незакрытая приемка")
	}

	checkOpenReception, receptionID, err := r.receptionRepo.CheckOpenReception(ctx, receptionRequest.PvzID)
	if err != nil {
		myLogger.Error("Ошибка при проверке на наличие открытой приемки", sl.Err(err))
		return nil, fmt.Errorf("ошибка при проверке на наличие открытой приемки: %w", err)
	}

	if checkOpenReception {
		myLogger.Error(
			"Нельзя создать приемку, так как есть незакрытая id: %v",
			slog.String("receptionID", receptionID.String()),
		)
		return nil, errors.New("неверный запрос или есть незакрытая приемка")
	}

	reception := &Reception{
		ID:    uuid.New(),
		PvzID: receptionRequest.PvzID,
	}

	openReception, err := r.receptionRepo.CreateReception(ctx, reception)
	if err != nil {
		myLogger.Error("Ошибка создания приемки", sl.Err(err))
		return nil, fmt.Errorf("oшибка создания приемки: %w", err)
	}

	return openReception, nil
}

func (r *Service) ServiceCloseReception(
	ctx context.Context,
	receptionRequest *CloseReceptionRequest,
) (*Reception, error) {
	const op = "app.reception.ServiceCloseReception"

	myLogger := r.myLogger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(ctx)),
	)

	checkPVZ, err := r.receptionRepo.CheckValidPVZID(ctx, receptionRequest.PvzID)
	if err != nil {
		myLogger.Error("Ошибка при проверке на наличие пункта PVZ", sl.Err(err))
		return nil, fmt.Errorf("oшибка при проверке на наличие пункта PVZ: %w", err)
	}

	if !checkPVZ {
		myLogger.Error("Невалидный PvzID")
		return nil, errors.New("неверный запрос или приемка уже закрыта")
	}

	checkOpenReception, receptionID, err := r.receptionRepo.CheckOpenReception(ctx, receptionRequest.PvzID)
	if err != nil {
		myLogger.Error("Ошибка при проверке на наличие открытой приемки", sl.Err(err))
		return nil, fmt.Errorf("ошибка при проверке на наличие открытой приемки: %w", err)
	}

	if !checkOpenReception {
		myLogger.Error("Нельзя закрыть проверку, так как нет открытых")
		return nil, errors.New("неверный запрос или приемка уже закрыта")
	}

	closeReception, err := r.receptionRepo.CloseReception(ctx, receptionID)
	if err != nil {
		myLogger.Error("Ошибка закрытия приемки", sl.Err(err))
		return nil, fmt.Errorf("oшибка закрытия приемки: %w", err)
	}

	return closeReception, nil
}
