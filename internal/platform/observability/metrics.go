package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "banking_http_requests_total",
	Help: "Total HTTP requests",
}, []string{"method", "path", "status"})

var httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "banking_http_request_duration_seconds",
	Help:    "HTTP request duration",
	Buckets: prometheus.DefBuckets,
}, []string{"method", "path"})

// MetricsMiddleware records Prometheus metrics.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		path := r.URL.Path
		httpRequests.WithLabelValues(r.Method, path, strconv.Itoa(rec.status)).Inc()
		httpDuration.WithLabelValues(r.Method, path).Observe(time.Since(start).Seconds())
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// MetricsHandler exposes /metrics.
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

var (
	outboxRelayedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "banking_outbox_events_relayed_total",
		Help: "Total outbox events relayed to the event bus",
	})
	settlementsProcessedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "banking_settlements_processed_total",
		Help: "Total settlement batches processed by the worker",
	})
)

// RecordOutboxRelayed increments the outbox relay counter.
func RecordOutboxRelayed(count int) {
	if count > 0 {
		outboxRelayedTotal.Add(float64(count))
	}
}

// RecordSettlementsProcessed increments the settlement counter.
func RecordSettlementsProcessed(count int) {
	if count > 0 {
		settlementsProcessedTotal.Add(float64(count))
	}
}
