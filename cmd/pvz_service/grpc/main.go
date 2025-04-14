package main

import (
	"fmt"
	"log"
	"net"

	grpcServ "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"pvz_service/internal/app/pvz"
	"pvz_service/internal/config"
	"pvz_service/internal/lib/logger"
	"pvz_service/internal/lib/logger/sl"
	"pvz_service/internal/server/grpc"
	"pvz_service/internal/storage/postgres"
	pvzV1 "pvz_service/protos/gen/go"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatal("Ошибка загрузки конфига: ", err)
	}

	myLogger := logger.SetupLogger(cfg.Logging.Env)

	storage, err := postgres.New(&cfg.Database)
	if err != nil {
		myLogger.Error("Ошибка подключения к БД", sl.Err(err))
		return
	}
	defer storage.Close()

	repo := pvz.NewPVZRepository(storage)
	service := pvz.NewPVZService(myLogger, repo)

	addr := fmt.Sprintf(":%s", cfg.GRPC.Port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("не удалось запустить gRPC сервер: %v", err)
	}

	grpcServer := grpc.NewGRPCServer(service)

	s := grpcServ.NewServer()
	pvzV1.RegisterPVZServiceServer(s, grpcServer)

	reflection.Register(s)

	log.Printf("gRPC сервер запущен на порту %s", addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("ошибка при запуске gRPC сервера: %v", err)
	}
}
