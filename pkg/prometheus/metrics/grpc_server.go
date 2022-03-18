package metrics

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

// GrpcServerMetrics represents a collection of metrics to be registered on a
// Prometheus metrics registry for a gRPC server.
type GrpcServerMetrics struct {
	serverStartedCounter    *prometheus.CounterVec
	serverHandledCounter    *prometheus.CounterVec
	serverStreamMsgReceived *prometheus.CounterVec
	serverStreamMsgSent     *prometheus.CounterVec

	serverHandledHistogramEnabled bool
	serverHandledHistogramOpts    prometheus.HistogramOpts
	serverHandledHistogram        *prometheus.HistogramVec
}

// NewGrpcServerMetrics returns a ServerMetrics object. Use a new instance of
// ServerMetrics when not using the default Prometheus metrics registry, for
// example when wanting to control which metrics are added to a registry as
// opposed to automatically adding metrics via init functions.
func NewGrpcServerMetrics(namespace string, subsystem string) *GrpcServerMetrics {
	return &GrpcServerMetrics{
		serverStartedCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "grpc_server_started_total",
				Help:      "Total number of RPCs started on the server.",
			}, []string{"grpc_type", "grpc_service", "grpc_method"}),
		serverHandledCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "grpc_server_handled_total",
				Help:      "Total number of RPCs completed on the server, regardless of success or failure.",
			}, []string{"grpc_type", "grpc_service", "grpc_method", "grpc_code"}),
		serverStreamMsgReceived: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "grpc_server_msg_received_total",
				Help:      "Total number of RPC stream messages received on the server.",
			}, []string{"grpc_type", "grpc_service", "grpc_method"}),
		serverStreamMsgSent: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "grpc_server_msg_sent_total",
				Help:      "Total number of gRPC stream messages sent by the server.",
			}, []string{"grpc_type", "grpc_service", "grpc_method"}),
		serverHandledHistogramEnabled: false,
		serverHandledHistogramOpts: prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "grpc_server_handling_seconds",
			Help:      "Histogram of response latency (seconds) of gRPC that had been application-level handled by the server.",
			Buckets:   prometheus.DefBuckets,
		},
		serverHandledHistogram: nil,
	}
}

// EnableHandlingTimeHistogram enables histograms being registered when
// registering the ServerMetrics on a Prometheus registry. Histograms can be
// expensive on Prometheus servers. It takes options to configure histogram
// options such as the defined buckets.
func (m *GrpcServerMetrics) EnableHandlingTimeHistogram() {
	if !m.serverHandledHistogramEnabled {
		m.serverHandledHistogram = prometheus.NewHistogramVec(
			m.serverHandledHistogramOpts,
			[]string{"grpc_type", "grpc_service", "grpc_method"},
		)
	}
	m.serverHandledHistogramEnabled = true
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent.
func (m *GrpcServerMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.serverStartedCounter.Describe(ch)
	m.serverHandledCounter.Describe(ch)
	m.serverStreamMsgReceived.Describe(ch)
	m.serverStreamMsgSent.Describe(ch)
	if m.serverHandledHistogramEnabled {
		m.serverHandledHistogram.Describe(ch)
	}
}

// Collect is called by the Prometheus registry when collecting
// metrics. The implementation sends each collected metric via the
// provided channel and returns once the last metric has been sent.
func (m *GrpcServerMetrics) Collect(ch chan<- prometheus.Metric) {
	m.serverStartedCounter.Collect(ch)
	m.serverHandledCounter.Collect(ch)
	m.serverStreamMsgReceived.Collect(ch)
	m.serverStreamMsgSent.Collect(ch)
	if m.serverHandledHistogramEnabled {
		m.serverHandledHistogram.Collect(ch)
	}
}

// UnaryServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Unary RPCs.
func (m *GrpcServerMetrics) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		monitor := newGrpcServerMonitor(m, Unary, info.FullMethod)
		monitor.ReceivedMessage()
		resp, err := handler(ctx, req)
		s, _ := status.FromError(err)
		monitor.Handled(s.Code())
		if err == nil {
			monitor.SentMessage()
		}
		return resp, err
	}
}

// StreamServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Streaming RPCs.
func (m *GrpcServerMetrics) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		monitor := newGrpcServerMonitor(m, grpcStreamType(info), info.FullMethod)
		err := handler(srv, &monitoredServerStream{ss, monitor})
		s, _ := status.FromError(err)
		monitor.Handled(s.Code())
		return err
	}
}

type grpcServerMonitor struct {
	metrics     *GrpcServerMetrics
	grpcType    string
	serviceName string
	methodName  string
	startTime   time.Time
}

func newGrpcServerMonitor(metrics *GrpcServerMetrics, grpcType string, fullMethod string) *grpcServerMonitor {
	monitor := &grpcServerMonitor{
		metrics:  metrics,
		grpcType: grpcType,
	}
	if monitor.metrics.serverHandledHistogramEnabled {
		monitor.startTime = time.Now()
	}
	monitor.serviceName, monitor.methodName = splitGrpcMethodName(fullMethod)
	monitor.metrics.serverStartedCounter.WithLabelValues(monitor.grpcType, monitor.serviceName, monitor.methodName).Inc()
	return monitor
}

func (m *grpcServerMonitor) ReceivedMessage() {
	m.metrics.serverStreamMsgReceived.WithLabelValues(m.grpcType, m.serviceName, m.methodName).Inc()
}

func (m *grpcServerMonitor) SentMessage() {
	m.metrics.serverStreamMsgSent.WithLabelValues(m.grpcType, m.serviceName, m.methodName).Inc()
}

func (m *grpcServerMonitor) Handled(code codes.Code) {
	m.metrics.serverHandledCounter.WithLabelValues(m.grpcType, m.serviceName, m.methodName, code.String()).Inc()
	if m.metrics.serverHandledHistogramEnabled {
		m.metrics.serverHandledHistogram.WithLabelValues(m.grpcType, m.serviceName, m.methodName).Observe(time.Since(m.startTime).Seconds())
	}
}

// monitoredServerStream wraps grpc.ServerStream allowing each Sent/Recv of message to increment counters.
type monitoredServerStream struct {
	grpc.ServerStream
	monitor *grpcServerMonitor
}

func (ss *monitoredServerStream) SendMsg(m interface{}) error {
	err := ss.ServerStream.SendMsg(m)
	if err == nil {
		ss.monitor.SentMessage()
	}
	return err
}

func (ss *monitoredServerStream) RecvMsg(m interface{}) error {
	err := ss.ServerStream.RecvMsg(m)
	if err == nil {
		ss.monitor.ReceivedMessage()
	}
	return err
}
