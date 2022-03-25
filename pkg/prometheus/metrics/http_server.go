package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"time"
)

// HttpServerMetrics represents a collection of metrics to be registered on a
// Prometheus metrics registry for a http server.
type HttpServerMetrics struct {
	startedRequest *prometheus.CounterVec
	handledRequest *prometheus.CounterVec

	countsHandlingTimeEnabled bool
	countsHandlingTimeOpts    prometheus.HistogramOpts
	countsHandlingTime        *prometheus.HistogramVec
}

func NewHttpServerMetrics(namespace string, subsystem string) *HttpServerMetrics {
	return &HttpServerMetrics{
		startedRequest: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "http_server_started_total",
				Help:      "Total number of http request started on the server.",
			}, []string{"method", "path"}),
		handledRequest: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "http_server_handled_total",
				Help:      "Total number of http request completed on the server, regardless of success or failure.",
			}, []string{"code", "method", "path"}),
		countsHandlingTimeEnabled: false,
		countsHandlingTimeOpts: prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "http_server_handling_seconds",
			Help:      "Histogram of response latency (seconds) of http that had been application-level handled by the server.",
			Buckets:   prometheus.DefBuckets,
		},
		countsHandlingTime: nil,
	}
}

// EnableCountsHandlingTime enables histograms being registered when
// registering the ServerMetrics on a Prometheus registry. Histograms can be
// expensive on Prometheus servers. It takes options to configure histogram
// options such as the defined buckets.
func (m *HttpServerMetrics) EnableCountsHandlingTime() {
	if !m.countsHandlingTimeEnabled {
		m.countsHandlingTime = prometheus.NewHistogramVec(
			m.countsHandlingTimeOpts,
			[]string{"code", "method", "path"},
		)
	}
	m.countsHandlingTimeEnabled = true
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent.
func (m *HttpServerMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.startedRequest.Describe(ch)
	m.handledRequest.Describe(ch)
	if m.countsHandlingTimeEnabled {
		m.countsHandlingTime.Describe(ch)
	}
}

// Collect is called by the Prometheus registry when collecting
// metrics. The implementation sends each collected metric via the
// provided channel and returns once the last metric has been sent.
func (m *HttpServerMetrics) Collect(ch chan<- prometheus.Metric) {
	m.startedRequest.Collect(ch)
	m.handledRequest.Collect(ch)
	if m.countsHandlingTimeEnabled {
		m.countsHandlingTime.Collect(ch)
	}
}

func (m *HttpServerMetrics) Interceptor() func(http.ResponseWriter, *http.Request, http.Handler) {
	return func(w http.ResponseWriter, r *http.Request, handler http.Handler) {
		monitor := newHttpServerMonitor(m, r)
		handler.ServeHTTP(&monitoredResponseWriter{w, monitor}, r)
		monitor.metrics.handledRequest.WithLabelValues(http.StatusText(monitor.code), monitor.method, monitor.path)
		if m.countsHandlingTimeEnabled {
			monitor.metrics.countsHandlingTime.WithLabelValues(http.StatusText(monitor.code), monitor.method, monitor.path).Observe(time.Since(monitor.startTime).Seconds())
		}
	}
}

type httpServerMonitor struct {
	metrics   *HttpServerMetrics
	code      int
	method    string
	path      string
	startTime time.Time
}

func newHttpServerMonitor(metrics *HttpServerMetrics, r *http.Request) *httpServerMonitor {
	monitor := &httpServerMonitor{
		metrics: metrics,
		method:  r.Method,
		path:    r.URL.Path,
		code:    http.StatusOK,
	}
	if metrics.countsHandlingTimeEnabled {
		monitor.startTime = time.Now()
	}
	monitor.metrics.startedRequest.WithLabelValues(monitor.method, monitor.path).Inc()

	return monitor
}

type monitoredResponseWriter struct {
	http.ResponseWriter
	monitor *httpServerMonitor
}

func (m *monitoredResponseWriter) WriteHeader(statusCode int) {
	m.monitor.code = statusCode
	m.ResponseWriter.WriteHeader(statusCode)
}
