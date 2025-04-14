package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Общее количество HTTP-запросов",
	}, []string{"method", "path", "status"})

	HTTPResponseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_response_time_seconds",
		Help:    "Время выполнения HTTP-запросов в секундах",
		Buckets: []float64{0.01, 0.05, 0.1, 0.3, 0.5, 1, 3, 5},
	}, []string{"method", "path"})

	PVZCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pvz_created_total",
		Help: "Счётчик количества созданных ПВЗ",
	})

	ReceptionsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "receptions_created_total",
		Help: "Счётчик количества созданных приёмок",
	})

	ProductsAdded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_added_total",
		Help: "Счётчик добавленных товаров",
	})
)

func Init(port string) {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		_ = http.ListenAndServe(port, nil)
	}()
}

func RecordHTTPRequest(method, path string, status int, duration time.Duration) {
	HTTPRequestsTotal.WithLabelValues(method, path, http.StatusText(status)).Inc()
	HTTPResponseTime.WithLabelValues(method, path).Observe(duration.Seconds())
}
