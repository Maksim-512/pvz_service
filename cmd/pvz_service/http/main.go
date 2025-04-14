package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"pvz_service/internal/app/auth"
	"pvz_service/internal/app/product"
	"pvz_service/internal/app/pvz"
	"pvz_service/internal/app/reception"
	"pvz_service/internal/config"
	"pvz_service/internal/lib/logger"
	"pvz_service/internal/lib/logger/sl"
	"pvz_service/internal/server"
	"pvz_service/internal/storage/postgres"
	"pvz_service/pkg/jwt"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatal("Ошибка загрузки конфига: ", err)
	}

	jwtService := jwt.NewJWTService(cfg.Auth.SecretKey)

	myLogger := logger.SetupLogger(cfg.Logging.Env)

	//promPort := fmt.Sprintf(":%s", cfg.Prometheus.Port)
	//metrics.Init(promPort)

	myLogger.Info("Конфигурация логера подгружена")

	storage, err := postgres.New(&cfg.Database)
	if err != nil {
		myLogger.Error("Ошибка подключения к БД", sl.Err(err))
		return
	}
	defer storage.Close()

	ctx := context.Background()
	err = storage.Ping(ctx)
	if err != nil {
		myLogger.Error("БД не пингуется", sl.Err(err))
		return
	}

	myLogger.Info("БД подключена")

	userRepo := auth.NewUserRepository(storage)
	authService := auth.NewAuthService(myLogger, userRepo, jwtService)
	authHandler := auth.NewAuthHandler(myLogger, authService)

	pvzRepo := pvz.NewPVZRepository(storage)
	pvzService := pvz.NewPVZService(myLogger, pvzRepo)
	pvzHandler := pvz.NewPVZHandler(myLogger, pvzService)

	receptionRepo := reception.NewReceptionRepository(storage)
	receptionService := reception.NewReceptionService(myLogger, receptionRepo)
	receptionHandler := reception.NewReceptionHandler(myLogger, receptionService)

	productRepo := product.NewProductRepository(storage)
	productService := product.NewProductService(myLogger, productRepo, receptionRepo)
	productHandler := product.NewProductHandler(myLogger, productService)

	router := server.NewRouter(myLogger, jwtService, authHandler, pvzHandler, receptionHandler, productHandler)

	httpAddr := fmt.Sprintf(":%s", cfg.HTTP.Port)
	srv := &http.Server{
		Addr:         httpAddr,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	myLogger.Info("Запуск сервера")

	err = srv.ListenAndServe()
	if err != nil {
		myLogger.Error("Ошибка запуска сервера", sl.Err(err))
		return
	}

	myLogger.Error("Сервер остановлен")

}
