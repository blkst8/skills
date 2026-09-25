package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler returns the HTTP handler exposing Prometheus metrics. It is served
// by internal/http/handlers.Metrics on GET /metrics.
func Handler() http.Handler {
	return promhttp.Handler()
}
