package clients

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusMiddleware struct {
	Gauge *prometheus.Gauge
}

func (p *PrometheusMiddleware) Run() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}

func NewPrometheusMiddleware() *PrometheusMiddleware {
	gauge := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "coralogics_operator_pods_total",
		Help: "The current number of running pods inside the k8s cluster",
	})

	return &PrometheusMiddleware{Gauge: &gauge}
}
