package metrics

import "github.com/prometheus/client_golang/prometheus/promhttp"

// Handler exposes Prometheus metrics.
var Handler = promhttp.Handler
