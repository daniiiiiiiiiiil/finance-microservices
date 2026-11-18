package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	UpdatesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telegram_updates_total",
			Help: "Total number of processed Telegram updates",
		},
		[]string{"kind", "status"},
	)

	UpdateDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "telegram_update_duration_seconds",
			Help:    "Telegram update processing duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60},
		},
		[]string{"kind"},
	)

	CommandsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telegram_commands_total",
			Help: "Total number of processed bot commands",
		},
		[]string{"command", "status"},
	)

	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telegram_errors_total",
			Help: "Total number of Telegram bot service errors",
		},
		[]string{"component"},
	)

	NotificationsSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telegram_notifications_sent_total",
			Help: "Total number of delivered Telegram notifications",
		},
		[]string{"kind"},
	)

	NotificationsErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telegram_notifications_errors_total",
			Help: "Total number of Telegram notification delivery errors",
		},
		[]string{"kind"},
	)

	GatewayRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_requests_total",
			Help: "Total number of Gateway requests made by the bot",
		},
		[]string{"method", "path", "status"},
	)

	GatewayDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_request_duration_seconds",
			Help:    "Gateway request duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	TelegramAPIRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telegram_api_requests_total",
			Help: "Total number of Telegram Bot API requests",
		},
		[]string{"method", "status"},
	)

	KafkaMessagesProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telegram_kafka_messages_processed_total",
			Help: "Total number of processed Kafka events",
		},
		[]string{"event_type", "status"},
	)

	GrpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telegram_grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"service", "method", "status"},
	)

	GrpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "telegram_grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"service", "method"},
	)

	GrpcRequestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "telegram_grpc_requests_in_flight",
			Help: "Current number of in-flight gRPC requests",
		},
		[]string{"service"},
	)
)

func RecordUpdate(kind string, success bool, duration time.Duration) {
	status := "success"
	if !success {
		status = "error"
	}
	UpdatesTotal.WithLabelValues(kind, status).Inc()
	UpdateDuration.WithLabelValues(kind).Observe(duration.Seconds())
}

func RecordCommand(command string, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	CommandsTotal.WithLabelValues(command, status).Inc()
}

func RecordError(component string) {
	ErrorsTotal.WithLabelValues(component).Inc()
}

func RecordNotification(kind string, success bool) {
	if success {
		NotificationsSent.WithLabelValues(kind).Inc()
		return
	}
	NotificationsErrors.WithLabelValues(kind).Inc()
}

func RecordGateway(method, path, status string, duration time.Duration) {
	GatewayRequests.WithLabelValues(method, path, status).Inc()
	GatewayDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

func RecordTelegramAPI(method string, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	TelegramAPIRequests.WithLabelValues(method, status).Inc()
}

func RecordKafkaMessage(eventType string, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	KafkaMessagesProcessed.WithLabelValues(eventType, status).Inc()
}

func RecordGrpcRequest(service, method, status string, duration time.Duration) {
	GrpcRequestsTotal.WithLabelValues(service, method, status).Inc()
	GrpcRequestDuration.WithLabelValues(service, method).Observe(duration.Seconds())
}

func StartGrpcRequest(service string) func() {
	GrpcRequestsInFlight.WithLabelValues(service).Inc()
	return func() {
		GrpcRequestsInFlight.WithLabelValues(service).Dec()
	}
}
