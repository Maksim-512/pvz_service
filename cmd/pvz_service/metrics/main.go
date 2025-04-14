package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"pvz_service/internal/config"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatal("Ошибка загрузки конфига: ", err)
	}

	http.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%s", cfg.Prometheus.Port)
	log.Printf("Сервер метрик запущен на %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
