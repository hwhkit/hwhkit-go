package observability

import (
	"github.com/hwhkit/hwhkit-go/buildinfo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

func InitMetrics() *prometheus.Registry {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
	)
	info := buildinfo.Get()
	reg.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "hwhkit_build_info",
		Help: "Build metadata of the running hwhkit-go binary.",
		ConstLabels: prometheus.Labels{
			"version":    info.Version,
			"git_sha":    info.GitSHA,
			"build_time": info.BuildTime,
			"go_version": info.GoVersion,
		},
	}, func() float64 { return 1 }))
	return reg
}

var (
	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests, partitioned by method, route, and response status.",
	}, []string{"method", "route", "status"})

	httpRequestDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})
)

func RegisterHTTPMetrics(reg *prometheus.Registry) {
	reg.MustRegister(httpRequestsTotal, httpRequestDurationSeconds)
}

func HTTPRequestsCounter() *prometheus.CounterVec       { return httpRequestsTotal }
func HTTPRequestDuration() *prometheus.HistogramVec     { return httpRequestDurationSeconds }
