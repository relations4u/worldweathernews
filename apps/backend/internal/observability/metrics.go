package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// Metrics bündelt die zentralen Prometheus-Metriken des Backends.
type Metrics struct {
	Registry *prometheus.Registry

	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	DBPoolConnections   *prometheus.GaugeVec
	RedisCommandsTotal  *prometheus.CounterVec
}

// NewMetrics erzeugt eine eigene Registry mit Standard-Go-Collectors plus
// den Backend-spezifischen Metriken. Eigene Registry statt Default, damit
// Tests sauber bleiben.
func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()

	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	m := &Metrics{
		Registry: reg,
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "wwn_http_requests_total",
				Help: "Anzahl HTTP-Requests, gelabelt nach method, path, status.",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "wwn_http_request_duration_seconds",
				Help:    "HTTP-Request-Latenz, gelabelt nach method, path.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		DBPoolConnections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "wwn_db_pool_connections",
				Help: "Anzahl pgxpool-Verbindungen, gelabelt nach state.",
			},
			[]string{"state"},
		),
		RedisCommandsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "wwn_redis_commands_total",
				Help: "Anzahl Redis-Kommandos, gelabelt nach cmd, result.",
			},
			[]string{"cmd", "result"},
		),
	}

	reg.MustRegister(
		m.HTTPRequestsTotal,
		m.HTTPRequestDuration,
		m.DBPoolConnections,
		m.RedisCommandsTotal,
	)
	return m
}
