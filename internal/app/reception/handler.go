package reception

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"pvz_service/internal/lib/logger/sl"
	"pvz_service/pkg/metrics"
	"pvz_service/pkg/response"
)

type Handler struct {
	myLogger         *slog.Logger
	receptionService ServiceInterface
}

func NewReceptionHandler(myLogger *slog.Logger, receptionService ServiceInterface) *Handler {
	return &Handler{
		myLogger:         myLogger,
		receptionService: receptionService,
	}
}

func (rec *Handler) HandlerOpenReception(w http.ResponseWriter, r *http.Request) {
	const op = "app.reception.HandlerOpenReception"

	myLogger := rec.myLogger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var receptionRequest *OpenReceptionRequest
	err := json.NewDecoder(r.Body).Decode(&receptionRequest)
	if err != nil {
		myLogger.Error("Ошибка декодирования", sl.Err(err))
		response.MyResponseError(w, http.StatusBadRequest, "Неверный запрос или есть незакрытая приемка")
		return
	}

	openReception, err := rec.receptionService.ServiceOpenReception(r.Context(), receptionRequest)
	if err != nil {
		myLogger.Error("Ошибка открытия приемки", sl.Err(err))
		response.MyResponseError(w, http.StatusBadRequest, "Неверный запрос или есть незакрытая приемка")
		return
	}

	metrics.ReceptionsCreated.Inc()

	myLogger.Info("Приемка открыта", "id", openReception.ID)

	response.MyResponseJSON(w, http.StatusCreated, &Reception{
		ID:       openReception.ID,
		DateTime: openReception.DateTime,
		PvzID:    openReception.PvzID,
		Status:   openReception.Status,
	})
}

func (rec *Handler) HandlerCloseReception(w http.ResponseWriter, r *http.Request) {
	const op = "app.reception.HandlerCloseReception"

	myLogger := rec.myLogger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	pvzIDInput := chi.URLParam(r, "pvzID")
	pvzID, err := uuid.Parse(pvzIDInput)
	if err != nil {
		myLogger.Error("Неверный UUID ПВЗ", sl.Err(err))
		response.MyResponseError(w, http.StatusBadRequest, "Неверный запрос или приемка уже закрыта")
		return
	}

	receptionRequest := &CloseReceptionRequest{
		PvzID: pvzID,
	}

	closeReception, err := rec.receptionService.ServiceCloseReception(r.Context(), receptionRequest)
	if err != nil {
		myLogger.Error("Ошибка закрытия приемки", sl.Err(err))
		response.MyResponseError(w, http.StatusBadRequest, "Неверный запрос или приемка уже закрыта")
		return
	}

	myLogger.Info("Приемка закрыта", "id", closeReception.ID)

	response.MyResponseJSON(w, http.StatusOK, &Reception{
		ID:       closeReception.ID,
		DateTime: closeReception.DateTime,
		PvzID:    closeReception.PvzID,
		Status:   closeReception.Status,
	})
}
