package pvz

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"pvz_service/internal/lib/logger/sl"
	"pvz_service/pkg/metrics"
	"pvz_service/pkg/response"
)

type Handler struct {
	myLogger   *slog.Logger
	pvzService ServiceInterface
}

func NewPVZHandler(myLogger *slog.Logger, pvzService ServiceInterface) *Handler {
	return &Handler{
		myLogger:   myLogger,
		pvzService: pvzService,
	}
}

func (p *Handler) HandlerAddPVZ(w http.ResponseWriter, r *http.Request) {
	const op = "app.pvz.HandlerAddPVZ"

	myLogger := p.myLogger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	role, ok := r.Context().Value("role").(string)
	if !ok {
		myLogger.Error("Не была передана role")
		response.MyResponseError(w, http.StatusBadRequest, "Неверный запрос")
		return
	}

	myLogger.Info("Пользователь с ролью: ", slog.String("role", role))

	var pvzRequest *CreatePVZRequest
	err := json.NewDecoder(r.Body).Decode(&pvzRequest)
	if err != nil {
		myLogger.Error("Ошибка декодирования", sl.Err(err))
		response.MyResponseError(w, http.StatusBadRequest, "Неверный запрос")
		return
	}

	createdPVZ, err := p.pvzService.ServiceAddPVZ(r.Context(), pvzRequest)
	if err != nil {
		myLogger.Error("ошибка создания ПВЗ", sl.Err(err))
		response.MyResponseError(w, http.StatusBadRequest, "Неверный запрос")
		return
	}

	metrics.PVZCreated.Inc()

	myLogger.Info("ПВЗ создан", "id", createdPVZ.ID)

	response.MyResponseJSON(w, http.StatusCreated, &PVZ{
		ID:               createdPVZ.ID,
		RegistrationDate: createdPVZ.RegistrationDate,
		City:             createdPVZ.City,
	})

}

func (p *Handler) HandlerGetPVZ(w http.ResponseWriter, r *http.Request) {
	const op = "app.pvz.HandlerGetPVZ"

	myLogger := p.myLogger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	query := r.URL.Query()

	page := query.Get("page")
	limit := query.Get("limit")
	startDate := query.Get("startDate")
	endDate := query.Get("endDate")

	result, err := p.pvzService.ServiceGetPVZs(r.Context(), startDate, endDate, page, limit)
	if err != nil {
		myLogger.Error("ошибка получения списка ПВЗ", "err", err)
		response.MyResponseError(w, http.StatusBadRequest, "Неверный запрос")
		return
	}

	response.MyResponseJSON(w, http.StatusOK, result)
}
