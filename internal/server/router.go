package server

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"pvz_service/internal/app/auth"
	"pvz_service/internal/app/product"
	"pvz_service/internal/app/pvz"
	"pvz_service/internal/app/reception"
	"pvz_service/pkg/jwt"
	mwAuth "pvz_service/pkg/middleware/auth"
	mwMetrics "pvz_service/pkg/middleware/metrics"
	"pvz_service/pkg/middleware/mwLogger"
)

func NewRouter(
	myLogger *slog.Logger,
	jwtService *jwt.JWTService,
	authHandler *auth.Handler,
	pvzHandler *pvz.Handler,
	receptionHandler *reception.Handler,
	productHandler *product.Handler,
) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(myLogger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(mwMetrics.MetricsMiddleware)

	router.Post("/register", authHandler.HandlerRegister)
	router.Post("/login", authHandler.HandlerLogin)
	router.Post("/dummyLogin", authHandler.HandlerDummyLogin)
	router.With(
		mwAuth.CheckAuthMiddleware(jwtService),
		mwAuth.RoleMiddle(auth.RoleModerator),
	).Post("/pvz", pvzHandler.HandlerAddPVZ)
	router.With(
		mwAuth.CheckAuthMiddleware(jwtService),
		mwAuth.RoleMiddle(auth.RoleModerator, auth.RoleEmployee),
	).Get("/pvz", pvzHandler.HandlerGetPVZ)
	router.With(
		mwAuth.CheckAuthMiddleware(jwtService),
		mwAuth.RoleMiddle(auth.RoleEmployee),
	).Post("/receptions", receptionHandler.HandlerOpenReception)
	router.With(
		mwAuth.CheckAuthMiddleware(jwtService),
		mwAuth.RoleMiddle(auth.RoleEmployee),
	).Post("/pvz/{pvzID}/close_last_reception", receptionHandler.HandlerCloseReception)
	router.With(
		mwAuth.CheckAuthMiddleware(jwtService),
		mwAuth.RoleMiddle(auth.RoleEmployee),
	).Post("/products", productHandler.HandlerAddProduct)
	router.With(
		mwAuth.CheckAuthMiddleware(jwtService),
		mwAuth.RoleMiddle(auth.RoleEmployee),
	).Post("/pvz/{pvzID}/delete_last_product", productHandler.HandlerDeleteProduct)

	return router
}
