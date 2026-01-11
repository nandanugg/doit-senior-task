package config

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewPrometheusHandler returns a new HTTP handler for Prometheus metrics.
func NewPrometheusHandler() http.Handler {
	return promhttp.Handler()
}
