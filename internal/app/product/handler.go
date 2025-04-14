package product

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
	myLogger       *slog.Logger
	productService ServiceInterface
}

func NewProductHandler(myLogger *slog.Logger, productService ServiceInterface) *Handler {
	return &Handler{
		myLogger:       myLogger,
		productService: productService,
	}
}

func (p *Handler) HandlerAddProduct(w http.ResponseWriter, r *http.Request) {
	const op = "app.product.HandlerAddProduct"

	myLogger := p.myLogger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var createProductRequest *CreateProductRequest
	err := json.NewDecoder(r.Body).Decode(&createProductRequest)
	if err != nil {
		myLogger.Error("Ошибка декодирования", sl.Err(err))
		response.MyResponseError(w, http.StatusBadRequest, "Неверный запрос или нет активной приемки")
		return
	}

	product, err := p.productService.ServiceAddProduct(r.Context(), createProductRequest)
	if err != nil {
		myLogger.Error("Ошибка добавления товара", sl.Err(err))
		response.MyResponseError(w, http.StatusBadRequest, "Неверный запрос или нет активной приемки")
		return
	}

	metrics.ProductsAdded.Inc()

	myLogger.Info("Товар добавлен", "id", product.ID)

	response.MyResponseJSON(w, http.StatusCreated, &Product{
		ID:          product.ID,
		DateTime:    product.DateTime,
		Type:        product.Type,
		ReceptionID: product.ReceptionID,
	})

	return
}

func (p *Handler) HandlerDeleteProduct(w http.ResponseWriter, r *http.Request) {
	const op = "app.product.HandlerDeleteProduct"

	myLogger := p.myLogger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	pvzIDInput := chi.URLParam(r, "pvzID")
	pvzID, err := uuid.Parse(pvzIDInput)
	if err != nil {
		myLogger.Error("Неверный pvzID", sl.Err(err))
		response.MyResponseError(
			w,
			http.StatusBadRequest,
			"Неверный запрос, нет активной приемки или нет товаров для удаления",
		)
		return
	}

	myLogger.Info("Получен pvzID", slog.String("pvzID", pvzID.String()))

	product, err := p.productService.ServiceDeleteProduct(r.Context(), pvzID)
	if err != nil {
		myLogger.Error("Ошибка удаления товара", sl.Err(err))
		response.MyResponseError(
			w,
			http.StatusBadRequest,
			"Неверный запрос, нет активной приемки или нет товаров для удаления",
		)
		return
	}

	myLogger.Info("Товар добавлен", "id", product.ID)

	response.MyResponseJSON(w, http.StatusOK, &Product{
		ID:          product.ID,
		DateTime:    product.DateTime,
		Type:        product.Type,
		ReceptionID: product.ReceptionID,
	})

	return
}
