package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	GrpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"service", "method", "status"},
	)

	GrpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"service", "method"},
	)

	GrpcRequestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "grpc_requests_in_flight",
			Help: "Current number of in-flight gRPC requests",
		},
		[]string{"service"},
	)

	KafkaConsumerLag = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "kafka_consumer_lag_seconds",
			Help: "Kafka consumer lag in seconds",
		},
	)

	KafkaMessagesProcessed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_messages_processed_total",
			Help: "Total number of Kafka messages processed",
		},
	)

	KafkaMessagesErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_messages_errors_total",
			Help: "Total number of Kafka message processing errors",
		},
	)

	SagaTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "saga_total",
			Help: "Total number of saga operations",
		},
		[]string{"operation", "status"},
	)

	SagaDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "saga_duration_seconds",
			Help:    "Saga operation duration in seconds",
			Buckets: []float64{.1, .5, 1, 2, 5, 10, 30},
		},
		[]string{"operation"},
	)

	DbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"service", "operation", "table"},
	)

	DbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "operation", "table"},
	)

	DbQueryErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_query_errors_total",
			Help: "Total number of database query errors",
		},
		[]string{"service", "operation", "table"},
	)
)

// записывает метрику gRPC запроса
func RecordGrpcRequest(service, method, status string, duration float64) {
	GrpcRequestsTotal.WithLabelValues(service, method, status).Inc()
	GrpcRequestDuration.WithLabelValues(service, method).Observe(duration)
}

// увеличивает счётчик in-flight запросов
func RecordGrpcRequestStart(service string) {
	GrpcRequestsInFlight.WithLabelValues(service).Inc()
}

// уменьшает счётчик in-flight запросов
func RecordGrpcRequestFinish(service string) {
	GrpcRequestsInFlight.WithLabelValues(service).Dec()
}

// записывает метрику саги
func RecordSaga(operation string, success bool, duration float64) {
	status := "success"
	if !success {
		status = "error"
	}
	SagaTotal.WithLabelValues(operation, status).Inc()
	SagaDuration.WithLabelValues(operation).Observe(duration)
}

// записывает метрику обработки Kafka сообщения
func RecordKafkaMessage(success bool) {
	KafkaMessagesProcessed.Inc()
	if !success {
		KafkaMessagesErrors.Inc()
	}
}

// записывает метрику запроса к БД
func RecordDbQuery(service, operation, table string, duration float64, err error) {
	DbQueriesTotal.WithLabelValues(service, operation, table).Inc()
	DbQueryDuration.WithLabelValues(service, operation, table).Observe(duration)
	if err != nil {
		DbQueryErrors.WithLabelValues(service, operation, table).Inc()
	}
}
